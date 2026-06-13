---
sidebar_position: 6
---

# Peers and resolution

A *peer* is a user, chat or channel. The raw API identifies them by numeric IDs plus an
`access_hash`, and many methods want an `InputPeer`/`InputChannel`. The
[`peers`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/peers.html) package turns
human-friendly references into rich peer objects and caches the access hashes.

## The manager

```go
import "github.com/gotd/td/telegram/peers"

manager := peers.Options{}.Build(client.API())
```

`Resolve` accepts usernames, `t.me` links, deeplinks, phone numbers and bare domains:

```go
p, err := manager.Resolve(ctx, "@telegram")
if err != nil {
	return err
}
fmt.Println(p.VisibleName(), p.ID())
```

The returned `peers.Peer` exposes common attributes uniformly: `ID()`, `VisibleName()`,
`Username()`, `Verified()`, `Scam()`, `InputPeer()` and more.

## Concrete peer types

Type-assert to reach type-specific behaviour. To list a channel's members you need a
`peers.Channel` and its `InputChannel()`:

```go
p, err := manager.Resolve(ctx, "@some_channel")
if err != nil {
	return err
}
channel, ok := p.(peers.Channel)
if !ok {
	return errors.New("not a channel")
}

// channel.InputChannel() feeds the participants iterator.
```

See [Pagination and iterators](./query-iterators.md) for the participants example, and
`manager.Self(ctx)` for the current account.

## Why caching matters

Resolving the same username repeatedly costs an RPC each time and contributes to rate
limits. Backing resolution with persistent storage means you resolve once and reuse the
cached access hash forever. The
[`gotd/contrib/storage`](https://github.com/gotd/contrib) package provides a
`PeerStorage` and a `ResolverCache`, plus an `UpdateHook` that learns peers as updates
arrive:

```go
import (
	"github.com/gotd/contrib/storage"
	"github.com/gotd/td/telegram/message/peer"
)

// peerDB implements storage.PeerStorage (e.g. backed by Pebble or bbolt).
resolver := storage.NewResolverCache(peer.Plain(client.API()), peerDB)
updateHandler := storage.UpdateHook(dispatcher, peerDB)

// Look up a cached peer by its ID later:
p, err := storage.FindPeer(ctx, peerDB, msg.GetPeerID())
```

The [`userbot`](https://github.com/gotd/td/tree/main/examples/userbot) example wires all
of this together for a production-grade setup.

## Offline parsing with deeplinks

If you only need to *parse* a link without contacting Telegram, use the
[`deeplink`](./deeplinks.md) package directly.
