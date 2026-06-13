---
sidebar_position: 1
---

# User authentication

Signing in as a user means a phone number, a login code Telegram sends you, and — if the
account has it enabled — a 2FA password. gotd wraps this multi-step dance in
[`auth.Flow`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/auth.html).

## The flow

You provide an [`auth.UserAuthenticator`](#the-userauthenticator-interface) — something
that can supply the phone, code and password on demand — and gotd drives the rest:

```go
import "github.com/gotd/td/telegram/auth"

flow := auth.NewFlow(authenticator, auth.SendCodeOptions{})

if err := client.Auth().IfNecessary(ctx, flow); err != nil {
	return err
}
```

`client.Auth().IfNecessary` first checks whether the stored session is already
authorized and only runs the flow if needed — combine it with a
[persistent session](./sessions.md) so users log in just once.

You can also check status yourself:

```go
status, err := client.Auth().Status(ctx)
if err != nil {
	return err
}
if !status.Authorized {
	if err := client.Auth().IfNecessary(ctx, flow); err != nil {
		return err
	}
}
```

## The `UserAuthenticator` interface

```go
type UserAuthenticator interface {
	Phone(ctx context.Context) (string, error)
	Password(ctx context.Context) (string, error)
	Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error)
	AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error
	SignUp(ctx context.Context) (UserInfo, error)
}
```

You rarely implement all of it by hand. gotd ships constructors that fill in the boring
parts:

| Constructor                                  | Use when                                              |
|----------------------------------------------|-------------------------------------------------------|
| `auth.Constant(phone, password, code)`       | You already know the phone and 2FA password           |
| `auth.CodeOnly(phone, code)`                  | Account has **no** 2FA password                       |
| `auth.Env(prefix, code)`                      | Read phone/password from environment variables        |

Only the **code** truly has to come from the user, because Telegram sends it live. Wrap
a function with `auth.CodeAuthenticatorFunc`:

```go
codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

flow := auth.NewFlow(
	// If the account has no 2FA password, use auth.CodeOnly instead.
	auth.Constant(phone, password, auth.CodeAuthenticatorFunc(codePrompt)),
	auth.SendCodeOptions{},
)
if err := flow.Run(ctx, client.Auth()); err != nil {
	return err
}
```

## A reusable terminal authenticator

For a CLI, implement the interface once. This is essentially what `examples.Terminal`
does:

```go
type termAuth struct {
	phone string
}

func (a termAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (a termAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func (a termAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	pwd, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(pwd)), nil
}

// Refuse sign-up and ToS in this minimal example.
func (a termAuth) AcceptTermsOfService(context.Context, tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tg.HelpTermsOfService{}}
}

func (a termAuth) SignUp(context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("sign up not implemented")
}
```

## Putting it together

```go
client := telegram.NewClient(appID, appHash, telegram.Options{
	SessionStorage: &session.FileStorage{Path: "session.json"},
})

flow := auth.NewFlow(termAuth{phone: phone}, auth.SendCodeOptions{})

return client.Run(ctx, func(ctx context.Context) error {
	if err := client.Auth().IfNecessary(ctx, flow); err != nil {
		return err
	}
	self, err := client.Self(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("signed in as %s\n", self.Username)
	return nil
})
```

:::tip
The runnable [`send-message`](https://github.com/gotd/td/tree/main/examples/send-message)
example shows code login plus persistence end to end.
:::

See also: [Bot authentication](./bot.md), [QR login](./qr-login.md),
[Two-factor passwords](./two-factor.md), [Sessions and storage](./sessions.md).
