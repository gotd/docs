---
sidebar_position: 9
---

# Persistence & pooling

## Persistence

By default everything is kept in memory and **nothing survives a restart**.
Provide a `Storage` to persist the MTProto session, peer access hashes and peer
cache, and the update gap state. A single
[`storage.BBoltStorage`](https://github.com/gotd/botapi/blob/main/storage)
satisfies all of these at once, backed by one bbolt file:

```go
import (
	"go.etcd.io/bbolt"

	"github.com/gotd/botapi"
	"github.com/gotd/botapi/storage"
)

db, err := bbolt.Open("bot.bbolt", 0o666, nil)
if err != nil {
	return err
}
opts := botapi.Options{
	AppID: appID, AppHash: appHash,
	Storage: storage.NewBBoltStorage(db),
}
```

:::info[Why persist peers?]
The most important thing storage keeps is **peer access hashes**. The Bot API
speaks in bare `int64` chat IDs, but MTProto needs an access hash for each peer.
botapi harvests those from every update it sees; persisting them means the bot
can keep addressing chats it has interacted with after a restart, instead of
re-discovering them. See [peer resolution](../helpers/peers.md) for the
underlying mechanism.
:::

## Running many bots

`pool.Pool` lazily starts and multiplexes bots by token over one process — the
multi-bot front end for a service serving many bots:

```go
import "github.com/gotd/botapi/pool"

p, err := pool.New(pool.Options{
	AppID: appID, AppHash: appHash,
	StateDir:    "state",     // per-token <id>.bbolt files; in-memory if empty
	IdleTimeout: time.Hour,   // GC bots idle this long
})
if err != nil {
	return err
}
go p.RunGC(ctx)

err = p.Do(ctx, token, func(b *botapi.Bot) error {
	_, err := b.SendMessage(ctx, botapi.ID(chatID), "hi")
	return err
})
```

`Do` starts and authorizes the bot on first use — concurrent callers share one
startup and a failure is returned to all of them — and gives each token its own
storage under `StateDir`. `RunGC` reaps bots that have been idle longer than
`IdleTimeout`; `Kill` and `Close` shut bots down explicitly.

## The escape hatch

Anything the Bot API surface does not cover is one call away:

```go
api := bot.Raw()         // *tg.Client — direct MTProto
disp := bot.Dispatcher() // the raw update dispatcher
```

`Raw()` returns the underlying `gotd/td` [`*tg.Client`](../basics/calling-the-api.md),
so you can invoke any MTProto method the typed surface hasn't reached yet, and
`Dispatcher()` exposes the raw update dispatcher. This mirrors `gotd/td`'s own
philosophy: a high-level API that never traps you below it.
