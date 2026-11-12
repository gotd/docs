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

## What a session holds

The stored data (`session.Data`, JSON with a format version) is more than the home DC key:

- the auth key and salt of the home DC, and when the key was created;
- the auth keys of other DCs the client already transferred the authorization to (`DCs`),
  so after a restart file downloads and `*_MIGRATE` redirects reuse them instead of
  running a new key exchange and `auth.exportAuthorization`/`importAuthorization`;
- the server time offset the client observed (see
  [Long-running clients](../advanced/long-running.md#clock-skew)), used from the first
  message of the next run;
- the ID of the signed-in user or bot, so `client.SelfID()` answers without a request.

These fields are optional additions to format version 1: older gotd versions ignore
them, and session files written by older versions load with them unset. A file written by
a newer format version fails to load with `session.ErrVersionUnsupported` instead of
looking like a missing session.

## Logging out

`client.LogOut(ctx)` ends the authorization the way Telegram Desktop and TDLib do:

1. `auth.logOut` on the home DC (a 401 counts as success);
2. `auth.logOut` on every other DC the authorization was transferred to, then those
   connections are closed;
3. the stored session is wiped, so the next start creates a new key and is unauthorized.

```go
res, err := client.LogOut(ctx)
if err != nil {
	// The local steps already ran: the session is wiped either way.
	return err
}
// res.FutureAuthToken, if set, can be passed to a later sign-in
// (auth.sendCode logout_tokens) to skip the code.
```

Logging out is irrevocable, and your own logout isn't reported to `OnAuthLost`. The
client keeps running unauthorized on its current key until you sign in again or stop it;
gotd doesn't send `destroy_auth_key` yet.

## From the environment

`telegram.ClientFromEnvironment` and `BotFromEnvironment` derive a `FileStorage` from
`SESSION_FILE` (or `SESSION_DIR`) for you — see
[Environment helpers](../getting-started/environment-helpers.md).

## More backends

The [`gotd/contrib`](https://github.com/gotd/contrib) module provides session storage
adapters for various databases if you need shared or distributed session storage.
