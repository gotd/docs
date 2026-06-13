---
sidebar_position: 6
---

# Callback & inline queries

## Callback queries

A callback query arrives when a user taps an
[inline-keyboard](./sending-messages.md#inline-keyboards) button that carries
`CallbackData`. Acknowledge it with `AnswerCallback` (which clears the
client-side loading spinner), optionally showing a toast or alert, then usually
edit the original message:

```go
bot.OnCallbackQuery(func(c *botapi.Context) error {
	choice := c.Update.CallbackQuery.Data // e.g. "vote:up"

	// Acknowledge the tap. WithCallbackText shows a toast.
	if err := c.AnswerCallback(botapi.WithCallbackText("Thanks for voting!")); err != nil {
		return err
	}

	// Edit the message the button was attached to.
	msg := c.Update.CallbackQuery.Message
	if msg == nil {
		return nil
	}
	_, err := c.Bot.EditMessageText(c, botapi.ID(msg.Chat.ID), msg.MessageID,
		"You voted "+choice)
	return err
}, botapi.CallbackPrefix("vote:"))
```

The `CallbackPrefix("vote:")` predicate routes only the buttons whose data
starts with `vote:` to this handler, so different button families can have
different handlers.

See [`examples/buttons`](https://github.com/gotd/botapi/tree/main/examples/buttons)
for the complete keyboard-plus-callback flow.

## Inline queries

Inline mode lets users type your bot's `@username` followed by a query in *any*
chat. Enable it for your bot in [@BotFather](https://t.me/BotFather) first, then
answer with a list of results:

```go
bot.OnInlineQuery(func(c *botapi.Context) error {
	q := strings.TrimSpace(c.Update.InlineQuery.Query)
	if q == "" {
		return c.AnswerInline(nil)
	}
	results := []botapi.InlineQueryResult{
		&botapi.InlineQueryResultArticle{
			ID:          "upper",
			Title:       "UPPERCASE",
			Description: strings.ToUpper(q),
			InputMessageContent: &botapi.InputTextMessageContent{
				MessageText: strings.ToUpper(q),
			},
		},
	}
	return c.AnswerInline(results, botapi.WithInlineCacheTime(1))
})
```

Picking a result sends its `InputMessageContent` to the chat.

### The result and content unions

Both `InlineQueryResult` and `InputMessageContent` are sealed unions:

* **`InlineQueryResult`** — `InlineQueryResultArticle`; photo/gif/mpeg4-gif by
  URL; cached photo/gif/sticker/document/video/voice/audio by `file_id`; and
  contact/location/venue results.
* **`InputMessageContent`** — `InputTextMessageContent`, plus location, venue and
  contact content.

The runnable [`examples/inline`](https://github.com/gotd/botapi/tree/main/examples/inline)
bot offers article results that echo the query.
