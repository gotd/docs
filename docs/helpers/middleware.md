---
sidebar_position: 8
---

# Middleware

Middleware wraps every RPC the client makes, letting you retry, throttle, log or trace
requests without touching call sites. You install middleware via `Options.Middlewares`,
and they apply in order.

```go
type Middleware interface {
	Handle(next tg.Invoker) InvokeFunc
}
```

A middleware receives the next invoker and returns a function that calls it — the
classic onion pattern. The most useful implementations live in
[`gotd/contrib`](https://github.com/gotd/contrib).

## FLOOD_WAIT handling

Telegram rate-limits with `FLOOD_WAIT` errors that say "retry after N seconds". The
[`floodwait`](https://github.com/gotd/contrib/tree/master/middleware/floodwait)
middleware catches these and retries automatically so your code never sees them.

`Waiter` is the scheduler-based implementation for long-running clients. It both acts as
a middleware **and** must wrap your run loop so it can schedule retries:

```go
import "github.com/gotd/contrib/middleware/floodwait"

waiter := floodwait.NewWaiter().
	WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		log.Printf("FLOOD_WAIT, retrying in %s", wait.Duration)
	})

client := telegram.NewClient(appID, appHash, telegram.Options{
	Middlewares: []telegram.Middleware{waiter},
})

// Run the client *inside* the waiter so retries can be scheduled.
return waiter.Run(ctx, func(ctx context.Context) error {
	return client.Run(ctx, func(ctx context.Context) error {
		// ... your work ...
		return nil
	})
})
```

For one-shot scripts, `floodwait.NewSimpleWaiter()` is a simpler timer-based variant
that needs no `Run` wrapper. Both support `WithMaxRetries` and `WithMaxWait`.

## Rate limiting

To stay *under* the limits in the first place, add the
[`ratelimit`](https://github.com/gotd/contrib/tree/master/middleware/ratelimit)
middleware, which paces outgoing requests with a token bucket
(`golang.org/x/time/rate`):

```go
import (
	"github.com/gotd/contrib/middleware/ratelimit"
	"golang.org/x/time/rate"
)

telegram.Options{
	Middlewares: []telegram.Middleware{
		waiter,                                      // retry on FLOOD_WAIT
		ratelimit.New(rate.Every(time.Millisecond*100), 5), // ~10 req/s, burst 5
	},
}
```

Order matters: putting the waiter first means it wraps (and retries through) the rate
limiter. This pairing — proactive rate limiting plus reactive flood-wait retries — is
the recommended setup for userbots, and is exactly what the
[`userbot`](https://github.com/gotd/td/tree/main/examples/userbot) example uses.

## Writing your own

A middleware is just a function. The
[`pretty-print`](https://github.com/gotd/td/tree/main/examples/pretty-print) example
logs every request and response with timing:

```go
func logging() telegram.MiddlewareFunc {
	return func(next tg.Invoker) telegram.InvokeFunc {
		return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			start := time.Now()
			err := next.Invoke(ctx, input, output)
			log.Printf("%T in %s (err=%v)", input, time.Since(start), err)
			return err
		}
	}
}
```

See [Debugging and tracing](../advanced/debugging.md) for more on inspecting traffic.
