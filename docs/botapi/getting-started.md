---
sidebar_position: 2
---

# Getting started

A bot needs an **MTProto app identity** (`AppID` / `AppHash` from
[my.telegram.org](https://my.telegram.org)) and a **bot token** from
[@BotFather](https://t.me/BotFather). The app identity is required even for bots
— it identifies the *application*, not the bot, and is **not** the token.

## Your first bot

```go
bot, err := botapi.New(token, botapi.Options{AppID: appID, AppHash: appHash})
if err != nil {
	return err
}

bot.OnCommand("start", "Start the bot", func(c *botapi.Context) error {
	_, err := c.Reply("Hello!")
	return err
})

// Run connects, authorizes as a bot and serves updates until ctx is canceled.
return bot.Run(ctx)
```

[`New`](https://github.com/gotd/botapi/blob/main/bot.go) does **no network
I/O** — it just builds an unconnected bot. Register your handlers, then call
`Run`. `Run` connects, authorizes the bot with its token, publishes the commands
you registered via [`OnCommand`](./receiving-updates.md#commands), and then
serves updates until the context is canceled.

:::tip[Cancel cleanly]
Wire `Run` to an interrupt signal so Ctrl-C shuts the bot down:

```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()

if err := bot.Run(ctx); err != nil {
	log.Fatal(err)
}
```
:::

## A complete echo bot

This is the [`examples/echo`](https://github.com/gotd/botapi/tree/main/examples/echo)
bot — it greets on `/start` and echoes any other text back:

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/botapi"
)

func main() {
	log, _ := zap.NewDevelopment()
	defer func() { _ = log.Sync() }()

	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatal("APP_ID must be a number (see https://my.telegram.org)", zap.Error(err))
	}

	bot, err := botapi.New(os.Getenv("BOT_TOKEN"), botapi.Options{
		AppID:   appID,
		AppHash: os.Getenv("APP_HASH"),
		Logger:  log,
	})
	if err != nil {
		log.Fatal("Create bot", zap.Error(err))
	}

	// Middleware applies to every handler.
	bot.Use(botapi.Recover(), botapi.Timeout(30*time.Second))

	bot.OnCommand("start", "Show the welcome message", func(c *botapi.Context) error {
		_, err := c.Reply("Hi! Send me any text and I'll echo it back.")
		return err
	})

	// Any text message that is not a command.
	bot.OnMessage(func(c *botapi.Context) error {
		_, err := c.Reply(c.Message().Text)
		return err
	}, botapi.HasText(), botapi.Not(botapi.HasPrefix("/")))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Info("Starting echo bot")
	if err := bot.Run(ctx); err != nil {
		log.Fatal("Run", zap.Error(err))
	}
}
```

Run it with the three environment variables:

```bash
APP_ID=12345 APP_HASH=abcdef BOT_TOKEN=123:abc go run ./examples/echo
```

## Options

[`botapi.Options`](https://github.com/gotd/botapi/blob/main/options.go)
configures the bot. Only `AppID` and `AppHash` are required:

| Field | Purpose |
| --- | --- |
| `AppID`, `AppHash` | MTProto app identity (required). |
| `Logger` | A `*zap.Logger`. Defaults to a no-op logger. |
| `Device` | Device info reported at session init. |
| `Storage` | Persists session, peers and update state. In-memory if nil — see [Persistence](./persistence-and-pooling.md). |
| `OnStart` | Called once, after the bot is authorized and gap recovery is live. |
| `FloodWait` | Transparently retry `FLOOD_WAIT`-limited requests — see [Errors & resilience](./errors-and-resilience.md). |
| `RequestsPerSecond` | Proactive global rate limit (token bucket). |
| `DisableCommandRegistration` | Stop `Run` from publishing `OnCommand` handlers via `SetMyCommands`. |

Leaving `Storage` nil keeps the session, peers and update state in memory:
nothing survives a restart. That is fine for development; for production, give it
a [`storage.BBoltStorage`](./persistence-and-pooling.md#persistence).

Continue with [Sending messages](./sending-messages.md).
