---
sidebar_position: 1
---

# Transports and proxies

How gotd reaches Telegram's datacenters is controlled by `Options.Resolver`. The default
connects directly with the standard MTProto transport, but you can swap it for a proxy
or a different fingerprint.

## MTProxy

To route through an [MTProxy](https://core.telegram.org/mtproto/mtproxy), build a
resolver with the proxy address and secret:

```go
import "github.com/gotd/td/telegram/dcs"

resolver, err := dcs.MTProxy(addr, secret, dcs.MTProxyOptions{})
if err != nil {
	return err
}

client := telegram.NewClient(appID, appHash, telegram.Options{
	Resolver: resolver,
})
```

`secret` is the proxy's secret as raw bytes; FakeTLS secrets are supported. You can
verify connectivity before logging in with an unauthenticated call like
`api.HelpGetNearestDC(ctx)`, as the
[`mtproxy-connect`](https://github.com/gotd/td/tree/main/examples/mtproxy-connect)
example does.

## Mimicking Telegram Desktop

Some setups want traffic that looks like the official Desktop client. gotd ships presets
for the resolver (Obfuscated2 + abridged transport) and the device info:

```go
client := telegram.NewClient(appID, appHash, telegram.Options{
	Resolver: telegram.TDesktopResolver(),
	Device:   telegram.DeviceTDesktopWindows(),
})
```

See the
[`tdesktop-mimic`](https://github.com/gotd/td/tree/main/examples/tdesktop-mimic)
example.

## SOCKS and HTTP proxies

For ordinary proxies, plug a custom dialer into a plain resolver via
`dcs.Plain(dcs.PlainOptions{Dial: ...})`. Any
`golang.org/x/net/proxy` dialer (SOCKS5, etc.) works here.

## WebSocket transport

gotd supports a WebSocket transport (`dcs.Websocket`), which is what lets it run in
**WASM** in the browser — browsers cannot open raw TCP sockets. On the `js/wasm`
platform it is selected automatically, so the rest of your code is unchanged. See
[Running in the browser (WASM)](./wasm-websocket.mdx) for a runnable example.

## Test servers

For experiments against Telegram's test datacenters, use the bundled test credentials
and DC list:

```go
client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
	DC:     2,
	DCList: dcs.Test(),
})
```
