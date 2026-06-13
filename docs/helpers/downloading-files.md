---
sidebar_position: 4
---

# Downloading files

The [`downloader`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/downloader.html)
package fetches files by their `tg.InputFileLocationClass`, transparently following CDN
redirects and downloading in parallel chunks.

```go
import "github.com/gotd/td/telegram/downloader"

d := downloader.NewDownloader()
```

The `telegram.Client` also exposes a ready-made one as `client.Downloader()`.

## Downloading to a destination

`Download` returns a builder with terminal methods:

```go
// To a file on disk.
_, err := d.Download(client.API(), location).ToPath(ctx, "out.jpg")

// To any io.Writer (a buffer, an HTTP response, …).
_, err = d.Download(client.API(), location).Stream(ctx, w)
```

Both return the `tg.StorageFileTypeClass` describing the downloaded content.

## Getting a location

You rarely build an `InputFileLocation` by hand. Most media you encounter — in a message,
a participant's photo, a document — exposes one. The query helpers make this easy: an
`Elem.File()` from the [messages iterator](./query-iterators.md) gives you both a name
and a location, as in the
[`save-media`](https://github.com/gotd/td/tree/main/examples/save-media) example:

```go
import "github.com/gotd/td/telegram/query/messages"

file, ok := messages.Elem{Msg: msg}.File()
if !ok {
	return nil // message has no downloadable media
}

if _, err := d.Download(client.API(), file.Location).ToPath(ctx, file.Name); err != nil {
	return err
}
```

## Tuning

```go
d := downloader.NewDownloader().
	WithPartSize(512 * 1024). // chunk size (must divide 1 MB / be a multiple of 4 KB)
	WithAllowCDN(true)         // allow CDN redirects
```

CDN downloads must also be permitted on the client via `Options.AllowCDN`.

## Bulk downloads

To back up many files, combine the downloader with a
[pagination iterator](./query-iterators.md) and a concurrency limiter such as
`golang.org/x/sync/errgroup`, optionally throttled by the
[ratelimit middleware](./middleware.md). The
[`gif-download`](https://github.com/gotd/td/tree/main/examples/gif-download) example
downloads every saved GIF this way.
