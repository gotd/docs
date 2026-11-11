---
sidebar_position: 4
---

# Tutorial: an echo bot

This ties together [bot auth](../authentication/bot.md),
[updates](./handling-updates.md) and [sending messages](./sending-messages.md) into a
complete, runnable program — a bot that replies with whatever you send it.

It is the [`bot-echo`](https://github.com/gotd/td/tree/main/examples/bot-echo) example,
annotated.

## The full program

```go
// Binary bot-echo implements a basic echo bot.
package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/bot"
	"github.com/gotd/td/tg"
)

func main() {
	// examples.Run sets up a logger and context; see Environment helpers.
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()

		appID, err := strconv.Atoi(os.Getenv("APP_ID"))
		if err != nil {
			return errors.Wrap(err, "APP_ID")
		}
		dir := os.Getenv("STATE_DIR")
		if dir == "" {
			dir = "bot-echo"
		}

		// 1. The dispatcher routes incoming updates to handlers.
		d := tg.NewUpdateDispatcher()

		// 2. The bot: session and update state in dir.
		b, err := bot.New(appID, os.Getenv("APP_HASH"), os.Getenv("BOT_TOKEN"), bot.Options{
			Dir:     dir,
			Handler: d,
			Client:  telegram.Options{Logger: logzap.New(log)},
		})
		if err != nil {
			return err
		}

		// 3. Reply to every incoming text message.
		d.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
			m, ok := u.Message.(*tg.Message)
			if !ok || m.Out {
				return nil // not a plain message, or our own — ignore
			}
			_, err := b.Sender().Reply(e, u).Text(ctx, m.Message)
			return err
		})

		// 4. Connect, sign in (or check the stored session) and start updates.
		if err := b.Start(ctx, telegram.WithLifetime(ctx)); err != nil {
			return err
		}
		defer func() { _ = b.Close() }()
		log.Info("Bot started", zap.String("username", b.Self().Username))

		// 5. Run until Ctrl+C or a fatal error, e.g. a revoked token.
		return b.Wait(context.Background())
	})
}
```

The example also answers in supergroups with `OnNewChannelMessage`, the same way.

## Walking through it

1. **Dispatcher** — `tg.NewUpdateDispatcher()` is where you register handlers. The bot
   passes it the updates.
2. **`bot.New`** — checks the token and sets up the client, the
   [updates manager](../helpers/updates-recovery.md) and the peer manager, without
   connecting yet. The session and the update state go to `dir`, so the bot doesn't sign
   in again on every start and gets the messages sent while it was stopped.
3. **Handler** — `OnNewMessage` fires for each message. We skip anything that isn't a
   `*tg.Message` and skip our own outgoing messages (`m.Out`), then
   `b.Sender().Reply(e, u)` targets the originating chat and `.Text(ctx, m.Message)`
   echoes the text back.
4. **`Start`** — connects, signs in with the token, or checks that the stored session is
   this bot's, and starts the updates manager. `WithLifetime(ctx)` stops the bot when
   `ctx` is cancelled.
5. **`Wait`** — blocks until the bot stops, and returns why. `Close` releases everything.

See [Bot authentication](../authentication/bot.md#the-telegrambot-package) for what the
`bot` package does, and [Running without `Run`](../advanced/without-run.md) for
`Start`, `Wait` and `Close`.

## Run it

```bash
export APP_ID=1234567
export APP_HASH=abcdef0123456789abcdef0123456789
export BOT_TOKEN=12345:your-bot-token
export STATE_DIR=~/bot-echo

go run .
```

Message your bot and it replies with the same text. Stop it, send a few messages, start
it again: it answers them too.

## Next steps

* Reply with **formatting** → [Styled text and HTML](../helpers/styling.md)
* Send **files** back → [Uploading files](../helpers/uploading-files.md)
* **Save** media users send you → the
  [`save-media`](https://github.com/gotd/td/tree/main/examples/save-media) example and
  [Downloading files](../helpers/downloading-files.md)
* How missed updates are recovered →
  [Update recovery](../helpers/updates-recovery.md)
