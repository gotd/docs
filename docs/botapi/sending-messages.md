---
sidebar_position: 3
---

# Sending messages

## Targeting chats

Outgoing methods take a `ChatID`, a sealed union you build with `ID` (numeric)
or `Username`:

```go
botapi.ID(123456789)        // a numeric chat id
botapi.Username("@channel") // an @username (leading @ optional)
```

:::info[Why a union?]
A common HTTP-client approach models a chat target as a two-field struct
(`{ID int64; Username string}`), where illegal states are representable. botapi
uses a *sealed interface* instead — an interface with an unexported marker
method and a fixed set of implementations — so an illegal target cannot be
constructed and switches over it are checked for exhaustiveness. The same
pattern applies to `InputFile`, `ReplyMarkup`, `ChatMember`, `InputMedia`, and
the inline-query result and message-content unions.
:::

## Send methods

Send methods hang off `*Bot`, take a `context.Context` first, a `ChatID`, and a
variadic of shared `SendOption`s:

```go
msg, err := bot.SendMessage(ctx, botapi.ID(chatID), "hi",
	botapi.ReplyTo(replyID),
	botapi.Silent(),
	botapi.DisableWebPagePreview(),
)
```

Common options (all `SendOption`):

| Option | Effect |
| --- | --- |
| `ReplyTo(id)` | Reply to a message. |
| `Silent()` | Send without a notification. |
| `ProtectContent()` | Disallow forwarding/saving. |
| `DisableWebPagePreview()` | No link preview. |
| `WithReplyMarkup(m)` | Attach a keyboard (see [Keyboards](#keyboards)). |
| `WithParseMode(m)` | Format text (see [Formatting](#formatting)). |

Inside a handler the [`Context`](./receiving-updates.md#the-context) shortcuts
are usually enough:

```go
c.Send("text")  // send to the update's chat
c.Reply("text") // reply to the incoming message
```

## Formatting

Pass `WithParseMode` with `ParseModeHTML`, `ParseModeMarkdownV2`, or the legacy
`ParseModeMarkdown`:

```go
bot.SendMessage(ctx, chat, "<b>bold</b> <i>italic</i>",
	botapi.WithParseMode(botapi.ParseModeHTML))

bot.SendMessage(ctx, chat, "*bold* _italic_ ||spoiler||",
	botapi.WithParseMode(botapi.ParseModeMarkdownV2))
```

## Keyboards

`ReplyMarkup` is a sealed union: `*InlineKeyboardMarkup`,
`*ReplyKeyboardMarkup`, `*ReplyKeyboardRemove`, `*ForceReply`.

### Inline keyboards

Build them with the helpers — each row is a slice of buttons:

```go
kb := botapi.InlineKeyboard(
	[]botapi.InlineKeyboardButton{
		botapi.InlineButtonData("👍", "vote:up"),
		botapi.InlineButtonData("👎", "vote:down"),
	},
	[]botapi.InlineKeyboardButton{
		botapi.InlineButtonURL("source", "https://github.com/gotd/td"),
	},
)
bot.SendMessage(ctx, chat, "Vote:", botapi.WithReplyMarkup(kb))
```

Or construct the struct directly:

```go
kb := &botapi.InlineKeyboardMarkup{
	InlineKeyboard: [][]botapi.InlineKeyboardButton{
		{
			{Text: "👍 Like", CallbackData: "vote:up"},
			{Text: "👎 Dislike", CallbackData: "vote:down"},
		},
		{
			{Text: "gotd/td", URL: "https://github.com/gotd/td"},
		},
	},
}
```

Tapping a `CallbackData` button delivers a callback query — see
[Callback & inline queries](./callback-and-inline-queries.md).

### Reply (custom) keyboards

Reply keyboards use `ReplyKeyboardMarkup` with `Button`, `ButtonContact`,
`ButtonLocation`:

```go
kb := &botapi.ReplyKeyboardMarkup{
	Keyboard: [][]botapi.KeyboardButton{
		{botapi.Button("📊 Poll"), botapi.Button("🎲 Dice")},
		{botapi.ButtonLocation("📍 Location"), botapi.ButtonContact("📇 Contact")},
	},
	ResizeKeyboard: true,
}
```

Reply-keyboard buttons arrive as plain text messages. Remove the keyboard with
`&botapi.ReplyKeyboardRemove{RemoveKeyboard: true}`.

## Other sends

Beyond text and media, there are typed sends for structured content:

```go
bot.SendChatAction(ctx, chat, botapi.ChatActionTyping)
bot.SendPoll(ctx, chat, "Question?", []string{"A", "B", "C"})
bot.SendDice(ctx, chat, botapi.DiceDie)
bot.SendLocation(ctx, chat, 55.7558, 37.6173)
bot.SendVenue(ctx, chat, 55.7520, 37.6175, "Red Square", "Moscow, Russia")
bot.SendContact(ctx, chat, "+1234567890", "Ada", "Lovelace")
```

For photos, documents, video, audio, stickers and albums, see
[Media & files](./media-and-files.md).
