---
sidebar_position: 1
---

# Calling the API

gotd generates a Go method for **every** method in the Telegram schema. You reach them
through `client.API()`, which returns a
[`*tg.Client`](https://ref.gotd.dev/pkg/github.com/gotd/td/tg.html).

```go
return client.Run(ctx, func(ctx context.Context) error {
	api := client.API()

	state, err := api.UpdatesGetState(ctx)
	if err != nil {
		return err
	}
	// state: &tg.UpdatesState{Pts:197 Qts:0 Date:1606855030 Seq:1 UnreadCount:106}
	fmt.Printf("%+v\n", state)
	return nil
})
```

`api` is also obtainable as `tg.NewClient(client)` — the two are equivalent, and the
latter is handy when you want the raw client without the `telegram.Client` wrapper (for
example inside a [takeout session](../advanced/data-export.md)).

## Naming convention

Schema method names map to Go method names by dropping dots and PascalCasing:

| Telegram method            | gotd method                  |
|----------------------------|------------------------------|
| `help.getNearestDC`        | `api.HelpGetNearestDC`       |
| `auth.signIn`              | `api.AuthSignIn`             |
| `messages.sendMessage`     | `api.MessagesSendMessage`    |
| `channels.getParticipants` | `api.ChannelsGetParticipants`|

Methods with several parameters take a generated request struct:

```go
res, err := api.MessagesGetMessages(ctx, []tg.InputMessageClass{
	&tg.InputMessageID{ID: 42},
})
```

Each generated type carries the official Telegram documentation as Go doc comments,
including links back to [core.telegram.org](https://core.telegram.org/schema).

## Sum types: the `...Class` interfaces

Where the schema has a type with multiple constructors, gotd generates an interface
named `SomethingClass` and one struct per constructor. Use a type switch to handle them:

```go
res, err := api.MessagesGetMessages(ctx, ids)
if err != nil {
	return err
}
switch m := res.(type) {
case *tg.MessagesMessages:
	useMessages(m.Messages)
case *tg.MessagesMessagesSlice:
	useMessages(m.Messages)
case *tg.MessagesChannelMessages:
	useMessages(m.Messages)
}
```

Many generated types also offer helpers like `AsNotEmpty()` and typed getters
(`GetUsername()`) that return `(value, ok)` for optional fields.

## Errors and FLOOD_WAIT

RPC errors come back as `*tgerr.Error`. Match them by type with the `tgerr` package:

```go
import "github.com/gotd/td/tgerr"

if _, err := api.MessagesSendMessage(ctx, req); err != nil {
	if tgerr.Is(err, "FLOOD_WAIT") {
		// rate limited; back off
	}
	return err
}
```

Rather than handle `FLOOD_WAIT` everywhere, install the
[floodwait middleware](../helpers/middleware.md), which retries automatically.

## MTProto JSON

You can also invoke methods using Telegram's MTProto JSON format via
`api.InvokeJSON`, passing JSON with an `@type` field and receiving JSON back. This is
useful for bridging dynamic callers; for Go code the typed methods above are preferred.

## When to drop to the raw API

The [helpers](../helpers/message-sender.md) cover the common cases (sending messages,
uploads, downloads, pagination) far more ergonomically. Reach for the raw API when you
need a method the helpers don't wrap — which is most of the schema.
