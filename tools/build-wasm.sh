#!/usr/bin/env bash
# Build the WASM WebSocket demo into static/wasm/ so the embedded "Run" widget
# on /docs/advanced/wasm-websocket has something to load.
#
# Invoked automatically by the prebuild/prestart npm scripts. Requires the Go
# toolchain; if Go is absent it warns and skips (the rest of the site still
# builds, but the live demo will not load).
set -euo pipefail

cd "$(dirname "$0")/.."
root="$(pwd)"
src="$root/examples/wasm-websocket"
out="$root/static/wasm"

if ! command -v go >/dev/null 2>&1; then
  echo "build-wasm: Go toolchain not found — skipping WASM demo build." >&2
  echo "build-wasm: the embedded demo on /docs/advanced/wasm-websocket will not load." >&2
  exit 0
fi

mkdir -p "$out"

echo "build-wasm: compiling $src -> $out/main.wasm"
( cd "$src" && GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o "$out/main.wasm" . )

# Go 1.24+ ships wasm_exec.js under lib/wasm; older toolchains use misc/wasm.
goroot="$(go env GOROOT)"
wasm_exec="$goroot/lib/wasm/wasm_exec.js"
[ -f "$wasm_exec" ] || wasm_exec="$goroot/misc/wasm/wasm_exec.js"
cp "$wasm_exec" "$out/wasm_exec.js"

echo "build-wasm: done ($(du -h "$out/main.wasm" | cut -f1) main.wasm)"
