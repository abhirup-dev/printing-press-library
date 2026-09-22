.PHONY: build test lint install clean

BIN_EXT := $(if $(filter windows,$(shell go env GOOS)),.exe,)

build:
	go build -trimpath -o bin/tickertape-pp-cli$(BIN_EXT) ./cmd/tickertape-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/tickertape-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -trimpath -o bin/tickertape-pp-mcp$(BIN_EXT) ./cmd/tickertape-pp-mcp

install-mcp:
	go install ./cmd/tickertape-pp-mcp

build-all: build build-mcp
