.PHONY: build test clean install lint fmt vet

# Binary name
BINARY_NAME=mdev
PACKAGE=github.com/sombi/mobile-dev-helper

# Build directories
BUILD_DIR=build
DIST_DIR=dist

# Version info (can be overridden at build time)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Go build flags
LDFLAGS=-ldflags "-X github.com/sombi/mobile-dev-helper/internal/version.Version=$(VERSION) -X github.com/sombi/mobile-dev-helper/internal/version.Commit=$(COMMIT) -X github.com/sombi/mobile-dev-helper/internal/version.Date=$(DATE)"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html

# Install locally
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME) 2>/dev/null || cp $(BUILD_DIR)/$(BINARY_NAME) ~/go/bin/$(BINARY_NAME) 2>/dev/null || echo "Please manually copy $(BUILD_DIR)/$(BINARY_NAME) to your PATH"

# Run the binary
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Format code
fmt:
	@echo "Formatting..."
	go fmt ./...

# Run linter
lint:
	@echo "Linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/"; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Check for issues (fmt, vet, lint)
check: fmt vet lint test

# Release using goreleaser (requires GITHUB_TOKEN)
release:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "goreleaser not installed. Install from https://goreleaser.com/install/"; \
	fi

# Snapshot release (local, no publishing)
release-snapshot:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed. Install from https://goreleaser.com/install/"; \
	fi

# Development helpers
dev-build: build

dev-test: test

dev-run: run
