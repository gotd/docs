---
sidebar_position: 2
---

# Sending messages

Sending a message with the raw API means building random IDs, resolving the peer into an
`InputPeer`, and assembling a request. The
[`message`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/message.html) package
does all of that behind a fluent builder.

```go
import "github.com/gotd/td/telegram/message"

sender := message.NewSender(client.API())

// Send to a username, resolving it automatically.
if _, err := sender.Resolve("@durov").Text(ctx, "Hello from gotd!"); err != nil {
	return err
}
```

## Choosing a target

A `Sender` produces a request builder per destination:

| Method                       | Target                                       |
|------------------------------|----------------------------------------------|
| `sender.Self()`              | Saved Messages (yourself)                    |
| `sender.Resolve("@name")`    | Username, phone, t.me link or deeplink       |
| `sender.To(inputPeer)`       | A peer you already have as `tg.InputPeerClass`|
| `sender.Reply(entities, upd)`| Reply within an update handler               |

`Resolve` accepts usernames (`@durov`), `t.me/durov` links, `tg:` deeplinks and phone
numbers — see [Peers and resolution](../helpers/peers.md).

## Beyond plain text

The builder returned for a target exposes the full range of message types:

```go
b := sender.Resolve("@channel")

b.Text(ctx, "plain text")
b.Textf(ctx, "formatted %d", 42)
b.StyledText(ctx, styling.Bold("bold"), styling.Plain(" and normal"))
b.Photo(ctx, inputFile)
b.Media(ctx, document)        // any uploaded media, see Uploading files
b.Dice(ctx)                   // 🎲
b.Poll(ctx, /* options */)
```

For formatted text see [Styled text and HTML](../helpers/styling.md); for sending files
see [Uploading files](../helpers/uploading-files.md).

## Message options

Options chain before the terminal call:

```go
sender.Resolve("@channel").
	Silent().          // no notification
	NoWebpage().       // no link preview
	Reply(replyMsgID). // reply to a message
	Text(ctx, "quiet reply without preview")
```

Other options include `Schedule(time.Time)`, `NoForwards()`, `Markup(...)` for inline
keyboards (see [Styled text and HTML](../helpers/styling.md)) and `SendAs(peer)`.

## Replying to updates

Inside an update handler you usually want to reply in the same chat. `sender.Reply`
takes the update's entities and the update itself and figures out the peer for you —
this is the heart of the [Echo bot](./echo-bot.md):

```go
dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
	m, ok := u.Message.(*tg.Message)
	if !ok || m.Out {
		return nil
	}
	_, err := sender.Reply(e, u).Text(ctx, m.Message)
	return err
})
```
