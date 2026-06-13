---
sidebar_position: 3
---

# Debugging and tracing

When a request misbehaves, it helps to see exactly what goes over the wire. gotd gives
you several levels of visibility.

## Logging

Pass a [`github.com/gotd/log`](https://github.com/gotd/log)-compatible logger via
`Options.Logger`. The examples use the zap adapter:

```go
import "github.com/gotd/log/logzap"

telegram.Options{Logger: logzap.New(zapLogger)}
```

At debug level this reports connection state, reconnects, and the RPCs being sent.

## Pretty-printing requests and responses

For a focused view of just the API calls, add a logging
[middleware](../helpers/middleware.md). The
[`pretty-print`](https://github.com/gotd/td/tree/main/examples/pretty-print) example
prints each request, the response and the elapsed time:

```go
func prettyMiddleware() telegram.MiddlewareFunc {
	return func(next tg.Invoker) telegram.InvokeFunc {
		return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			fmt.Println("→", formatObject(input))
			start := time.Now()
			if err := next.Invoke(ctx, input, output); err != nil {
				fmt.Println("←", err)
				return err
			}
			fmt.Printf("← (%s) %s\n", time.Since(start).Round(time.Millisecond), formatObject(output))
			return nil
		}
	}
}
```

It formats objects with the
[`tdp`](https://ref.gotd.dev/pkg/github.com/gotd/td/tdp.html) package, which renders any
generated type in a readable, indented form — useful well beyond this middleware.

## Inspecting updates

To see *every* update (not just the ones your dispatcher handles), set a custom
`UpdateHandler` that logs the raw `tg.UpdatesClass` before delegating. The
`pretty-print` example also demonstrates this.

## Error details

RPC errors are `*tgerr.Error` and carry the Telegram error code and message. Use
`tgerr.Is(err, "CODE")` to match, and print the error directly for the full type and
argument (for example `FLOOD_WAIT` carries the wait seconds).

## OpenTelemetry

For production observability, the
[`oteltg`](https://ref.gotd.dev/pkg/github.com/gotd/td/oteltg.html) package provides
OpenTelemetry metrics and traces for the client, so MTProto calls show up in your
tracing backend.
