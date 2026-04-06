BINARY   := dlpbuster
BUILD_DIR := bin
CMD_DIR   := ./cmd/dlpbuster

GO        := go
GOFLAGS   := -trimpath
LDFLAGS   := -ldflags "-s -w -X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)"

.PHONY: all build test lint clean release fmt vet

all: build

## build: compile binary to ./bin/
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD_DIR)

## test: run tests with race detector
test:
	$(GO) test -race -count=1 -timeout 120s ./...

## test-cover: run tests with coverage
test-cover:
	$(GO) test -race -count=1 -timeout 120s -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## fmt: run gofmt + goimports
fmt:
	gofmt -w .
	goimports -w . 2>/dev/null || true

## vet: run go vet
vet:
	$(GO) vet ./...

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

## release: goreleaser snapshot build (no publish)
release:
	goreleaser release --snapshot --clean

## install: install binary to GOPATH/bin
install:
	$(GO) install $(GOFLAGS) $(LDFLAGS) $(CMD_DIR)

## help: show this message
help:
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
