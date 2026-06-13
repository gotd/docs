---
sidebar_position: 3
---

# Uploading files

Telegram uploads happen in chunks, with a separate path for files over 10 MB. The
[`uploader`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/uploader.html) package
handles chunking, big-file selection and parallelism, returning a `tg.InputFileClass`
you then attach to a message.

```go
import "github.com/gotd/td/telegram/uploader"

u := uploader.NewUploader(client.API())
```

## Sources

| Method                              | Uploads from              |
|-------------------------------------|---------------------------|
| `u.FromPath(ctx, path)`             | A file on disk            |
| `u.FromBytes(ctx, name, b)`         | An in-memory `[]byte`     |
| `u.FromReader(ctx, name, r)`        | Any `io.Reader`           |
| `u.FromSource(ctx, src, url)`       | A custom `source.Source`  |

```go
file, err := u.FromPath(ctx, "report.pdf")
if err != nil {
	return err
}
```

The result is just an uploaded file — to send it you wrap it in a media option.

## Sending the uploaded file

Pair the uploader with the [message sender](./message-sender.md). Give the sender the
uploader and use `message.UploadedDocument` / `message.UploadedPhoto`:

```go
sender := message.NewSender(client.API()).WithUploader(u)

doc := message.UploadedDocument(file,
	html.String(nil, `Upload: <b>from gotd</b>`),
).
	MIME("audio/mp3").
	Filename("track.mp3").
	Audio()

if _, err := sender.Resolve("@channel").Media(ctx, doc); err != nil {
	return err
}
```

`UploadedDocument` takes optional caption styling options and exposes builder methods
(`MIME`, `Filename`, `Audio`, `Video`, `Voice`, …) to describe the file. This is the
[`bot-upload`](https://github.com/gotd/td/tree/main/examples/bot-upload) example.

## Tuning

The uploader is configured with chained options:

```go
u := uploader.NewUploader(client.API()).
	WithThreads(4).           // parallel chunk streams
	WithPartSize(512 * 1024). // chunk size
	WithProgress(progress{})  // progress tracker, see below
```

`WithProgress` takes anything implementing the `uploader.Progress` interface
(`Chunk(ctx, state) error`):

```go
type progress struct{}

func (progress) Chunk(_ context.Context, s uploader.ProgressState) error {
	log.Printf("uploaded %d/%d bytes", s.Uploaded, s.Total)
	return nil
}
```

## Streaming from a URL

`FromSource` with the
[`uploader/source`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/uploader/source.html)
HTTP source streams a remote file straight to Telegram without buffering it to disk —
the [`upload-url`](https://github.com/gotd/td/tree/main/examples/upload-url) example:

```go
import "github.com/gotd/td/telegram/uploader/source"

src := source.NewHTTPSource()
f, err := u.FromSource(ctx, src, "https://example.com/big.zip")
if err != nil {
	return err
}
sender.Self().Media(ctx, message.UploadedDocument(f))
```

Next: [Downloading files](./downloading-files.md).
