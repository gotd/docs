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

## Wire traces

To see the MTProto messages themselves, after decryption — containers, acks, pings,
salts, wrappers like `invokeWithLayer` — set `Options.WireTracer` to an implementation of
[`mtproto.WireTracer`](https://ref.gotd.dev/pkg/github.com/gotd/td/mtproto.html#WireTracer):

```go
type tracer struct{}

func (tracer) OnMessage(m mtproto.WireMessage, body []byte) {
	// m.Conn.DC, m.Direction, m.MessageID, m.SeqNo; body is only valid during the call.
}

func (tracer) OnEvent(e mtproto.WireEvent) {
	// e.Type: conn_start, conn_up, key_set, session_new, salt_change, request, ...
}

client := telegram.NewClient(appID, appHash, telegram.Options{WireTracer: tracer{}})
```

The tracer is called for every connection of the client (primary, other DCs, media and
CDN), synchronously, so it must return quickly. It never gets key material, only auth
key ids.

The [`wiretrace`](https://ref.gotd.dev/pkg/github.com/gotd/td/wiretrace.html) package
reads and writes wire traces in the format of the traced Telegram Desktop build, and the
`cmd/wiretrace` tool analyzes them: `decode` (one JSON line per logical message),
`timeline`, `stats` (per-method counts, latencies, errors, flood waits) and `diff`, which
compares two runs, e.g. Telegram Desktop against a gotd client, phase by phase or step by
step with `--align marker`. With UI events in a Telegram Desktop trace, `timeline --ui`
shows which click or key press sent which requests.

```bash
go run github.com/gotd/td/cmd/wiretrace@latest stats RUN_DIR
```

## Error details

RPC errors are `*tgerr.Error` and carry the Telegram error code and message. Use
`tgerr.Is(err, "CODE")` to match, and print the error directly for the full type and
argument (for example `FLOOD_WAIT` carries the wait seconds).

## OpenTelemetry

For production observability, the
[`oteltg`](https://ref.gotd.dev/pkg/github.com/gotd/td/oteltg.html) package provides
OpenTelemetry metrics and traces for the client, so MTProto calls show up in your
tracing backend.
