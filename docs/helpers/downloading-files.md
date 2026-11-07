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

## Resuming, size and progress

To resume a partial download, start from an offset. `ToPath` keeps the existing file
content before the offset, so a `.part` file can be continued after a restart. Pass the
known size (e.g. `Document.Size`) so workers stop at the end of the file instead of
probing past it, and a progress callback:

```go
_, err := client.Download(location).
	WithOffset(offset). // divisible by the part size (4 KB with WithVerify(true))
	WithSize(doc.Size).
	WithProgress(func(done, total int64) {
		fmt.Printf("%d/%d\n", done, total)
	}).
	ToPath(ctx, "video.mp4.part")
```

`done` counts from the start of the file, including the offset.

To check what is already on disk without downloading it, verify the local data
against the file hashes from the server (`upload.getFileHashes`). `Verify` returns the
length of the longest verified prefix, which is a safe offset to resume from:

```go
f, _ := os.Open("video.mp4.part")
st, _ := f.Stat()
valid, err := client.Download(location).Verify(ctx, f, st.Size())
// Round down to the part size unless verification is enabled.
offset := valid - valid%(512*1024)
```

## Expired file references

File references expire, and a long download (or a download from an old message)
fails with `FILE_REFERENCE_EXPIRED`. Give the builder a refresh function: on a file
reference error the location is refreshed once for all workers and the download
continues from the same offset. The
[`fileref`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/downloader/fileref.html)
package fetches the origin of the file again:

```go
import "github.com/gotd/td/telegram/downloader/fileref"

_, err := client.Download(location).
	WithRefresh(fileref.Message(client.API(), peer, msgID, location)).
	ToPath(ctx, path)
```

There are also `fileref.Story` and `fileref.UserPhoto`, and `downloader.IsFileReferenceError`
to detect the error yourself.

## Other datacenters and CDN

Files live on the datacenter in their `DCID`. `client.FileDC(ctx, dc)` returns an invoker
for that DC: the pool is created once and cached, wrapped with your
`Options.Middlewares`, uses the primary session for the home DC, and follows
`FILE_MIGRATE`. Pool size is `Options.FilePoolSize` (default 4).

Downloads through such an invoker can still use CDN redirects with `DownloadVia`:

```go
inv, err := client.FileDC(ctx, doc.DCID)
if err != nil {
	return err
}
_, err = client.DownloadVia(inv, location).ToPath(ctx, path)
```

For a custom `downloader.Client`, wrap it with
`downloader.WithCDNProvider(c, client.CDNProvider())`.

## Sharing a download window

`WithThreads` limits workers of one download. To share a budget between concurrent
downloads to the same DC, use a `Scheduler`: an adaptive limit of in-flight part
requests that grows on success and halves on flood waits and timeouts, like the
download window of Telegram Desktop:

```go
sched := downloader.NewScheduler(downloader.SchedulerOptions{Start: 4, Max: 16})
d := client.Downloader().WithScheduler(sched) // one scheduler per DC

_, err := d.Download(api, location).WithThreads(16).ToPath(ctx, path)

// For metrics.
snap := sched.Snapshot() // Window, InFlight, Waiting, Decreases
```

To use schedulers for `client.Download` and `client.DownloadWeb`, set the options on the
client. It creates one scheduler per DC on first use:

```go
client := telegram.NewClient(appID, appHash, telegram.Options{
	DownloadScheduler: &downloader.SchedulerOptions{Start: 4, Max: 16},
})

snap := client.DownloadScheduler(dc).Snapshot()
```

`DownloadVia` can't know the DC of its invoker, so it uses a scheduler only when you
pass one with `WithScheduler`.

## Bulk downloads

To back up many files, combine the downloader with a
[pagination iterator](./query-iterators.md) and a concurrency limiter such as
`golang.org/x/sync/errgroup`, optionally throttled by the
[ratelimit middleware](./middleware.md). The
[`gif-download`](https://github.com/gotd/td/tree/main/examples/gif-download) example
downloads every saved GIF this way.
