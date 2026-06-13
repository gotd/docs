---
sidebar_position: 3
---

# Environment helpers

The [examples in the gotd repository](https://github.com/gotd/td/tree/main/examples)
share a tiny amount of boilerplate so each one can focus on a single feature. You will
see it everywhere, so it is worth understanding once. None of it is required for your
own programs — it is just convenience built on the public API.

## `examples.Run`

Sets up a development logger and a background context, then calls your function and
fatally logs any error:

```go
examples.Run(func(ctx context.Context, log *zap.Logger) error {
	// ... your code ...
	return nil
})
```

In your own code this is just:

```go
ctx := context.Background()
log, _ := zap.NewDevelopment()
```

## `ClientFromEnvironment` and `BotFromEnvironment`

gotd itself provides two constructors that read `APP_ID`, `APP_HASH` and the session
path from the environment, so the examples don't have to parse flags.

```go
// User/raw client built from APP_ID, APP_HASH, SESSION_FILE/SESSION_DIR.
client, err := telegram.ClientFromEnvironment(telegram.Options{
	Logger:        logzap.New(log),
	UpdateHandler: dispatcher,
})

// Bot helper: builds the client, runs it, authenticates with BOT_TOKEN,
// then calls your setup function and finally the lifecycle function.
err := telegram.BotFromEnvironment(ctx, opts,
	func(ctx context.Context, client *telegram.Client) error {
		// register handlers, etc.
		return nil
	},
	telegram.RunUntilCanceled, // block until ctx is cancelled
)
```

`RunUntilCanceled` is a ready-made lifecycle function that simply blocks until the
context is cancelled — exactly what a long-running bot wants.

## `examples.Terminal`

An [`auth.UserAuthenticator`](../authentication/user.md) that prompts on the terminal
for phone number, login code and 2FA password. It is used by every user-login example:

```go
flow := auth.NewFlow(examples.Terminal{PhoneNumber: os.Getenv("PHONE")}, auth.SendCodeOptions{})
if err := client.Auth().IfNecessary(ctx, flow); err != nil {
	return err
}
```

See [User authentication](../authentication/user.md) for how to implement your own
authenticator (web prompt, message queue, etc.).

## Running an example

```bash
export APP_ID=1234567
export APP_HASH=abcdef0123456789abcdef0123456789
export SESSION_FILE=~/session.mybot.json
export BOT_TOKEN=12345:token   # bots only

cd examples/bot-echo
go run .
```

With those basics in place, continue to [Authentication](../authentication/user.md).
