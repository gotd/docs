---
sidebar_position: 10
---

# contrib packages

[`gotd/contrib`](https://github.com/gotd/contrib) is a companion module to
[`gotd/td`](https://github.com/gotd/td). It collects optional,
batteries-included helpers — storage backends, middleware, authenticators,
streaming I/O — that bring in heavier third-party dependencies (databases,
object stores, OpenTelemetry, NTP, …). They live in a separate module so the
core `gotd/td` stays dependency-light.

```bash
go get github.com/gotd/contrib
```

Each package is independent: importing one only pulls in the dependencies that
package actually needs.

## Running & lifecycle

### `bg` — run a client in the background

The usual pattern is to do all your work inside `client.Run`'s callback. When
that does not fit your control flow,
[`bg`](https://github.com/gotd/contrib/tree/master/bg) runs the client in a
goroutine and blocks only until it is connected and **ready for requests**:

```go
import "github.com/gotd/contrib/bg"

stop, err := bg.Connect(client)
if err != nil {
	return err
}
defer func() { _ = stop() }()

// client is connected and ready here.
if _, err := client.Auth().Status(ctx); err != nil {
	return err
}
```

`Connect` accepts `bg.WithContext` to supply a base context and
`bg.WithStartupTimeout` to bound how long it waits for the client to become
ready (the client otherwise retries connection attempts indefinitely). The
returned `StopFunc` cancels the client and waits for `Run` to return.

## Middleware & RPC

These plug into `telegram.Options.Middlewares`. See
[Middleware](./middleware.md) for the full FLOOD_WAIT / rate-limit guide.

| Package | Description |
| --- | --- |
| [`middleware/floodwait`](https://github.com/gotd/contrib/tree/master/middleware/floodwait) | Catches `FLOOD_WAIT` errors and retries transparently. `Waiter` for long-running, concurrent programs; `SimpleWaiter` for one-off scripts. |
| [`middleware/ratelimit`](https://github.com/gotd/contrib/tree/master/middleware/ratelimit) | Token-bucket limiter that paces outgoing requests to stay under Telegram's limits. |
| [`invoker`](https://github.com/gotd/contrib/tree/master/invoker) | RPC invoker helpers and middlewares (debug invoker, update-aware invoker). |
| [`oteltg`](https://github.com/gotd/contrib/tree/master/oteltg) | OpenTelemetry traces and metrics for outgoing RPCs. |

## Authentication

Implementations and helpers for `telegram.AuthFlow` / `auth.UserAuthenticator`.
See [User authentication](../authentication/user.md).

| Package | Description |
| --- | --- |
| [`auth`](https://github.com/gotd/contrib/tree/master/auth) | Read credentials from constructors/env, ask interactively, compose sign-up flows. |
| [`auth/terminal`](https://github.com/gotd/contrib/tree/master/auth/terminal) | Prompts for phone, code, password and sign-up info. Uses an interactive terminal when stdin is a tty and falls back to a buffered reader for pipes, files and CI. |
| [`auth/dialog`](https://github.com/gotd/contrib/tree/master/auth/dialog) | Build an authenticator from individual dialog functions. |
| [`auth/kv`](https://github.com/gotd/contrib/tree/master/auth/kv) | Credential/session helpers over a generic key-value store. |
| [`auth/localization`](https://github.com/gotd/contrib/tree/master/auth/localization) | Localizable prompt strings for the terminal authenticator. |

## Storage — sessions, peers & state

`storage` defines the common peer abstractions; the rest are backend
implementations of the session, peer and update-state storage interfaces. See
[Sessions](../authentication/sessions.md) for how session storage fits into the
client.

| Package | Backend |
| --- | --- |
| [`storage`](https://github.com/gotd/contrib/tree/master/storage) | Common `PeerStorage` interface, peer collector, resolver cache and iteration helpers. |
| [`bbolt`](https://github.com/gotd/contrib/tree/master/bbolt) | Embedded [etcd bbolt](https://github.com/etcd-io/bbolt). |
| [`pebble`](https://github.com/gotd/contrib/tree/master/pebble) | Embedded [CockroachDB Pebble](https://github.com/cockroachdb/pebble). |
| [`redis`](https://github.com/gotd/contrib/tree/master/redis) | [Redis](https://redis.io). |
| [`s3`](https://github.com/gotd/contrib/tree/master/s3) | Any S3-compatible object store (MinIO client). |
| [`vault`](https://github.com/gotd/contrib/tree/master/vault) | [HashiCorp Vault](https://www.vaultproject.io). |

```go
import (
	"github.com/gotd/contrib/bbolt"
	"github.com/gotd/td/telegram"

	bolt "go.etcd.io/bbolt"
)

db, err := bolt.Open("session.db", 0600, nil)
if err != nil {
	return err
}

client := telegram.NewClient(appID, appHash, telegram.Options{
	SessionStorage: bbolt.NewSessionStorage(db, "session", []byte("sessions")),
})
```

## I/O & streaming

| Package | Description |
| --- | --- |
| [`tg_io`](https://github.com/gotd/contrib/tree/master/tg_io) | Partial (ranged) I/O over Telegram — download arbitrary byte ranges of a file. |
| [`partio`](https://github.com/gotd/contrib/tree/master/partio) | Chunk-based reader/writer primitives that align arbitrary reads/writes to fixed-size chunks. |
| [`http_io`](https://github.com/gotd/contrib/tree/master/http_io) | HTTP handlers built on the partial I/O primitives (e.g. serving Telegram media over HTTP). |
| [`http_range`](https://github.com/gotd/contrib/tree/master/http_range) | Parser for HTTP `Range` request headers. |

## Utilities

| Package | Description |
| --- | --- |
| [`clock`](https://github.com/gotd/contrib/tree/master/clock) | Clock sources, including an NTP-backed clock so MTProto time sync works on hosts with a skewed system clock. |

## Reference

Full per-package API documentation is on
[pkg.go.dev/github.com/gotd/contrib](https://pkg.go.dev/github.com/gotd/contrib).
