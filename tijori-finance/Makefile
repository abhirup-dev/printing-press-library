.PHONY: build test lint install clean

BIN_EXT := $(if $(filter windows,$(shell go env GOOS)),.exe,)

build:
	go build -trimpath -o bin/tijori-finance-pp-cli$(BIN_EXT) ./cmd/tijori-finance-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/tijori-finance-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -trimpath -o bin/tijori-finance-pp-mcp$(BIN_EXT) ./cmd/tijori-finance-pp-mcp

install-mcp:
	go install ./cmd/tijori-finance-pp-mcp

build-all: build build-mcp
