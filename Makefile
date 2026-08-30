GO ?= go
WASM ?= river-raid.wasm

.PHONY: all fmt-check lint test wasm clean

all: fmt-check lint test wasm

fmt-check:
	$(GO) tool golangci-lint fmt --diff

lint:
	$(GO) tool golangci-lint run

test:
	$(GO) test ./...

wasm:
	GOOS=js GOARCH=wasm $(GO) build -o $(WASM) .

clean:
	rm -f $(WASM)
