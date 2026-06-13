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

	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

func main() {
	// examples.Run sets up a logger and context; see Environment helpers.
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		// 1. The dispatcher routes incoming updates to handlers.
		dispatcher := tg.NewUpdateDispatcher()
		opts := telegram.Options{
			Logger:        logzap.New(log),
			UpdateHandler: dispatcher,
		}

		// 2. BotFromEnvironment builds the client, runs it, and signs in
		//    with BOT_TOKEN from the environment.
		return telegram.BotFromEnvironment(ctx, opts, func(ctx context.Context, client *telegram.Client) error {
			// 3. Raw API client and a message sender built on top of it.
			api := tg.NewClient(client)
			sender := message.NewSender(api)

			// 4. Reply to every incoming text message.
			dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
				m, ok := u.Message.(*tg.Message)
				if !ok || m.Out {
					return nil // not a plain message, or our own — ignore
				}
				_, err := sender.Reply(e, u).Text(ctx, m.Message)
				return err
			})
			return nil
		}, telegram.RunUntilCanceled) // 5. Block until cancelled.
	})
}
```

## Walking through it

1. **Dispatcher** — `tg.NewUpdateDispatcher()` is both the thing you register handlers
   on and the `UpdateHandler` the client delivers updates to.
2. **`BotFromEnvironment`** — reads `APP_ID`, `APP_HASH`, `BOT_TOKEN` and the session
   path, then handles connect + auth so the inner function starts already signed in.
3. **Sender** — `message.NewSender` wraps the raw API with the
   [message builder](./sending-messages.md).
4. **Handler** — `OnNewMessage` fires for each message. We skip anything that isn't a
   `*tg.Message` and skip our own outgoing messages (`m.Out`), then
   `sender.Reply(e, u)` targets the originating chat and `.Text(ctx, m.Message)` echoes
   the text back.
5. **`RunUntilCanceled`** — keeps the client connected so updates keep flowing until you
   stop the process.

## Run it

```bash
export APP_ID=1234567
export APP_HASH=abcdef0123456789abcdef0123456789
export BOT_TOKEN=12345:your-bot-token
export SESSION_FILE=~/session.echo.json

go run .
```

Message your bot and it replies with the same text.

## Next steps

* Reply with **formatting** → [Styled text and HTML](../helpers/styling.md)
* Send **files** back → [Uploading files](../helpers/uploading-files.md)
* **Save** media users send you → the
  [`save-media`](https://github.com/gotd/td/tree/main/examples/save-media) example and
  [Downloading files](../helpers/downloading-files.md)
* Never miss a message after downtime →
  [Update recovery](../helpers/updates-recovery.md)
