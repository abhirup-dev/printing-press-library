.PHONY: build test lint install clean

BIN_EXT := $(if $(filter windows,$(shell go env GOOS)),.exe,)

build:
	go build -trimpath -o bin/unitypredict-pp-cli$(BIN_EXT) ./cmd/unitypredict-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/unitypredict-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -trimpath -o bin/unitypredict-pp-mcp$(BIN_EXT) ./cmd/unitypredict-pp-mcp

install-mcp:
	go install ./cmd/unitypredict-pp-mcp

build-all: build build-mcp
