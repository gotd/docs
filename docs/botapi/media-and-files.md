---
sidebar_position: 4
---

# Media & files

## InputFile

A file to send is an `InputFile`, a sealed union with three kinds of source:

| Constructor | Source |
| --- | --- |
| `FileID(id)` | A file already on Telegram — no upload. |
| `FileURL(url)` | A URL; Telegram fetches it server-side. |
| `FileFromPath(p)` / `FileFromBytes(b)` / `FileFromReader(r)` | A local upload. |

```go
bot.SendPhoto(ctx, chat, botapi.FileURL("https://.../cat.jpg"), "caption")
bot.SendDocument(ctx, chat, botapi.FileFromPath("/tmp/report.pdf"), "")
bot.SendVideo(ctx, chat, botapi.FileID(fileID), "")
```

## Typed media sends

Each media kind has its own method, all `func(ctx, ChatID, InputFile, caption string, ...SendOption)`:

`SendPhoto`, `SendDocument`, `SendVideo`, `SendAudio`, `SendVoice`,
`SendAnimation`, `SendVideoNote`, `SendSticker`.

They share the same [`SendOption`s](./sending-messages.md#send-methods) as
`SendMessage` (`ReplyTo`, `Silent`, `WithReplyMarkup`, `WithParseMode`, …), so
captions can be formatted and messages can carry keyboards.

### Albums

`SendMediaGroup` sends several items as one album:

```go
bot.SendMediaGroup(ctx, chat, []botapi.InputMedia{ /* uploaded items */ })
```

:::note
The high-level album API composes **uploaded** items. Sending an album of
`file_id`/URL items is deferred — see the
[roadmap](https://github.com/gotd/botapi/blob/main/docs/roadmap.md).
:::

## Receiving media

Incoming media populates the typed fields on `Message` (`Photo`, `Document`,
`Video`, `Sticker`, `Location`, `Contact`, …), each carrying a usable `file_id`.
Photos arrive as a slice of sizes, smallest to largest:

```go
bot.OnMessage(func(c *botapi.Context) error {
	photos := c.Message().Photo
	largest := photos[len(photos)-1] // last size is the highest resolution
	chat, _ := c.Chat()
	// Echo it straight back by file_id — no download/re-upload round trip.
	_, err := c.Bot.SendPhoto(c, chat, botapi.FileID(largest.FileID),
		fmt.Sprintf("%d×%d", largest.Width, largest.Height))
	return err
}, hasPhoto)

func hasPhoto(u *botapi.Update) bool {
	m := u.EffectiveMessage()
	return m != nil && len(m.Photo) > 0
}
```

Because the `file_id` is decoded locally, echoing media back never leaves
Telegram's servers — there is no download/re-upload step.

## GetFile and downloads

There is no HTTP file server in the MTProto model. `GetFile` decodes a `file_id`
**locally** (no network) and derives the `file_unique_id`; you download with
`DownloadFile` or `DownloadFileToPath`, which follow datacenter migration
automatically:

```go
f, err := bot.GetFile(ctx, fileID)        // local decode: size, unique id, etc.
n, err := bot.DownloadFile(ctx, fileID, w) // streams into an io.Writer
err = bot.DownloadFileToPath(ctx, fileID, "/tmp/out.bin")
```

A document round-trip — read its metadata, then echo it back:

```go
bot.OnMessage(func(c *botapi.Context) error {
	doc := c.Message().Document
	chat, _ := c.Chat()

	file, err := c.Bot.GetFile(c, doc.FileID)
	if err != nil {
		return err
	}
	caption := fmt.Sprintf("%s (%d bytes)\nfile_unique_id=%s",
		doc.FileName, doc.FileSize, file.FileUniqueID)

	_, err = c.Bot.SendDocument(c, chat, botapi.FileID(doc.FileID), caption)
	return err
}, hasDocument)
```

:::info[file_id and file_unique_id]
The `file_id` codec is local: botapi decodes the user's `file_id` into the
underlying MTProto location and builds the send without re-uploading. The
`file_unique_id` is derived locally too, using the TDLib scheme (web/document
exact; photos best-effort but stable per file). See the
[downloading-files](../helpers/downloading-files.md) helper for the lower-level
`fileid` machinery.
:::

The full media flow is in
[`examples/media`](https://github.com/gotd/botapi/tree/main/examples/media).
