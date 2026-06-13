---
sidebar_position: 2
---

# Styled text and HTML

Telegram represents formatting as a list of *message entities* (offset, length, type)
alongside the plain text. Building those by hand is error-prone, so gotd offers two
approaches: a typed `styling` DSL and an HTML parser.

## The `styling` package

[`message/styling`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/message/styling.html)
provides one function per entity type. Pass them to `StyledText`:

```go
import "github.com/gotd/td/telegram/message/styling"

sender.Resolve("@channel").StyledText(ctx,
	styling.Plain("Normal, "),
	styling.Bold("bold, "),
	styling.Italic("italic, "),
	styling.Code("monospace"),
	styling.Plain(" and a "),
	styling.TextURL("link", "https://gotd.dev"),
)
```

Available options include `Plain`, `Bold`, `Italic`, `Underline`, `Strike`, `Code`,
`Pre(text, language)`, `Spoiler`, `Blockquote(text, collapsed)`, `TextURL(text, url)`,
`URL`, `Mention`, `MentionName(text, user)`, `Hashtag`, `Cashtag`, `BotCommand`,
`Email`, `Phone`, `BankCard` and `CustomEmoji(text, documentID)`.

The same options are used as **captions** when sending media, and by
[inline results](#inline-bot-results).

## HTML

If you already have HTML (for example from a templating layer), the
[`message/html`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/message/html.html)
package parses Telegram's
[supported HTML subset](https://core.telegram.org/bots/api#html-style) into a styled
option:

```go
import "github.com/gotd/td/telegram/message/html"

sender.Self().StyledText(ctx,
	html.String(nil, `Upload: <b>done</b> — <i>see attached</i>`),
)
```

The first argument is an optional resolver for `tg://user?id=...` mention links; pass
`nil` when you don't use them. `html.Format(...)` accepts a format string.

## Inline keyboards

Buttons are built with the
[`message/markup`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/message/markup.html)
package and attached with `Markup`:

```go
import "github.com/gotd/td/telegram/message/markup"

sender.Resolve("@user").Markup(
	markup.InlineKeyboard(
		markup.Row(
			markup.Callback("Yes", []byte("yes")),
			markup.Callback("No", []byte("no")),
		),
	),
).Text(ctx, "Proceed?")
```

Handle presses with `dispatcher.OnBotCallbackQuery` (see
[Handling updates](../basics/handling-updates.md)).

## Inline bot results

For inline bots, the
[`message/inline`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/message/inline.html)
package builds answers, and result messages reuse `styling`:

```go
import "github.com/gotd/td/telegram/message/inline"

dispatcher.OnBotInlineQuery(func(ctx context.Context, e tg.Entities, u *tg.UpdateBotInlineQuery) error {
	_, err := inline.New(api, rand.Reader, u.QueryID).
		Set(ctx,
			inline.Article("Styled greeting",
				inline.MessageStyledText(
					styling.Bold("Hello"),
					styling.Plain(", "),
					styling.Italic(u.Query),
				),
			).ID("greeting").Description("Send a styled greeting"),
		)
	return err
})
```

This is the [`bot-inline`](https://github.com/gotd/td/tree/main/examples/bot-inline)
example.

## Rich (page-block) messages

For instant-view-style structured content — titles, headers, paragraphs, lists,
dividers — the
[`message/rich`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/message/rich.html)
package builds `RichMessage` payloads, as shown in
[`rich-message`](https://github.com/gotd/td/tree/main/examples/rich-message):

```go
import "github.com/gotd/td/telegram/message/rich"

msg := rich.New(
	rich.Title(rich.Plain("gotd")),
	rich.Paragraph(rich.Concat(
		rich.Plain("Mixing "), rich.Bold(rich.Plain("bold")),
		rich.Plain(" and "), rich.Italic(rich.Plain("italic")), rich.Plain("."),
	)),
	rich.List(
		rich.ListItem(rich.Plain("first")),
		rich.ListItem(rich.Plain("second")),
	),
).Input()

sender.Self().RichMessage(ctx, msg)
```
