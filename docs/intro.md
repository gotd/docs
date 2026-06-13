---
sidebar_position: 1
---

# Introduction

[**gotd**](https://github.com/gotd/td) is a Telegram MTProto API client in Go for
**users and bots**. It speaks the same low-level protocol as the official apps and
[TDLib](https://core.telegram.org/tdlib), giving you direct access to every method of
the [Telegram API](https://core.telegram.org/schema) — not just the Bot API.

```go
package main

import (
	"context"

	"github.com/gotd/td/telegram"
)

func main() {
	// Grab these from https://my.telegram.org/apps.
	client := telegram.NewClient(appID, appHash, telegram.Options{})
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// It is only valid to use client while this function has not returned
		// and ctx is not cancelled.
		api := client.API()

		// Now you can invoke MTProto RPC requests by calling the API.
		_ = api

		// Return to close client connection and free up resources.
		return nil
	}); err != nil {
		panic(err)
	}
	// Client is closed.
}
```

## Why gotd

* **Full MTProto 2.0** implementation in pure Go — call any method via
  [`client.API()`](./basics/calling-the-api.md).
* **Highly optimized**: low memory (≈150 KB per idle client) and CPU overhead; can
  handle thousands of concurrent clients.
* **Generated types** for the whole Telegram schema, with embedded official
  documentation and links.
* **Helpers** that hide the sharp edges of the raw API: a
  [message sender](./helpers/message-sender.md),
  [uploads](./helpers/uploading-files.md),
  [downloads](./helpers/downloading-files.md),
  [pagination iterators](./helpers/query-iterators.md),
  [peer resolution](./helpers/peers.md) and an
  [update-recovery engine](./helpers/updates-recovery.md).
* **Robust**: automatic reconnects with keepalive, datacenter migration, request
  cancellation via context, and middleware for
  [rate limiting and FLOOD_WAIT handling](./helpers/middleware.md).
* **Secure**: conforms to Telegram's
  [security guidelines](https://core.telegram.org/mtproto/security_guidelines),
  with secure PRNG, replay-attack protection, 2FA and MTProxy support.

:::warning[Read this first]
Before using this library on a real account, read the
[**How To Not Get Banned**](https://github.com/gotd/td/blob/main/.github/SUPPORT.md#how-to-not-get-banned)
guide. Telegram may limit or ban accounts that behave abusively.
:::

## Installation

```bash
go get github.com/gotd/td@latest
```

gotd requires a reasonably recent Go version. The only mandatory inputs are your
application's `api_id` and `api_hash` — see
[Obtaining API credentials](./getting-started/obtaining-api-credentials.md).

## How the docs are organized

* **[Getting started](./getting-started/obtaining-api-credentials.md)** — credentials,
  your first client, and the boilerplate the examples use.
* **[Authentication](./authentication/user.md)** — signing in as a user or bot, QR
  login, 2FA, and persisting sessions.
* **[Basics](./basics/calling-the-api.md)** — calling the raw API, sending messages,
  handling updates, and a complete echo bot.
* **[Helpers](./helpers/message-sender.md)** — the high-level packages that make the
  API pleasant to use.
* **[Advanced](./advanced/transports-and-proxy.md)** — transports and proxies, running
  without `Run`, debugging, data export, and voice calls.

:::tip[Generated API reference]
Because of `pkg.go.dev` limitations, the generated `tg` package docs are hosted
separately at [**ref.gotd.dev**](https://ref.gotd.dev/pkg/github.com/gotd/td/tg.html).
This site covers the hand-written, higher-level parts of the library.
:::

## Looking for something higher level?

gotd is intentionally low level. If you want a more opinionated wrapper, the community
maintains [GoTGProto](https://github.com/celestix/gotgproto), which adds session
strings, peer storage and convenience helpers on top of gotd.
