---
sidebar_position: 3
---

# Handling updates

Updates are how Telegram pushes events to your client: new messages, edits, typing
notifications, inline queries, and dozens more. You receive them by setting an
`UpdateHandler` in `telegram.Options`.

## The update dispatcher

`tg.NewUpdateDispatcher` returns a handler that routes each update type to a callback
you register. Pass it as both the dispatcher you register on and the `UpdateHandler`:

```go
dispatcher := tg.NewUpdateDispatcher()

client := telegram.NewClient(appID, appHash, telegram.Options{
	UpdateHandler: dispatcher,
})

dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
	m, ok := u.Message.(*tg.Message)
	if !ok {
		return nil
	}
	fmt.Println("message:", m.Message)
	return nil
})
```

Common handlers:

| Method                          | Fires on                                  |
|---------------------------------|-------------------------------------------|
| `OnNewMessage`                  | New message in a private chat or group    |
| `OnNewChannelMessage`           | New message in a channel/supergroup       |
| `OnEditMessage`                 | A message was edited                      |
| `OnBotCallbackQuery`            | Inline keyboard button pressed (bots)     |
| `OnBotInlineQuery`              | Inline query to your bot                  |
| `OnUserTyping` / `OnChatUserTyping` | Typing indicators                     |

There are handlers for the full update set — see the
[`tg` reference](https://ref.gotd.dev/pkg/github.com/gotd/td/tg.html).

The dispatcher can be passed by value: copies share handlers, so you can register
handlers (including `OnFallback`) after passing it to the client, even while updates
are being handled.

### One handler for all chats

Private chats, basic groups and channels have separate update types. The `OnAny*`
helpers register one handler for all of them:

```go
dispatcher.OnAnyNewMessage(func(ctx context.Context, e tg.Entities, msg tg.MessageClass) error {
	return archive(msg)
})
dispatcher.OnAnyEditMessage(onEdit)
dispatcher.OnAnyDeleteMessages(func(ctx context.Context, e tg.Entities, channelID int64, ids []int) error {
	return nil // channelID is 0 for private chats and basic groups
})
dispatcher.OnAnyReadInbox(onRead)  // also OnAnyReadOutbox
dispatcher.OnAnyTyping(onTyping)

// Called for every dispatched update before its handler, e.g. for metrics.
dispatcher.OnBefore(func(ctx context.Context, e tg.Entities, u tg.UpdateClass) {
	updatesTotal.Inc()
})
```

## The `Entities` argument

Updates reference users and chats by ID. The `tg.Entities` value passed to each handler
resolves those IDs to full objects that arrived alongside the update:

```go
dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
	m, ok := u.Message.(*tg.Message)
	if !ok {
		return nil
	}
	if peer, ok := m.PeerID.(*tg.PeerUser); ok {
		if user, ok := e.Users[peer.UserID]; ok {
			fmt.Printf("from %s\n", user.FirstName)
		}
	}
	return nil
})
```

To call handlers yourself with results of other methods (e.g. replaying stored updates or
a `getDifference` result), build the entities with
`tg.EntitiesFrom(users, chats)` or `tg.EntitiesFromResult(result)`.

## Filtering your own messages

For userbots, skip outgoing messages or you may echo yourself into a loop:

```go
if !ok || m.Out {
	return nil // outgoing, ignore
}
```

## Keeping the client alive

Updates only arrive while the client is connected, so a bot must block inside `Run`
instead of returning immediately. `telegram.RunUntilCanceled` does exactly that:

```go
return client.Run(ctx, func(ctx context.Context) error {
	// ... authenticate, register handlers ...
	return telegram.RunUntilCanceled(ctx, client)
})
```

`BotFromEnvironment` accepts `RunUntilCanceled` as its lifecycle function so you don't
write this by hand — see the [Echo bot tutorial](./echo-bot.md).

## Don't miss updates

The plain dispatcher delivers updates as they arrive, but a client that was offline can
**miss** updates, and channel updates need explicit gap handling. For correctness across
reconnects, wrap the dispatcher in the
[update-recovery engine](../helpers/updates-recovery.md).
