---
sidebar_position: 2
---

# Bot authentication

Bots authenticate with a single token from
[@BotFather](https://t.me/BotFather) — no phone, code, or password. You still need an
`api_id` / `api_hash` to identify your application.

## Signing in with a token

```go
client := telegram.NewClient(appID, appHash, telegram.Options{
	SessionStorage: &session.FileStorage{Path: "session.bot.json"},
})

return client.Run(ctx, func(ctx context.Context) error {
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return err
	}
	if !status.Authorized {
		if _, err := client.Auth().Bot(ctx, "12345:your-bot-token"); err != nil {
			return err
		}
	}
	// Bot is authorized; use client.API() from here.
	return nil
})
```

Checking `Status` first means you only spend the token call on the first run — after
that the [stored session](./sessions.md) is reused.

## The `BotFromEnvironment` shortcut

For a typical bot, gotd bundles the whole setup — build client, run it, authenticate
with `BOT_TOKEN`, register handlers, and block — into one helper:

```go
dispatcher := tg.NewUpdateDispatcher()
opts := telegram.Options{
	Logger:        logzap.New(log),
	UpdateHandler: dispatcher,
}

return telegram.BotFromEnvironment(ctx, opts,
	func(ctx context.Context, client *telegram.Client) error {
		// Register update handlers here.
		api := tg.NewClient(client)
		_ = api
		return nil
	},
	telegram.RunUntilCanceled,
)
```

It reads `APP_ID`, `APP_HASH`, `BOT_TOKEN` and the session path from the environment.
This is the foundation of the [Echo bot tutorial](../basics/echo-bot.md).

## One-shot bots

If your bot just performs an action and exits (for example,
[`bot-upload`](https://github.com/gotd/td/tree/main/examples/bot-upload)), set
`NoUpdates: true` so the client doesn't subscribe to the update stream:

```go
opts := telegram.Options{NoUpdates: true}
```

## Manual setup without environment variables

You don't have to use the environment. The
[`bot-auth-manual`](https://github.com/gotd/td/tree/main/examples/bot-auth-manual)
example builds everything explicitly with a custom session store — useful when
embedding gotd in a larger application. See [Sessions and storage](./sessions.md) for
how to implement `session.Storage`.
