---
sidebar_position: 7
---

# Update recovery

The plain [dispatcher](../basics/handling-updates.md) delivers updates as they arrive,
but Telegram's update stream is *stateful*: each update carries a sequence number, and a
client that was briefly offline can miss updates or receive them out of order. Recovering
those gaps means tracking state and calling `updates.getDifference` /
`updates.getChannelDifference` at the right moments.

The [`updates`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/updates.html)
package — the "gap engine" — does this for you.

## Wiring it up

The manager sits between the client and your dispatcher. It needs to both **receive**
updates (as the `UpdateHandler`) and **observe** RPC responses (as a middleware), so it
is installed in two places:

```go
import (
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
)

d := tg.NewUpdateDispatcher()
gaps := updates.New(updates.Config{
	Handler: d, // your dispatcher receives recovered, ordered updates
	Logger:  logzap.New(log.Named("gaps")),
})

client, err := telegram.ClientFromEnvironment(telegram.Options{
	UpdateHandler: gaps,
	Middlewares: []telegram.Middleware{
		updhook.UpdateHook(gaps.Handle),
		// Keep local pts in sync after self-initiated reads/deletes.
		updhook.AffectedHook(gaps),
	},
})
```

## Self-initiated reads and deletes

`UpdateHook` forwards `tg.UpdatesClass` results to the manager, but some methods
return a different shape. `messages.readHistory`, `messages.deleteMessages`,
`channels.deleteMessages`, `messages.readMentions` and `messages.deleteHistory`
return `messages.affectedMessages` / `messages.affectedHistory` — results that
carry a **pts increment the client must apply**, yet are not updates and so never
reach the gap engine.

