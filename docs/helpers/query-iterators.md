---
sidebar_position: 5
---

# Pagination and iterators

Many Telegram methods return results in pages, expecting you to track offsets and
re-request until exhausted. The
[`query`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/query.html) package wraps
the common ones in iterators that handle paging for you.

Each iterator offers two styles:

* `ForEach(ctx, func(ctx, elem) error)` — a callback per element (most convenient).
* `Iter()` / `Next(ctx)` / `Value()` / `Err()` — a manual loop when you need full control.

## Dialogs

List every dialog (chat) on the account:

```go
import (
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/dialogs"
)

err := query.GetDialogs(client.API()).ForEach(ctx,
	func(ctx context.Context, elem dialogs.Elem) error {
		if elem.Deleted() {
			return nil
		}
		fmt.Println(elem.Peer)
		return nil
	},
)
```

## Message history

Walk a chat's history newest-first:

```go
import "github.com/gotd/td/telegram/query/messages"

err := query.Messages(client.API()).
	GetHistory(&tg.InputPeerSelf{}). // Saved Messages
	ForEach(ctx, func(ctx context.Context, elem messages.Elem) error {
		if msg, ok := elem.Msg.(*tg.Message); ok {
			fmt.Printf("[%d] %s\n", msg.ID, msg.Message)
		}
		return nil
	})
```

`Elem` also has `elem.File()` to extract downloadable media — see
[Downloading files](./downloading-files.md).

### Stopping early

Return a sentinel error from the callback and ignore it after the loop:

```go
var errDone = errors.New("done")

err := query.Messages(client.API()).GetHistory(peer).
	ForEach(ctx, func(ctx context.Context, elem messages.Elem) error {
		seen++
		if seen >= limit {
			return errDone
		}
		return nil
	})
if err != nil && !errors.Is(err, errDone) {
	return err
}
```

## Channel participants

List the members of a channel or supergroup. Resolve the channel first (see
[Peers](./peers.md)), then iterate — this is the
[`get-participants`](https://github.com/gotd/td/tree/main/examples/get-participants)
example:

```go
import "github.com/gotd/td/telegram/query/channels/participants"

q := participants.NewQueryBuilder(client.API()).
	GetParticipants(channel.InputChannel()).
	Recent()

// A cheap total without iterating.
if total, err := q.Count(ctx); err == nil {
	fmt.Println("members:", total)
}

err := q.ForEach(ctx, func(ctx context.Context, p participants.Elem) error {
	user, ok := p.User()
	if !ok {
		return nil
	}
	fmt.Printf("%d %s %s\n", user.ID, user.FirstName, user.LastName)
	return nil
})
```

## Other iterators

`query.NewQuery(api)` is the umbrella builder; alongside dialogs, messages and
participants it covers blocked users, profile photos and featured sticker sets. See the
[reference](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/query.html) for the
full set.

For low-level pagination that the package doesn't wrap (for example
`messages.getSavedGifs`), the
[`query/hasher`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/query/hasher.html)
package computes Telegram's pagination hashes for you, as in the
[`gif-download`](https://github.com/gotd/td/tree/main/examples/gif-download) example.
