---
sidebar_position: 1
---

# Obtaining API credentials

Every gotd client needs an `api_id` (an integer) and an `api_hash` (a string). These
identify *your application* to Telegram and are required for **both** user and bot
clients.

## Get `api_id` and `api_hash`

1. Sign in at [my.telegram.org](https://my.telegram.org/apps).
2. Open **API development tools**.
3. Create an application. You will be shown your `api_id` and `api_hash`.

Follow the official guide on
[obtaining api_id](https://core.telegram.org/api/obtaining_api_id) for details.

:::danger Keep your credentials secret
Never hardcode `api_id` / `api_hash` in source you publish, and never share them — they
**cannot be rotated easily**. Read them from the environment or a config file instead.
:::

## Bot token

If you are building a bot, you *also* need a bot token from
[@BotFather](https://t.me/BotFather). The `api_id` / `api_hash` still identify your
application; the token authenticates the bot account.

## Supplying credentials

You can pass credentials explicitly:

```go
client := telegram.NewClient(appID, appHash, telegram.Options{})
```

or read them from the environment, which is what every example in the gotd repository
does (see [Environment helpers](./environment-helpers.md)):

| Variable       | Meaning                                                                |
|----------------|------------------------------------------------------------------------|
| `APP_ID`       | `api_id` from my.telegram.org                                          |
| `APP_HASH`     | `api_hash` from my.telegram.org                                        |
| `BOT_TOKEN`    | Token from @BotFather (bots only)                                     |
| `SESSION_FILE` | Path to a session file for persistent auth, e.g. `~/session.bot.json` |
| `SESSION_DIR`  | Directory for the session file, used when `SESSION_FILE` is unset      |

Next: [Your first client](./first-client.md).
