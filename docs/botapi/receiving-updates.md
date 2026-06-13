---
sidebar_position: 5
---

# Receiving updates

There is **no `getUpdates` and no webhook**. Updates arrive on the persistent
MTProto connection and are dispatched through a small handler framework. You
register handlers, then call [`bot.Run(ctx)`](./getting-started.md).

## Handlers

Register handlers with the `On*` methods. A `Handler` is `func(*Context) error`:

```go
bot.OnMessage(func(c *botapi.Context) error {
	return c.Reply("you said: " + c.Message().Text)
})

bot.OnEditedMessage(handler)
bot.OnChannelPost(handler)
bot.OnCallbackQuery(handler)
bot.OnInlineQuery(handler)
bot.OnCommand("start", "Start the bot", handler)
```

:::note[Self-messages are filtered]
Updates for the bot's own outgoing messages are filtered out (the HTTP Bot API
never delivers them), so reply handlers won't answer themselves.
:::

## The Context

Each handler receives a `*Context`. It carries the `*Bot` and the `Update`, and
is *itself* a `context.Context` — pass it straight to method calls. Helpers:

| Helper | Returns |
| --- | --- |
| `c.Message()` | The effective message. |
| `c.Sender()` | The `*User` who sent it. |
| `c.Chat()` | The chat `(ChatID, ok)`. |
| `c.Send(text, ...)` | Send to the update's chat. |
| `c.Reply(text, ...)` | Reply to the incoming message. |
| `c.AnswerCallback(...)` | Acknowledge a callback query. |
| `c.AnswerInline(results, ...)` | Answer an inline query. |
| `c.Bot` | The `*Bot`, for any method. |
| `c.Update` | The raw `*Update`. |

```go
bot.OnCommand("photo", "Send a photo", func(c *botapi.Context) error {
	chat, ok := c.Chat()
	if !ok {
		return nil
	}
	_, err := c.Bot.SendPhoto(c, chat, botapi.FileURL(url), "caption")
	return err
})
```

## Predicates

Every `On*` method accepts trailing `Predicate`s (`func(*Update) bool`); the
handler runs only when **all** of them match. The first matching handler wins
across handlers.

```go
bot.OnMessage(handler, botapi.HasText(), botapi.Not(botapi.HasPrefix("/")))
```

Built-ins:

| Predicate | Matches |
| --- | --- |
| `Command(name)` | A specific `/command`. |
| `HasPrefix(s)` | Text starting with `s`. |
| `HasText()` | Any non-empty text. |
| `TextEquals(s)` | Exact text. |
| `Regex(expr)` | Text matching a regexp. |
| `ChatTypeIs(t)` | A chat type (`ChatTypePrivate`, `ChatTypeSupergroup`, …). |
| `CallbackData(s)` / `CallbackPrefix(s)` | Callback-query data. |
| `Not(p)` / `Or(p...)` | Combinators. |

A predicate is just a function — write your own:

```go
func hasPhoto(u *botapi.Update) bool {
	m := u.EffectiveMessage()
	return m != nil && len(m.Photo) > 0
}

bot.OnMessage(handlePhoto, hasPhoto)
```

## Middleware

A `Middleware` is `func(Handler) Handler`. Register global middleware with
`Use`; it wraps every handler:

```go
bot.Use(botapi.Recover(), botapi.Timeout(30*time.Second), botapi.Logging())
```

Built-ins:

* `Recover()` — turns panics into errors.
* `Timeout(d)` — bounds each handler's runtime.
* `Logging()` — logs handler outcomes.

## Groups

`Group` scopes shared predicates and middleware to a subset of handlers:

```go
admin := bot.Group(botapi.ChatTypeIs(botapi.ChatTypeSupergroup))
admin.Use(requireAdmin)
admin.OnCommand("ban", "Ban a user", banHandler)
```

Every handler registered on `admin` inherits the group's predicate
(`ChatTypeIs(...)`) and middleware (`requireAdmin`) on top of any global ones.

## Commands

`OnCommand(name, description, handler, predicates...)` registers a command
handler. On `Run`, the bot publishes all registered commands to Telegram via
`SetMyCommands`, so the client's command menu stays in sync automatically. Opt
out with `Options.DisableCommandRegistration`.

You can still drive the command list yourself with `SetMyCommands`,
`GetMyCommands` and `DeleteMyCommands`, using scopes
(`BotCommandScopeChat`, …) for per-chat menus.

The [`examples/echo`](https://github.com/gotd/botapi/tree/main/examples/echo) and
[`examples/advanced`](https://github.com/gotd/botapi/tree/main/examples/advanced)
bots show the full handler framework end to end.
