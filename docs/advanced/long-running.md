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

### Banned or deleted accounts

`telegram.AuthLostKindOf(err)`, or `Kind()` of `*AuthLostError`, tells why authorization
was lost. Use it to decide whether signing in again can help:

| Kind | Error | Meaning |
| --- | --- | --- |
| `AuthLostBanned` | `USER_DEACTIVATED_BAN` | The account is banned by Telegram. |
| `AuthLostDeleted` | `USER_DEACTIVATED` | The account is deleted. |
| `AuthLostRevoked` | `SESSION_REVOKED`, `AUTH_KEY_DROP_*` | The session was terminated, or the server asked to drop the key. |
| `AuthLostUnregistered` | `AUTH_KEY_UNREGISTERED` | The key isn't authorized: usually a terminated session. |
| `AuthLostExpired` | `SESSION_EXPIRED` | The session expired. |
| `AuthLostKeyInvalid` | `AUTH_KEY_INVALID`, `AUTH_KEY_DUPLICATED` | The key is invalid or used by another connection. |

```go
OnAuthLost: func(ctx context.Context, err error) error {
	switch telegram.AuthLostKindOf(err) {
	case telegram.AuthLostBanned, telegram.AuthLostDeleted:
		markDead(account) // signing in again won't help
	default:
		askToSignIn(account)
	}
	return err
},
```

- **What clients do.** TDLib and the official apps treat any unauthorized (401) error on
  the home DC as a logout, and only TDLib adds a ban message. gotd does the same: the kind
  is logged and added to `gotd.client.auth.lost`, and bans and deletions are logged at
  error level.
- **`AUTH_KEY_DROP_*` notifications.** The server can ask a client to drop its key with
  an `updateServiceNotification` of this type, e.g. after a key leak. Telegram Desktop
  logs out on it, so gotd reports it via `OnAuthLost` as a `*tgerr.Error` with code 401
  and the notification type and text (`telegram.AuthKeyDropError`). Return `nil` to keep
  running: if the session still works, a later loss is reported again.
- **The kind isn't always known.** A ban or deletion can drop the auth key on the server.
  The server then doesn't know the key at all (transport error `-404`), gotd creates a
  new one, and the next request fails with `AUTH_KEY_UNREGISTERED`. The kind is then
  `AuthLostUnregistered`, not `AuthLostBanned`.
- **Signing in with a banned number** fails with `PHONE_NUMBER_BANNED`. The client logs it
  at error level and counts it in `gotd.client.account.restricted`.
  `telegram.IsAccountBanned(err)` matches it and `USER_DEACTIVATED_BAN`.

## Frozen accounts

Telegram can freeze an account instead of banning it. A frozen account stays signed in
and can read, but most methods fail with `FROZEN_METHOD_INVALID` (code 420, not a flood
wait) or `FROZEN_PARTICIPANT_MISSING`, until it is deleted at the end of the freeze or
an appeal succeeds. Authorization isn't lost, so `OnAuthLost` isn't called.

The freeze state comes from app config (`freeze_since_date`, `freeze_until_date`,
`freeze_appeal_url`). After a frozen error, the client fetches app config, at most once
an hour like Telegram Desktop, and calls `OnFrozen` if the account is frozen:

```go
client := telegram.NewClient(appID, appHash, telegram.Options{
	OnFrozen: func(ctx context.Context, state telegram.FreezeState) {
		alert(account, state.Since, state.Until, state.AppealURL)
	},
})
```

`OnFrozen` runs on its own goroutine with `Run`'s context. It is only called after a
request fails, so check `client.FreezeState(ctx)` at start too. Use `telegram.IsFrozen(err)`
to skip retries of refused requests. Frozen errors are counted in
`gotd.client.account.restricted`.

See [examples/account-banned](https://github.com/gotd/td/tree/main/examples/account-banned)
for a complete client.

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
