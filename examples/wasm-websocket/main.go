//go:build js && wasm

// Command wasm-websocket is a browser demo for gotd.
//
// It compiles to WebAssembly and connects to Telegram from inside the browser
// over the WebSocket transport, then reports the nearest datacenter. On the
// js/wasm platform gotd selects dcs.Websocket as the default resolver, so the
// usual client code works unchanged — there is no TCP in the browser.
//
// The program exports gotdConnect(appID, appHash[, onLog]) to JavaScript. It
// returns a Promise resolving to a short report, and streams structured log
// records to the optional onLog callback so a host page can show them live.
package main

import (
	"context"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/gotd/log"

	"github.com/gotd/td/telegram"
)

func main() {
	// Expose connect to JavaScript.
	js.Global().Set("gotdConnect", js.FuncOf(connect))

	// main must not return: that would tear down the Go runtime and make the
	// exported function unusable. Block forever instead.
	select {}
}

// connect bridges the JS call into Go. Browser APIs forbid blocking the main
// thread, so the actual network work happens in a goroutine and the result is
// delivered through a Promise.
func connect(_ js.Value, args []js.Value) any {
	if len(args) < 2 {
		return reject("usage: gotdConnect(appID, appHash[, onLog])")
	}

	appID, err := strconv.Atoi(args[0].String())
	if err != nil {
		return reject("invalid app ID: " + err.Error())
	}
	appHash := args[1].String()

	var onLog js.Value
	if len(args) >= 3 && args[2].Type() == js.TypeFunction {
		onLog = args[2]
	}

	handler := js.FuncOf(func(_ js.Value, promise []js.Value) any {
		resolve, rejectFn := promise[0], promise[1]
		go func() {
			report, err := run(appID, appHash, onLog)
			if err != nil {
				rejectFn.Invoke(err.Error())
				return
			}
			resolve.Invoke(report)
		}()
		return nil
	})
	return js.Global().Get("Promise").New(handler)
}

func run(appID int, appHash string, onLog js.Value) (string, error) {
	logger := log.Logger(log.Nop)
	if onLog.Truthy() {
		logger = jsLogger{fn: onLog, min: log.LevelInfo}
	}

	// No Resolver is set: on js/wasm telegram.NewClient defaults to the
	// WebSocket transport (wss://*.web.telegram.org/apiws).
	logger.Log(context.Background(), log.LevelInfo, "Building client over WebSocket transport")
	client := telegram.NewClient(appID, appHash, telegram.Options{
		Logger: logger,
	})

	var report string
	err := client.Run(context.Background(), func(ctx context.Context) error {
		logger.Log(ctx, log.LevelInfo, "Connected, requesting nearest DC")
		// help.getNearestDC needs no authentication — a clean connectivity check.
		dc, err := client.API().HelpGetNearestDC(ctx)
		if err != nil {
			return err
		}
		report = "nearest DC: " + strconv.Itoa(dc.NearestDC) + ", country: " + dc.Country
		logger.Log(ctx, log.LevelInfo, report)
		return nil
	})
	return report, err
}

// jsLogger implements log.Logger by rendering each record to one line and
// handing it to a JavaScript callback.
type jsLogger struct {
	fn  js.Value
	min log.Level
}

func (l jsLogger) Enabled(_ context.Context, level log.Level) bool {
	return level >= l.min
}

func (l jsLogger) Log(_ context.Context, level log.Level, msg string, attrs ...log.Attr) {
	var b strings.Builder
	b.WriteString(level.String())
	b.WriteByte('\t')
	b.WriteString(msg)
	for _, a := range attrs {
		b.WriteByte(' ')
		b.WriteString(a.Key)
		b.WriteByte('=')
		b.WriteString(a.Value.String())
	}
	l.fn.Invoke(b.String())
}

// reject returns an already-rejected Promise carrying msg.
func reject(msg string) js.Value {
	return js.Global().Get("Promise").Call("reject", msg)
}
