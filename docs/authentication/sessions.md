---
sidebar_position: 5
---

# Sessions and storage

A *session* is the authorization key gotd negotiates with Telegram. Persist it and your
program signs in once; lose it and you re-authenticate every run — which, for users, may
look like suspicious activity. Configure storage via `Options.SessionStorage`.

:::warning
A session file grants full access to the account. Treat it like a password: restrict
file permissions and never commit it to version control.
:::

## File storage

The simplest option writes the session to a JSON file:

```go
import "github.com/gotd/td/session"

client := telegram.NewClient(appID, appHash, telegram.Options{
	SessionStorage: &session.FileStorage{Path: "session.json"},
})
```

That is all persistence needs. Combine it with
[`client.Auth().IfNecessary`](./user.md) so the flow only runs when the stored session
is missing or invalid.

## The `Storage` interface

`SessionStorage` is a two-method interface, so you can back it with anything — a
database, a secrets manager, an encrypted blob:

```go
type Storage interface {
	LoadSession(ctx context.Context) ([]byte, error)
	StoreSession(ctx context.Context, data []byte) error
}
```

`LoadSession` returns `session.ErrNotFound` when there is no session yet. A minimal
in-memory implementation (from the
[`bot-auth-manual`](https://github.com/gotd/td/tree/main/examples/bot-auth-manual)
example):

```go
type memorySession struct {
	mux  sync.RWMutex
	data []byte
}

func (s *memorySession) LoadSession(context.Context) ([]byte, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if len(s.data) == 0 {
		return nil, session.ErrNotFound
	}
	return append([]byte(nil), s.data...), nil
}

func (s *memorySession) StoreSession(_ context.Context, data []byte) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data = append([]byte(nil), data...)
	return nil
}
```

An in-memory store means re-authenticating on every restart — fine for tests or
short-lived bots, not for production user clients.

## From the environment

`telegram.ClientFromEnvironment` and `BotFromEnvironment` derive a `FileStorage` from
`SESSION_FILE` (or `SESSION_DIR`) for you — see
[Environment helpers](../getting-started/environment-helpers.md).

## More backends

The [`gotd/contrib`](https://github.com/gotd/contrib) module provides session storage
adapters for various databases if you need shared or distributed session storage.
