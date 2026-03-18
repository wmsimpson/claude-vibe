# Vibe CLI Makefile

# Build variables
VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Paths
BIN_DIR := bin
BINARY := $(BIN_DIR)/vibe

# Zerobus credentials (set via environment variables for builds with telemetry)
ZEROBUS_CLIENT_ID ?=
ZEROBUS_CLIENT_SECRET ?=
ZEROBUS_ENDPOINT ?=
ZEROBUS_UNITY_CATALOG_URL ?=
ZEROBUS_TABLE_NAME ?=

# ldflags for version injection
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Add zerobus credentials if provided
ifneq ($(ZEROBUS_CLIENT_ID),)
LDFLAGS += -X github.com/will-simpson/individual-vibe-tool/cli/internal/telemetry.ZerobusClientID=$(ZEROBUS_CLIENT_ID)
endif
ifneq ($(ZEROBUS_CLIENT_SECRET),)
LDFLAGS += -X github.com/will-simpson/individual-vibe-tool/cli/internal/telemetry.ZerobusClientSecret=$(ZEROBUS_CLIENT_SECRET)
endif
ifneq ($(ZEROBUS_ENDPOINT),)
LDFLAGS += -X github.com/will-simpson/individual-vibe-tool/cli/internal/telemetry.ZerobusEndpoint=$(ZEROBUS_ENDPOINT)
endif
ifneq ($(ZEROBUS_UNITY_CATALOG_URL),)
LDFLAGS += -X github.com/will-simpson/individual-vibe-tool/cli/internal/telemetry.ZerobusUnityCatalogURL=$(ZEROBUS_UNITY_CATALOG_URL)
endif
ifneq ($(ZEROBUS_TABLE_NAME),)
LDFLAGS += -X github.com/will-simpson/individual-vibe-tool/cli/internal/telemetry.ZerobusTableName=$(ZEROBUS_TABLE_NAME)
endif

# Dictation helper (macOS only)
DICTATION_SRC := internal/voice/dictation/main.swift
DICTATION_BIN := $(BIN_DIR)/vibe-dictation

.PHONY: all build build-dictation test clean install lint fmt help

# Default target
all: build

# Build the Go binary
build:
	@echo "Building vibe $(VERSION) ($(COMMIT)) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/vibe

# Build the macOS dictation helper (requires macOS + Xcode CLI tools)
build-dictation:
	@echo "Building vibe-dictation helper..."
	@mkdir -p $(BIN_DIR)
	swiftc -O -o $(DICTATION_BIN) $(DICTATION_SRC) -framework Speech -framework AVFoundation
	@echo "Built $(DICTATION_BIN)"

# Build for all platforms
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64

build-darwin-amd64:
	@echo "Building for darwin/amd64..."
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/vibe-darwin-amd64 ./cmd/vibe

build-darwin-arm64:
	@echo "Building for darwin/arm64..."
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/vibe-darwin-arm64 ./cmd/vibe

build-linux-amd64:
	@echo "Building for linux/amd64..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/vibe-linux-amd64 ./cmd/vibe

build-linux-arm64:
	@echo "Building for linux/arm64..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/vibe-linux-arm64 ./cmd/vibe

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html

# Install to GOPATH/bin
install: build
	@echo "Installing vibe to $(GOPATH)/bin..."
	cp $(BINARY) $(GOPATH)/bin/vibe

# Install to /usr/local/bin (requires sudo)
install-global: build
	@echo "Installing vibe to /usr/local/bin..."
	sudo cp $(BINARY) /usr/local/bin/vibe

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run the binary
run: build
	./$(BINARY)

# Show help
help:
	@echo "Vibe CLI Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build            Build the Go binary"
	@echo "  make build-dictation  Build the macOS voice dictation helper"
	@echo "  make build-all        Build for all platforms"
	@echo "  make test             Run tests"
	@echo "  make test-coverage    Run tests with coverage report"
	@echo "  make clean            Remove build artifacts"
	@echo "  make install          Install to GOPATH/bin"
	@echo "  make install-global   Install to /usr/local/bin (requires sudo)"
	@echo "  make lint             Run linter"
	@echo "  make fmt              Format code"
	@echo "  make run              Build and run"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  COMMIT=$(COMMIT)"
	@echo "  DATE=$(DATE)"
	@echo ""
	@echo "Telemetry (Zerobus) Configuration:"
	@echo "  To build with telemetry support, set these environment variables:"
	@echo "    ZEROBUS_CLIENT_ID         Service principal client ID"
	@echo "    ZEROBUS_CLIENT_SECRET     Service principal secret"
	@echo "    ZEROBUS_ENDPOINT          Zerobus gRPC endpoint (e.g., https://zerobus.databricks.com)"
	@echo "    ZEROBUS_UNITY_CATALOG_URL Workspace URL for OAuth (e.g., https://workspace.databricks.com)"
	@echo "    ZEROBUS_TABLE_NAME        Target Delta table (e.g., catalog.schema.table)"
	@echo ""
	@echo "  Example:"
	@echo "    ZEROBUS_CLIENT_ID=xxx ZEROBUS_CLIENT_SECRET=yyy \\"
	@echo "    ZEROBUS_ENDPOINT=https://zerobus.databricks.com \\"
	@echo "    ZEROBUS_UNITY_CATALOG_URL=https://workspace.databricks.com \\"
	@echo "    ZEROBUS_TABLE_NAME=main.telemetry.events \\"
	@echo "    make build"
