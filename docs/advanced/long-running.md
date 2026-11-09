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

### In practice

- **What it looks like.** Terminating a session from the official app (Settings →
  Devices) shows up as `AUTH_KEY_UNREGISTERED`, not `SESSION_REVOKED`. That is the
  `reason` to expect on `gotd.client.auth.lost`.
- **When it's noticed.** Detection rides on requests: the loss is found by the next
  request that needs authorization. A client that is idle, or only receives updates,
  notices it on its next call.
- **Several accounts in one process.** "Stop one account, keep the others" holds only if
  each `Run` is supervised on its own. If all accounts' `Run` calls share one `errgroup`,
  the `*AuthLostError` of one cancels the group and stops every account. Either run each
  account under its own supervisor, or return `nil` from `OnAuthLost` (log, count and
  alert) and stop that account yourself.
- **Testing it.** Use a spare account, never one in use: a terminated session can't be
  restored, and whatever happens until a new login is lost. Give the test client a
  distinctive `DeviceConfig.DeviceModel` so its session is easy to find in the app's
  device list, then terminate it there while the client sends a request every few
  seconds.

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

## Files on other datacenters

Use `client.FileDC` for file requests to other datacenters: pools are cached, use your
middlewares and follow `FILE_MIGRATE`. See [Downloading files](../helpers/downloading-files.md#other-datacenters-and-cdn).
