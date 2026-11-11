---
sidebar_position: 2
---

# Running without `Run`

The [`client.Run`](../getting-started/first-client.md) callback model keeps the
connection lifecycle obvious: the client is alive for exactly the duration of your
function. But sometimes you want a client that connects once and lives alongside the rest
of a long-running application — for example, a web server that occasionally calls
Telegram, or a service where several components share one client.

For that, the client has a blocking `Start` and explicit `Stop`, `Close` and `Wait`.

## Start and Close

`Start` connects in the background and returns once the client is ready for requests —
the moment `Run` would call its callback. From then on the client is usable from anywhere
until you close it:

```go
client := telegram.NewClient(appID, appHash, telegram.Options{
	SessionStorage: &session.FileStorage{Path: "session.json"},
})

// ctx bounds the startup only, like net.Dialer.DialContext.
startCtx, cancel := context.WithTimeout(ctx, time.Minute)
defer cancel()
if err := client.Start(startCtx); err != nil {
	return err
}
defer func() { _ = client.Close() }()

// Use the client from any goroutine.
if _, err := client.Self(ctx); err != nil {
	return err
}
```

`telegram.Connect(ctx, appID, appHash, opts)` does `NewClient` and `Start` in one call.

The client reconnects on transient errors, so give the startup context a deadline. On
error, `Start` leaves nothing running: no goroutines, and `Close` returns at once. It
returns the error that stopped the startup, for example a
[`*telegram.ConfigError`](./long-running.md#misconfiguration) for an invalid `api_id`.

## Waiting and stopping

| Method | What it does |
|---|---|
| `Wait(ctx)` | Blocks until the client stops; returns why (`nil` after `Stop`/`Close`, `*telegram.AuthLostError`, a permanent connection error, …). It doesn't stop the client. |
| `Done()` | A channel closed when the client stops for any reason. |
| `Stop()` | Stops the client without waiting. Safe to call from update handlers and `OnAuthLost`. |
| `Shutdown(ctx)` | Stops the client and waits until everything is released, bounded by `ctx`. |
| `Close()` | `Shutdown` without a deadline. |

`Stop` and `Close` are idempotent and safe to call concurrently. Don't call `Close` or
`Shutdown` from update handlers or `OnAuthLost`: they wait for those callbacks and would
deadlock. Use `Stop` there.

A bot that runs until it's interrupted becomes:

```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()

client := telegram.NewClient(appID, appHash, opts)
if err := client.Start(ctx, telegram.WithLifetime(ctx)); err != nil {
	return err
}
defer func() { _ = client.Close() }()

// ... sign in, register handlers ...

return client.Wait(context.Background()) // "run until disconnected"
```

## The lifetime context

By default, cancelling the context passed to `Start` after it returned doesn't stop the
client: only the startup is bounded by it. To tie the client to a context, for example
the process's signal context, pass `telegram.WithLifetime(ctx)`. When it's done, the
client stops as if `Stop` was called, and `Wait` and `Close` return `nil`.

## Rules

* `Start` and `Run` are mutually exclusive, and a client can be started once: create a
  new `Client` to start again (`ErrAlreadyStarted`, `ErrClientClosed`).
* The client is valid for requests between a successful `Start` and `Close`, as inside
  the `Run` callback.

## `gotd/contrib/bg`

Before `Start`, the [`gotd/contrib/bg`](https://github.com/gotd/contrib/tree/master/bg)
package provided the same with `bg.Connect`, which runs `Run` in a goroutine and returns a
`stop` function. It still works, but `Start` replaces it: it reports startup errors
directly and tells you why the client stopped.
