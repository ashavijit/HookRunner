# HookRunner Makefile
# Optimized build with stripped symbols for smaller binaries

VERSION ?= $(shell cat VERSION 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Optimized ldflags: -s strips symbol table, -w strips DWARF debug info
LDFLAGS := -s -w \
	-X github.com/ashavijit/hookrunner/internal/version.Version=$(VERSION) \
	-X github.com/ashavijit/hookrunner/internal/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/ashavijit/hookrunner/internal/version.BuildDate=$(BUILD_DATE)

# Binary name
BINARY := hookrunner
ifeq ($(OS),Windows_NT)
	BINARY := hookrunner.exe
endif

.PHONY: all build build-dev install clean test lint release help

all: build

build:
	@echo "Building optimized binary..."
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/hookrunner
	@echo "Built $(BINARY) ($$(du -h $(BINARY) | cut -f1))"

build-dev:
	@echo "Building development binary..."
	go build -o $(BINARY) ./cmd/hookrunner
	@echo "Built $(BINARY) (dev mode)"

install:
	@echo "Installing hookrunner..."
	go install -ldflags "$(LDFLAGS)" ./cmd/hookrunner
	@echo "Installed to $$(go env GOPATH)/bin/hookrunner"

hooks:
	./$(BINARY) install

test:
	go test ./... -v -cover

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	golangci-lint run

clean:
	rm -f $(BINARY) hookrunner hookrunner.exe hookrunner_optimized.exe
	rm -f coverage.out coverage.html
	rm -rf dist/

release:
	@echo "Building release binaries..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/hookrunner-linux-amd64 ./cmd/hookrunner
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/hookrunner-linux-arm64 ./cmd/hookrunner
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/hookrunner-darwin-amd64 ./cmd/hookrunner
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/hookrunner-darwin-arm64 ./cmd/hookrunner
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/hookrunner-windows-amd64.exe ./cmd/hookrunner
	@echo "Release binaries:"
	@ls -lh dist/

size:
	@echo "=== Binary Size Comparison ==="
	@echo "Optimized build:"
	@go build -ldflags "-s -w" -o /tmp/hookrunner-opt ./cmd/hookrunner && du -h /tmp/hookrunner-opt
	@echo "Debug build:"
	@go build -o /tmp/hookrunner-dbg ./cmd/hookrunner && du -h /tmp/hookrunner-dbg
	@rm -f /tmp/hookrunner-opt /tmp/hookrunner-dbg

help:
	@echo "HookRunner Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build      Build optimized binary (default)"
	@echo "  build-dev  Build with debug info (faster compile)"
	@echo "  install    Install to GOPATH/bin"
	@echo "  hooks      Install git hooks"
	@echo "  test       Run tests"
	@echo "  coverage   Generate coverage report"
	@echo "  lint       Run golangci-lint"
	@echo "  clean      Remove build artifacts"
	@echo "  release    Cross-compile for all platforms"
	@echo "  size       Compare optimized vs debug binary sizes"
	@echo "  help       Show this help"
