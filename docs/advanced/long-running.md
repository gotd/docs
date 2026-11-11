---
sidebar_position: 7
---

# Long-running clients

Services that run for weeks, often with many accounts in one process, need to detect
broken connections and lost authorizations instead of hanging or retrying forever.

## Lost authorization

When a session is terminated from another device, requests start failing with
`AUTH_KEY_UNREGISTERED`, `SESSION_REVOKED` and similar errors. `OnAuthLost` is called once
when this happens to an authorized client. Return an error to stop `Run` with
`*telegram.AuthLostError`, so one account can be stopped and reported while others keep
running:

```go
client := telegram.NewClient(appID, appHash, telegram.Options{
	OnAuthLost: func(ctx context.Context, err error) error {
		alert(account, err)
		return err // stop Run
	},
})

err := client.Run(ctx, run)
var lost *telegram.AuthLostError
if errors.As(err, &lost) {
	// Ask the user to sign in again.
}
```

Return `nil` to keep running, e.g. to sign in again in the same `Run`. The callback runs
on its own goroutine with `Run`'s context, so the failed request returns right away, and
`Run` waits for it before returning. A client counts as authorized after any successful
request that needs authorization, including one restored from session storage. When the
connect-time `Self` call finds a lost authorization, `OnAuthLost` is called first, then
`OnSelfError` if it returned `nil`. Use `telegram.IsAuthLost(err)` to check errors
yourself.

## Connection events

`OnConnectionState` reports the primary connection state (connecting, ready,
disconnected), and `OnDead` reports dead connections. To know which DC and which kind
of connection an event is about, use the event variants:

```go
telegram.Options{
	OnConnectionStateEvent: func(e telegram.ConnectionStateEvent) {
		// e.DC, e.State
	},
	OnConnDead: func(e telegram.ConnDeadEvent) {
		// e.DC, e.Kind (ConnKindPrimary, ConnKindData, ConnKindCDN), e.Err
	},
}
```

## Stuck requests

A request that the server acknowledged but never answered (a half-dead connection, a
slow proxy) waits as long as its context allows. With `RequestTimeout`, the client
reconnects when the deadline passes and resends the request once on the new connection,
like Telegram Desktop does:

```go
telegram.Options{
	RequestTimeout: func(input bin.Encoder) time.Duration {
		switch input.(type) {
		case *tg.MessagesSendMessageRequest:
			return 0 // never send twice
		default:
			return 30 * time.Second
		}
	},
	// Liveness of connections.
	PingInterval: 30 * time.Second,
	PingTimeout:  10 * time.Second,
}
```

Without `RequestTimeout`, a request is resent on the same connection while the server
hasn't acknowledged it: every 5 seconds, up to 5 times (`RetryInterval`, `MaxRetries` in
`telegram.Options`). After the last resend the client keeps waiting
for the answer, bounded by the request's context: a slow server (answers after half a
minute happen on a busy DC) costs latency, not a failed request. A dead connection is
detected by pings and closed, and its pending requests fail.

## Lost connection state

The server can forget the state of a connection, for example after the account signed in
on another device. Requests then fail with `CONNECTION_NOT_INITED` or
`CONNECTION_LAYER_INVALID`. The client sends `initConnection` again with the failed
request and keeps doing so until the server accepts one, as Telegram Desktop and TDLib
do, so these errors don't reach your code and you don't need to restart.

## Misconfiguration

Errors of the connection setup that reconnecting can't fix — `API_ID_INVALID`,
`API_ID_PUBLISHED_FLOOD`, `CONNECTION_API_ID_INVALID` — stop `Run` (or fail `Start`)
with `*telegram.ConfigError` instead of reconnecting forever:

```go
var cfgErr *telegram.ConfigError
if errors.As(err, &cfgErr) {
	// Check APP_ID and APP_HASH; cfgErr.Err is the RPC error.
}
```

## Files on other datacenters

Use `client.FileDC` for file requests to other datacenters: pools are cached, use your
middlewares and follow `FILE_MIGRATE`. See [Downloading files](../helpers/downloading-files.md#other-datacenters-and-cdn).
