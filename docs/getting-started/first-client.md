---
sidebar_position: 2
---

# Your first client

A gotd program almost always follows the same shape:

1. Create a client with [`telegram.NewClient`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram.html).
2. Call `client.Run`, passing a callback.
3. Do all your work **inside** that callback.
4. Return from the callback to disconnect cleanly.

```go
package main

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
)

func main() {
	client := telegram.NewClient(appID, appHash, telegram.Options{})

	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// The client is connected only while this function runs.
		// ctx is cancelled when the connection drops or Run returns.
		api := client.API()

		// help.getNearestDC works without authentication — a good connectivity check.
		dc, err := api.HelpGetNearestDC(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("connected, nearest DC: %d\n", dc.NearestDC)
		return nil
	}); err != nil {
		panic(err)
	}
}
```

## The `Run` lifecycle

`Run` connects, performs the MTProto handshake, runs your callback, and then tears the
connection down. A few rules follow from this:

* **The client is only usable inside the callback.** `client.API()`, `client.Auth()`,
  `client.Self()` and friends must be called while `Run` has not returned.
* **`ctx` governs the connection.** If the context is cancelled, the callback's `ctx`
  is cancelled too. Use it for every API call so requests are cancelled gracefully.
* **Returning ends the session.** Return `nil` for a clean shutdown, or an error to
  propagate it out of `Run`.

If you authenticate inside the callback, the session lives only for that run unless you
persist it — see [Sessions and storage](../authentication/sessions.md). Without a stored
session you must re-authenticate every time.

## What about long-running bots?

A bot or userbot needs to stay connected and process updates indefinitely. For that you
block inside the callback until the context is cancelled. gotd ships a helper,
[`telegram.RunUntilCanceled`](../basics/handling-updates.md), and the
[`BotFromEnvironment`](../authentication/bot.md) wrapper that wires this up for you. See
[Handling updates](../basics/handling-updates.md) and the
[Echo bot tutorial](../basics/echo-bot.md).

## Configuring the client

[`telegram.Options`](https://ref.gotd.dev/pkg/github.com/gotd/td/telegram.html#Options)
controls almost everything. The fields you will reach for most often:

| Field             | Purpose                                                            |
|-------------------|--------------------------------------------------------------------|
| `SessionStorage`  | Persist auth between runs — see [Sessions](../authentication/sessions.md) |
| `UpdateHandler`   | Receive incoming updates — see [Handling updates](../basics/handling-updates.md) |
| `Logger`          | A `github.com/gotd/log`-compatible logger (e.g. via `logzap`)      |
| `Middlewares`     | Intercept every RPC — see [Middleware](../helpers/middleware.md)   |
| `NoUpdates`       | Disable update subscription for one-shot scripts                   |
| `Device`          | Device/app info reported to Telegram                               |
| `Resolver`        | Customize how datacenters are reached — see [Transports](../advanced/transports-and-proxy.md) |

All fields have sensible defaults; an empty `telegram.Options{}` is valid.
