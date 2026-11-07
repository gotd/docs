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

Return `nil` to keep running, e.g. to sign in again in the same `Run`. A lost
authorization found on connect is reported via `OnSelfError`. Use
`telegram.IsAuthLost(err)` to check errors yourself.

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