If you skip them, the server's pts advances while your local pts stays behind, so
the next genuine update looks like a gap and is buffered. A common symptom: you
read or delete a message in a handler, and a later edit in the same chat is then
postponed until some unrelated pts-changing update arrives
([gotd/td#1382](https://github.com/gotd/td/issues/1382)).

[`updhook.AffectedHook`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/updates/hook.html)
closes this gap. It inspects each RPC result, and when it sees an affected-pts
result it calls `Manager.HandleAffected` to advance the local pts (releasing any
updates postponed behind it), routing channel methods to the right channel
sequence and everything else to the common one. It is opt-in, mirrors TDLib's
behavior, and is a no-op before `gaps.Run` — so install it alongside `UpdateHook`
as shown above. See [gotd/td#1782](https://github.com/gotd/td/pull/1782).

## Your own sends

Telegram pushes the result of `messages.sendMessage`, `messages.sendMedia`,
`messages.editMessage` and similar calls to the account's **other** sessions, not to the
one that made the call: the session only gets it as the call's result. Without
`updhook.UpdateHook` (see [Wiring it up](#wiring-it-up)) your handler sees your own
message only after a later update reveals the pts gap and `getDifference` fetches it,
which can take minutes.

With `UpdateHook` the pts stays in sync, but one result still loses data. A text sent
with `messages.sendMessage` comes back as `updateShortSentMessage`, which has the new
message ID, date and pts, but no text and no peer. The engine can't rebuild it, so the
handler gets a `tg.MessageEmpty` with that ID. If you record your own messages, build the
message from the request in a middleware that sees both, as Telegram Web A does for its
local messages:

```go
func sentHook(gaps *updates.Manager, self int64) telegram.Middleware {
	return telegram.MiddlewareFunc(func(next tg.Invoker) telegram.InvokeFunc {
		return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			if err := next.Invoke(ctx, input, output); err != nil {
				return err
			}
			box, ok := output.(*tg.UpdatesBox)
			if !ok {
				return nil
			}
			u := box.Updates
			if s, ok := u.(*tg.UpdateShortSentMessage); ok {
				if req, ok := input.(*tg.MessagesSendMessageRequest); ok {
					u = &tg.UpdateShort{Date: s.Date, Update: &tg.UpdateNewMessage{
						Message: &tg.Message{
							ID: s.ID, Date: s.Date, Out: true,
							FromID:   &tg.PeerUser{UserID: self},
							PeerID:   peerOf(req.Peer), // your InputPeer → Peer mapping
							Message:  req.Message,
							Entities: s.Entities, // the server's entities, e.g. links
						},
						Pts: s.Pts, PtsCount: s.PtsCount,
					}}
				}
			}
			// The call succeeded: don't fail it because handling did; recovery fills
			// the gap later.
			_ = gaps.Handle(ctx, u)
			return nil
		}
	})
}
```

Use it instead of `updhook.UpdateHook(gaps.Handle)`, not next to it, so results aren't
handled twice. Media sends return full `updates`, so only plain text needs this.

## Running the engine

After authenticating, start the engine with your user ID. It blocks, keeping the client
alive and processing updates until the context is cancelled:

```go
return client.Run(ctx, func(ctx context.Context) error {
	// ... authenticate ...
	self, err := client.Self(ctx)
	if err != nil {
		return err
	}
	return gaps.Run(ctx, client.API(), self.ID, updates.AuthOptions{
		IsBot: self.Bot,
		OnStart: func(ctx context.Context) {
			log.Info("recovery started")
		},
	})
})
```

Your dispatcher handlers (`OnNewMessage`, `OnNewChannelMessage`, …) are registered
exactly as before — the difference is they now receive a gap-free, correctly ordered
stream.

## Persisting state

By default the engine keeps state in memory, so a restart re-syncs from scratch. For a
long-running service, give it a `Storage` and access-hash storage so it resumes from
where it left off: without stored channel access hashes it can't fetch channel
differences after a restart. The
[`gotd/contrib`](https://github.com/gotd/contrib) module provides bbolt- and
pebble-backed implementations:

```go
import "github.com/gotd/contrib/pebble"

gaps := updates.New(updates.Config{
	Handler:          d,
	Storage:          pebble.NewStateStorage(db),
	AccessHasher:     pebble.NewAccessHasher(db),
	UserAccessHasher: pebble.NewAccessHasher(db),
})
```

The bbolt equivalents are `bbolt.NewStateStorage(db)` and `bbolt.NewAccessHasher(db)`.

## Tracking channels

The engine tracks channels with a stored state, and learns about other channels from
their next update. To track channels joined while the client was offline (or reported
via `OnLoadChannelStateFailed`), pass them with the pts of their dialog. Stored pts take
precedence, so missed updates are fetched:

```go
gaps.Run(ctx, api, self.ID, updates.AuthOptions{
	OnStart: func(ctx context.Context) {
		for _, d := range dialogs { // e.g. from messages.getDialogs
			_ = gaps.TrackChannel(ctx, d.ChannelID, d.AccessHash, d.Pts)
		}
	},
})
```

### "Failed to load channel state"

On start the engine loads each channel with a stored pts, and needs its access hash
from `ChannelAccessHasher` to call `updates.getChannelDifference`. If the pts is stored
but the hash isn't, it logs "Failed to load channel state, skipping channel" (or calls
`OnLoadChannelStateFailed`) and skips the channel until its next update arrives, so
updates missed while offline are not fetched.

Usually the hash store is behind the state store: another database, a cache that was
cleared, or a channel the account left or that was deleted. Fix it by giving the hasher
the hash (any store you have works, e.g. your own archive of chats), or by calling
`TrackChannel` with the dialog's pts and hash as above. The stored state of a channel
that is gone stays, so the same warning repeats on every start.

## Updates before `Run`

Updates can arrive before the manager's `Run` has loaded the state, e.g. while the
client connects. By default they are handled right away (`BeforeRunPass`). To handle
them in order once the state is loaded, buffer them, or drop them and let gap
recovery fetch them again:

```go
updates.Config{
	BeforeRun:           updates.BeforeRunBuffer, // or updates.BeforeRunDrop
	BeforeRunBufferSize: 100,                     // batches; extra ones are dropped
}
```

## Channel gaps

When a channel gap is too long to recover (`updates.channelDifferenceTooLong`), the
engine moves to the new state and reports the gap with the latest messages returned by
the server. They are not passed to your handler, because they may include messages it
has already received. Backfill what you missed with `messages.getHistory`:

```go
updates.Config{
	OnChannelGap: func(gap updates.ChannelGap) {
		// gap.ChannelID, gap.AccessHash, gap.LocalPts, gap.RemotePts, gap.TopMessage
		// gap.Messages are the latest messages, with gap.Users and gap.Chats.
	},
}
```

## Channels that aren't pushed

Telegram doesn't push updates of every channel to every session. For such channels the
engine only learns of new posts from its idle `getChannelDifference`, which runs
15 minutes after the channel's last update and then every 15 minutes while none arrive
(difference trigger `channel-idle-timeout` in the [metrics](../advanced/observability)). Their messages therefore arrive 0–15
minutes late, about 7 minutes on average.

The official clients do the same for channels that aren't open, and poll often only the
open ones, at the `timeout` returned by the last `channelDifference`
(Telegram Desktop `Updates::addActiveChat`, TDLib `channel_get_difference_timeout_` for
open chats). With gotd, poll a channel you need sooner yourself, e.g. with
`messages.getHistory` from its last known message.

To see whether you're affected, compare update lag per channel with
`gotd.updates.difference.triggers{reason="channel-idle-timeout"}`: a median lag of
several minutes for a few channels matches the idle poll.

The [`userbot`](https://github.com/gotd/td/tree/main/examples/userbot) example uses
persistent state; the [`updates`](https://github.com/gotd/td/tree/main/examples/updates)
example is a minimal in-memory version.

:::note[Known limits]
The engine relies on the server for `getDifference` correctness, and messages missed in
a `ChannelDifferenceTooLong` gap must be fetched by the application
(see [Channel gaps](#channel-gaps)). Stateless updates can't be ordered. These are
inherent to the MTProto update model.
:::
