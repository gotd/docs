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
	},
})
```

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
long-running service, give it a `Storage` (and access-hash storage) so it resumes from
where it left off. The
[`gotd/contrib`](https://github.com/gotd/contrib) module provides bbolt-backed
implementations:

```go
gaps := updates.New(updates.Config{
	Handler: d,
	Storage: boltstor.NewStateStorage(boltdb),
})
```

The [`userbot`](https://github.com/gotd/td/tree/main/examples/userbot) example uses
persistent state; the [`updates`](https://github.com/gotd/td/tree/main/examples/updates)
example is a minimal in-memory version.

:::note[Known limits]
The engine relies on the server for `getDifference` correctness and cannot recover from
`ChannelDifferenceTooLong` (it resyncs that channel instead). Stateless updates can't be
ordered. These are inherent to the MTProto update model.
:::
