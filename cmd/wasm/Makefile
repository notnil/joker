# Build the browser client. Requires Go 1.26 (js/wasm).
GOROOT := $(shell go env GOROOT)
WASM_EXEC := $(GOROOT)/lib/wasm/wasm_exec.js
ifeq ($(wildcard $(WASM_EXEC)),)
WASM_EXEC := $(GOROOT)/misc/wasm/wasm_exec.js
endif

.PHONY: build
build:
	GOOS=js GOARCH=wasm go build -o ../server/public/main.wasm .
	cp "$(WASM_EXEC)" ../server/public/wasm_exec.js
