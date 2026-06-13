---
sidebar_position: 4
---

# Two-factor passwords

Telegram's two-step verification (2FA) adds a cloud password on top of the login code.
gotd never sends the password over the wire — it computes an
[SRP](https://core.telegram.org/api/srp) proof locally and sends only that.

## With the auth flow

When you use [`auth.Flow`](./user.md), the password comes from your
`UserAuthenticator.Password` method. If the account has a password, gotd calls it
automatically after the code step:

```go
func (a termAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	pwd, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(pwd)), nil
}
```

`auth.Constant(phone, password, code)` and `auth.Env(prefix, code)` supply the password
for you when you already have it.

## Completing 2FA directly

After a flow that stopped at `SESSION_PASSWORD_NEEDED` (for example
[QR login](./qr-login.md)), finish with:

```go
if _, err := client.Auth().Password(ctx, password); err != nil {
	return err
}
```

## Keeping the password out of plain memory

A plain Go `string` lingers in memory and may be swapped to disk. For hardened
applications, the [`secure-password`](https://github.com/gotd/td/tree/main/examples/secure-password)
example keeps the password in a [memguard](https://github.com/awnumar/memguard)
enclave and computes the SRP proof straight from it, never materializing a string. It
does so by overriding `PasswordHash` on the authenticator:

```go
import "github.com/gotd/td/telegram/auth/srpguard"

type securePassword struct {
	auth.UserAuthenticator
	enclave *memguard.Enclave
}

func (s securePassword) PasswordHash(
	ctx context.Context, p *tg.AccountPassword,
) (*tg.InputCheckPasswordSRP, error) {
	return srpguard.Enclave(s.enclave)(ctx, p)
}
```

Use this approach when running on shared or untrusted hosts.
