# Build the browser demo and run it locally.
#
#   make        # compile main.wasm and copy the Go wasm_exec.js shim
#   make serve  # build, then serve on http://localhost:8080

WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js

.PHONY: all serve clean

all: main.wasm wasm_exec.js

main.wasm: main.go go.mod
	GOOS=js GOARCH=wasm go build -o main.wasm .

# Go 1.24+ ships wasm_exec.js under lib/wasm; older toolchains use misc/wasm.
wasm_exec.js:
	cp "$(WASM_EXEC)" ./wasm_exec.js

serve: all
	go run ./serve

clean:
	rm -f main.wasm wasm_exec.js
