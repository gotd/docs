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

### Long-running programs

The `Waiter` keys flood waits by the method, unwrapping `invokeWith*` (so a flood
wait inside a [takeout session](../advanced/data-export.md) throttles only that
method), and dispatches requests concurrently. Further options:

```go
waiter := floodwait.NewWaiter().
	// Space out requests of a method after it hits FLOOD_WAIT: the interval
	// doubles on each flood wait and decays on success.
	WithAdaptivePacing(100*time.Millisecond, 10*time.Second).
	// Requests with their own pacing bypass the waiter.
	WithExempt(tg.UploadGetFileRequestTypeID).
	// Retry transient 5xx errors with exponential backoff.
	WithServerErrorRetries(3).
	WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		log.Printf("%s: wait %s", wait.Name, wait.Duration)
	})

// Per-request policy: background jobs wait as long as it takes,
// interactive ones fail fast.
bg := floodwait.WithPolicy(ctx, floodwait.Policy{MaxWait: -1, MaxRetries: -1})
ui := floodwait.WithPolicy(ctx, floodwait.Policy{NoRetry: true})

// Pacing state for metrics.
for _, s := range waiter.Snapshot() {
	fmt.Println(s.Name, s.Wait, s.Interval)
}
```

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
