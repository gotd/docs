---
sidebar_position: 1
---

# The message sender

[`message.NewSender`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/message.html)
is the workhorse for outgoing messages. [Sending messages](../basics/sending-messages.md)
covered the essentials; this page is the fuller tour.

```go
import "github.com/gotd/td/telegram/message"

sender := message.NewSender(client.API())
```

## Targets

Every send starts by choosing a destination, which returns a request builder:

```go
sender.Self()                  // Saved Messages
sender.Resolve("@durov")       // username / t.me link / phone / deeplink
sender.ResolveDomain("durov")  // bare username
sender.ResolvePhone("+123...") // phone number
sender.To(inputPeer)           // a tg.InputPeerClass you already hold
sender.Reply(entities, update) // reply inside an update handler
```

`Resolve` and friends are lazy *promises* — the network lookup happens when you call a
terminal method like `Text`. To cache and speed up repeated resolution, give the sender
a resolver backed by [peer storage](./peers.md) via `sender.WithResolver`.

## Message kinds

The builder has a terminal method for each kind of message:

```go
b := sender.Resolve("@channel")

b.Text(ctx, "plain")
b.Textf(ctx, "value = %d", 42)
b.StyledText(ctx, styling.Bold("formatted"))   // see Styled text
b.Media(ctx, mediaOption)                       // uploaded/external media
b.Album(ctx, m1, m2, m3)                         // media group
b.Photo(ctx, inputFile)
b.PhotoExternal(ctx, "https://example.com/a.jpg")
b.Poll(ctx, /* poll options */)
b.Contact(ctx, "+123", "Alice")
b.Location(ctx, lat, long)
b.Dice(ctx)                                      // 🎲 and other emoji games
```

`Media` accepts the media options produced by the
[uploader helpers](./uploading-files.md) (e.g. `message.UploadedDocument`,
`message.UploadedPhoto`).

## Modifiers

Modifiers chain before the terminal call and return the builder:

| Modifier              | Effect                                       |
|-----------------------|----------------------------------------------|
| `Silent()`            | Deliver without a notification sound         |
| `NoWebpage()`         | Suppress link previews                       |
| `Reply(msgID)`        | Reply to a specific message                  |
| `Schedule(t)`         | Schedule for a `time.Time`                   |
| `NoForwards()`        | Forbid forwarding/saving the message         |
| `InvertMedia()`       | Put media below the text                     |
| `Markup(m)` / `Row(...)` | Attach an inline/reply keyboard           |
| `SendAs(peer)`        | Post as a linked channel/anonymous admin     |

```go
sender.Resolve("@channel").
	Silent().
	NoWebpage().
	Reply(replyToID).
	Text(ctx, "scheduled-looking quiet reply")
```

## Editing and drafts

```go
sender.To(peer).Edit(msgID).Text(ctx, "edited text")
sender.To(peer).SaveDraft(ctx, "a draft I'll finish later")
```

## Uploads and randomness

`sender.WithUploader(u)` lets the sender upload local files for media sends, and
`sender.WithRand(r)` overrides the random source used for message IDs (useful in tests).

See next: [Styled text and HTML](./styling.md).
