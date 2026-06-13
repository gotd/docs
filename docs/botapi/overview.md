---
sidebar_position: 1
---

# Overview

[`github.com/gotd/botapi`](https://github.com/gotd/botapi) is a Telegram
[Bot API](https://core.telegram.org/bots/api) library for Go, implemented
directly over **MTProto** using [`gotd/td`](https://github.com/gotd/td) — *not*
over HTTP to `api.telegram.org`.

It exposes the familiar Bot API surface (types, methods, updates) but speaks
MTProto on a persistent connection. That sidesteps the Bot API server's rate
limits, removes the `getUpdates`/webhook round trip, and keeps the raw
`gotd/td` client one method call away ([`Bot.Raw()`](./persistence-and-pooling.md#the-escape-hatch))
for anything not yet covered.

```go
package main

import (
	"context"

	"github.com/gotd/botapi"
)

func main() {
	// Bots still need an MTProto app identity (https://my.telegram.org).
	// This is NOT the bot token.
	bot, err := botapi.New("<bot-token>", botapi.Options{
		AppID:   123456,
		AppHash: "<app-hash>",
	})
	if err != nil {
		panic(err)
	}

	bot.OnCommand("start", "Start the bot", func(c *botapi.Context) error {
		_, err := c.Reply("Hello!")
		return err
	})

	// Connects, authorizes as a bot, and serves updates until ctx is cancelled.
	if err := bot.Run(context.Background()); err != nil {
		panic(err)
	}
}
```

## Why MTProto instead of HTTP

|  | HTTP Bot API client | botapi |
| --- | --- | --- |
| Transport | HTTPS to `api.telegram.org` | MTProto via `gotd/td` |
| Updates | `getUpdates` long-poll / webhook | persistent connection, no polling |
| Rate limits | Bot API server limits | MTProto limits only |
| `file_id` | opaque, must round-trip the server | local [`fileid`](../helpers/downloading-files.md) codec |
| Escape hatch | none | raw `*tg.Client` via `Bot.Raw()` |

## Design goals

In priority order:

1. **Zero-reflection performance** — fully typed request/response building, no
   `reflect` in the hot path; allocation-tested like `gotd/td`.
2. **Type-safe unions & enums** — `ChatID`, `InputFile`, `ChatMember`,
   `ReplyMarkup`, parse modes, etc. as sealed interfaces and typed constants,
   not stringly-typed structs.
3. **First-class context & structured errors** — context-first API; typed
   errors (flood-wait, retry-after, network vs API vs not-implemented);
   proactive rate limiting.
4. **A great handler framework** — composable middleware, router and predicates
   over a native MTProto update stream.

See [Architecture](./architecture.md) for how those map onto the `gotd/td`
building blocks.

## Installation

```bash
go get github.com/gotd/botapi@latest
```

You need two things to run a bot:

1. An **MTProto app identity** — `AppID` and `AppHash` from
   [my.telegram.org](https://my.telegram.org). These identify the *application*, not the bot, and
   are required even for bots. See
   [Obtaining API credentials](../getting-started/obtaining-api-credentials.md).
2. A **bot token** from [@BotFather](https://t.me/BotFather).

:::info[Relationship to gotd/td]
botapi is a higher-level library that sits on top of `gotd/td`. The rest of this
site documents the low-level MTProto client; this section documents the Bot API
surface built on it. The two share the same engine — the
[message sender](../helpers/message-sender.md),
[peer resolution](../helpers/peers.md),
[uploads/downloads](../helpers/uploading-files.md) and
[update recovery](../helpers/updates-recovery.md) all power botapi under the hood.
:::

## What's here

* **[Getting started](./getting-started.md)** — your first bot, `New` and `Run`.
* **[Sending messages](./sending-messages.md)** — targeting chats, send options,
  formatting and keyboards.
* **[Media & files](./media-and-files.md)** — `InputFile`, typed media sends,
  albums, `GetFile` and downloads.
* **[Receiving updates](./receiving-updates.md)** — handlers, predicates,
  middleware, groups and commands.
* **[Callback & inline queries](./callback-and-inline-queries.md)** — answering
  button taps and inline mode.
* **[Editing & chat management](./chat-management.md)** — edits, forwards,
  members, admin and stickers.
* **[Errors & resilience](./errors-and-resilience.md)** — typed errors,
  flood-wait and rate limiting.
* **[Persistence & pooling](./persistence-and-pooling.md)** — storage, running
  many bots, and the raw escape hatch.
* **[Architecture](./architecture.md)** — how the Bot API surface maps onto
  MTProto.

Runnable bots live in
[`examples/`](https://github.com/gotd/botapi/tree/main/examples)
(`echo`, `buttons`, `inline`, `media`, `advanced`).

:::note[Status]
botapi is under active reconstruction from a codegen-first project into a
hand-written library. The bulk of the Bot API surface — sending, media,
keyboards, handlers, commands, files, chat management, errors and pooling — is in
place. A few methods are still deferred; see the
[roadmap](https://github.com/gotd/botapi/blob/main/docs/roadmap.md) for the
current state.
:::
