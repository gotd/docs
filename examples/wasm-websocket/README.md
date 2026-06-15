# gotd WASM WebSocket example

A runnable browser demo: gotd compiled to WebAssembly, connecting to Telegram
over the **WebSocket transport**. There is no TCP in the browser, so on the
`js/wasm` platform gotd defaults to `dcs.Websocket` automatically — the client
code is identical to a native program.

The page asks for an App ID / App Hash (from
[my.telegram.org/apps](https://my.telegram.org/apps)) and calls
`help.getNearestDC`, which needs no authentication.

## Run

```sh
make serve     # builds main.wasm + wasm_exec.js, then serves on :8080
```

Open <http://localhost:8080> and click **Connect**.

Or build and serve in separate steps:

```sh
make           # produces main.wasm and copies wasm_exec.js
go run ./serve # static file server on http://localhost:8080
```

## Files

| File          | Purpose                                                        |
| ------------- | ------------------------------------------------------------- |
| `main.go`     | The `js/wasm` program; exports `gotdConnect` to JavaScript.    |
| `index.html`  | Host page that loads `wasm_exec.js` and `main.wasm`.          |
| `serve/`      | Tiny static file server so the demo runs with the Go toolchain alone. |
| `Makefile`    | Builds the WASM binary and copies Go's `wasm_exec.js` shim.    |
