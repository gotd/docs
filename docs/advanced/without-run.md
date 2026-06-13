---
sidebar_position: 2
---

# Running without `Run`

The [`client.Run`](../getting-started/first-client.md) callback model keeps the
connection lifecycle obvious: the client is alive for exactly the duration of your
function. But sometimes you want a client that connects once and lives alongside the rest
of a long-running application — for example, a web server that occasionally calls
Telegram.

The [`gotd/contrib/bg`](https://github.com/gotd/contrib/tree/master/bg) package provides
this with `bg.Connect`:

```go
import "github.com/gotd/contrib/bg"

client, err := telegram.ClientFromEnvironment(telegram.Options{})
if err != nil {
	return err
}

stop, err := bg.Connect(client)
if err != nil {
	return err
}
defer func() { _ = stop() }()

// The client is connected for as long as you like — use it from anywhere.
if _, err := client.Auth().Status(ctx); err != nil {
	return err
}
```

`bg.Connect` runs the client in a background goroutine and returns a `stop` function that
disconnects it. This is the
[`bg-run`](https://github.com/gotd/td/tree/main/examples/bg-run) example.

## Trade-offs

* You are responsible for calling `stop()` to release resources cleanly.
* Errors from the background loop surface through the `stop()` return value rather than
  from a single `Run` call, so handle them there.
* The `Run` model is still preferred for self-contained programs and bots — reach for
  `bg.Connect` only when an external lifecycle (an HTTP server, a long-lived service)
  owns the process.
