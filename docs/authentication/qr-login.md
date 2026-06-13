---
sidebar_position: 3
---

# QR code login

QR login authenticates a **user** by displaying a QR code that you scan with an
already-signed-in Telegram app (Settings → Devices → Link Desktop Device). It avoids
typing a phone number and login code.

## How it works

gotd exports a login token, you render it as a QR code, and Telegram delivers an
[`UpdateLoginToken`](https://core.telegram.org/constructor/updateLoginToken) update the
moment the code is scanned. Because that update arrives asynchronously, QR login
**requires an update dispatcher**.

The `qrlogin.OnLoginToken` helper registers a handler on your dispatcher and returns a
channel that fires on scan:

```go
import (
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
)

dispatcher := tg.NewUpdateDispatcher()
loggedIn := qrlogin.OnLoginToken(&dispatcher)

client := telegram.NewClient(appID, appHash, telegram.Options{
	UpdateHandler: dispatcher,
})
```

## Driving the flow

`client.QR().Auth` exports the token, calls your `show` callback to display it, and
blocks until the scan signal arrives (re-showing if the token expires):

```go
show := func(ctx context.Context, token qrlogin.Token) error {
	// token.URL() is the "tg://login?token=..." string to encode as a QR.
	fmt.Println("Scan this URL as a QR code:", token.URL())
	return nil
}

if _, err := client.QR().Auth(ctx, loggedIn, show); err != nil {
	return err
}
```

`Token` also offers `token.Image(level)` to get an `image.Image` directly, and
`token.Expires()` for the expiry time.

## Handling 2FA

If the account has a 2FA password, `Auth` fails with `SESSION_PASSWORD_NEEDED`. Detect
it and complete the sign-in with the password:

```go
import "github.com/gotd/td/tgerr"

if _, err := client.QR().Auth(ctx, loggedIn, show); err != nil {
	if !tgerr.Is(err, "SESSION_PASSWORD_NEEDED") {
		return err
	}
	// Prompt for the 2FA password and finish.
	if _, err := client.Auth().Password(ctx, password); err != nil {
		return err
	}
}
```

See [Two-factor passwords](./two-factor.md) for secure handling of the password itself.

## Rendering in the terminal

The [`userbot`](https://github.com/gotd/td/tree/main/examples/userbot) and
[`call`](https://github.com/gotd/td/tree/main/examples/call) examples use
`github.com/mdp/qrterminal/v3` to draw the QR in the console:

```go
show := func(ctx context.Context, token qrlogin.Token) error {
	qrterminal.Generate(token.URL(), qrterminal.L, os.Stderr)
	return nil
}
```

The reusable `examples.QRAuth` helper bundles the show callback, the `loggedIn` channel
and the 2FA fallback together.
