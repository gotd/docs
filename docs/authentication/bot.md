---
sidebar_position: 2
---

# Bot authentication

Bots authenticate with a single token from
[@BotFather](https://t.me/BotFather) — no phone, code, or password. You still need an
`api_id` / `api_hash` to identify your application.

## The `telegram/bot` package

For a bot that runs as a service, the
[`telegram/bot`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram/bot.html) package
does the whole setup in one call and gets the parts that are easy to get wrong right:

```go
import "github.com/gotd/td/telegram/bot"

d := tg.NewUpdateDispatcher()
b, err := bot.Start(ctx, appID, appHash, os.Getenv("BOT_TOKEN"), bot.Options{
	Dir:     "state/mybot", // session and update state
	Handler: d,
}, telegram.WithLifetime(ctx))
if err != nil {
	return err
}
defer func() { _ = b.Close() }()

return b.Wait(context.Background())
```

* **Session in `Dir`.** The bot keeps its authorization across restarts
  (`session.json`). The stored session is checked against the token: a session of
  another account fails with `bot.ErrSessionMismatch` instead of running as it.
* **Updates while stopped.** The [updates manager](../helpers/updates-recovery.md) keeps
  its state in `Dir` (`updates.json`), so messages sent while the bot was down are
  delivered after the restart, and gaps are recovered.
* **Peers.** The peer manager is fed with the entities of every update, with its cache
  on, so peers seen in updates resolve without requests.
* **Flood waits on sign in** are waited out, up to `MaxLoginWait` (5 minutes by default).
* **Token check.** A malformed token fails before connecting, with an error wrapping
  `auth.ErrInvalidBotToken` that never contains the secret. `auth.ParseBotToken` does
  the same check on its own.

Register handlers on the dispatcher before `Start`; use `b.Sender()`, `b.API()`,
`b.Peers()` and `b.Self()` from them. Use `bot.New` and `b.Start(ctx)` to set things up
between creating and starting the bot. Like the client, a `Bot` has `Stop`, `Shutdown`,
`Close`, `Done` and `Wait` — see [Running without `Run`](../advanced/without-run.md).
Use one `Dir` per bot, and don't share it between running processes.

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

`BotFromEnvironment` bundles a simpler setup — build client, run it, authenticate with
`BOT_TOKEN`, register handlers, and block — into one helper. It doesn't run the updates
manager, so updates sent while the bot was stopped can be missed:

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
