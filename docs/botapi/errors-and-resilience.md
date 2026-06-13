---
sidebar_position: 8
---

# Errors & resilience

## Structured errors

Methods return errors shaped like the HTTP Bot API: an `*Error` with a `Code`
and a `Description`. Branch on it with `errors.As`, or use the helpers:

```go
if _, err := bot.SendMessage(ctx, chat, text); err != nil {
	if wait, ok := botapi.AsFloodWait(err); ok {
		time.Sleep(wait)
	} else if newID, ok := botapi.AsChatMigrated(err); ok {
		_ = newID // retry against newID (group upgraded to supergroup)
	} else if botapi.Code(err) == 403 {
		// blocked, or the bot is not a member of the chat
	}
}
```

| Helper | Returns |
| --- | --- |
| `Code(err)` | The Bot-API error code (`400`/`403`/`429`/…). |
| `AsFloodWait(err)` | `(retryAfter time.Duration, ok bool)`. |
| `AsChatMigrated(err)` | `(newChatID int64, ok bool)` when a group upgraded to a supergroup. |

The descriptions mirror the official Bot API verbatim — botapi maps the
underlying MTProto error (e.g. `FLOOD_WAIT_3`, `PEER_ID_INVALID`,
`USER_IS_BLOCKED`) onto the same `{error_code, description}` the HTTP server
would return.

:::tip[Context cancellation passes through]
`context.Canceled` and `context.DeadlineExceeded` are returned unchanged — even
when the failure was wrapped in an RPC error — so `errors.Is(err,
context.Canceled)` works as you'd expect.
:::

## Flood-wait & rate limiting

Both are opt-in via `Options`, and off by default:

```go
botapi.Options{
	AppID: appID, AppHash: appHash,
	FloodWait:         true, // retry FLOOD_WAIT-limited requests transparently
	RequestsPerSecond: 25,   // proactive global token-bucket limit
}
```

* **`FloodWait`** waits out a `FLOOD_WAIT` limit and retries the request instead
  of returning a `429`. Bound the retries with `MaxFloodWaitRetries`.
* **`RequestsPerSecond`** (with `RequestBurst`) caps outgoing MTProto requests
  with a global token bucket — a coarse guard against hitting Telegram's limits
  before they bite.

These are wired as client invoker middlewares, so they apply uniformly across
every method. Reconnection and update-gap recovery are handled transparently by
the underlying [`gotd/td`](../helpers/updates-recovery.md) client.

:::info[MTProto limits, not Bot API limits]
Because botapi talks MTProto directly, it is subject to MTProto's flood limits
rather than the public Bot API server's rate limits — generally more generous,
but the same `FloodWait`/rate-limit tools apply.
:::
