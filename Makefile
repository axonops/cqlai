# Makefile for cqlai - Cassandra CQL Shell

# Variables
BINARY_NAME := cqlai
MAIN_PATH := cmd/cqlai/main.go
BUILD_DIR := bin
INSTALL_DIR := /usr/local/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"
GOFLAGS := -v

# Platform detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
	PLATFORM := linux
endif
ifeq ($(UNAME_S),Darwin)
	PLATFORM := darwin
endif
ifeq ($(UNAME_S),Windows_NT)
	PLATFORM := windows
	BINARY_NAME := $(BINARY_NAME).exe
endif

ifeq ($(UNAME_M),x86_64)
	ARCH := amd64
endif
ifeq ($(UNAME_M),arm64)
	ARCH := arm64
endif
ifeq ($(UNAME_M),aarch64)
	ARCH := arm64
endif

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := build

# Phony targets
.PHONY: all build clean install uninstall run test lint fmt deps vendor grammar help release-all licenses

## all: Clean, format, lint, test, and build
all: clean fmt lint test build

## build: Build the binary for current platform
build:
	@echo "$(BLUE)Building $(BINARY_NAME) for $(PLATFORM)/$(ARCH)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-dev: Build with race detector and debug symbols (development build)
build-dev:
	@echo "$(BLUE)Building development version with race detector...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -race -gcflags="all=-N -l" $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Development build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## clean: Remove build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

## install: Install the binary to system path
install: build
	@echo "$(BLUE)Installing $(BINARY_NAME) to $(INSTALL_DIR)...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Installation complete$(NC)"
	@echo "$(GREEN)Run 'cqlai' to start the application$(NC)"

## uninstall: Remove the binary from system path
uninstall:
	@echo "$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)"
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Uninstallation complete$(NC)"

## run: Build and run the application
run: build
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME)

## test: Run all tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests complete$(NC)"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "$(BLUE)Generating coverage report...$(NC)"
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

## benchmark: Run benchmarks
benchmark:
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...

## lint: Run linters
lint:
	@echo "$(BLUE)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not found, using go vet$(NC)"; \
		go vet ./...; \
	fi
	@echo "$(GREEN)✓ Linting complete$(NC)"

## fmt: Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi
	@echo "$(GREEN)✓ Formatting complete$(NC)"

## deps: Download and tidy dependencies
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## vendor: Create vendor directory with all dependencies
vendor: deps
	@echo "$(BLUE)Creating vendor directory...$(NC)"
	@go mod vendor
	@echo "$(GREEN)✓ Vendor directory created$(NC)"

## grammar: Regenerate ANTLR grammar files
grammar:
	@echo "$(BLUE)Regenerating ANTLR grammar files...$(NC)"
	@if command -v antlr4 >/dev/null 2>&1; then \
		cd internal/parser/grammar && \
		antlr4 -Dlanguage=Go -package grammar CqlLexer.g4 && \
		antlr4 -Dlanguage=Go -package grammar -visitor CqlParser.g4; \
		echo "$(GREEN)✓ Grammar files regenerated$(NC)"; \
	else \
		echo "$(RED)✗ antlr4 not found. Install with: go install github.com/antlr4-go/antlr/v4/cmd/antlr4@latest$(NC)"; \
		exit 1; \
	fi

## licenses: Generate third-party license attributions
licenses:
	@echo "$(BLUE)Generating third-party license attributions...$(NC)"
	@if ! command -v go-licenses >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing go-licenses...$(NC)"; \
		go install github.com/google/go-licenses@latest; \
	fi
	@echo "$(BLUE)Collecting license files...$(NC)"
	@rm -rf THIRD-PARTY-LICENSES
	@mkdir -p THIRD-PARTY-LICENSES
	@PATH="$(HOME)/go/bin:$$PATH" go-licenses save ./cmd/cqlai --save_path=THIRD-PARTY-LICENSES --force || true
	@echo "$(BLUE)Generating license summary...$(NC)"
	@PATH="$(HOME)/go/bin:$$PATH" go-licenses report ./cmd/cqlai > THIRD-PARTY-LICENSES/NOTICES.txt 2>/dev/null || true
	@echo "$(GREEN)✓ License attributions generated in THIRD-PARTY-LICENSES/$(NC)"
	@echo "$(GREEN)  - Individual license files in subdirectories$(NC)"
	@echo "$(GREEN)  - License summary in NOTICES.txt$(NC)"

## release: Build release binaries for all platforms
release-all: clean
	@echo "$(BLUE)Building release binaries...$(NC)"
	@mkdir -p $(BUILD_DIR)/releases
	
	# Linux AMD64
	@echo "$(BLUE)Building for linux/amd64...$(NC)"
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	
	# Linux ARM64
	@echo "$(BLUE)Building for linux/arm64...$(NC)"
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	
	# macOS AMD64
	@echo "$(BLUE)Building for darwin/amd64...$(NC)"
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	
	# macOS ARM64 (Apple Silicon)
	@echo "$(BLUE)Building for darwin/arm64...$(NC)"
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	
	# Windows AMD64
	@echo "$(BLUE)Building for windows/amd64...$(NC)"
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	@echo "$(GREEN)✓ All release binaries built in $(BUILD_DIR)/releases/$(NC)"

## docker-build: Build Docker image
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t cqlai:$(VERSION) -t cqlai:latest .
	@echo "$(GREEN)✓ Docker image built: cqlai:$(VERSION)$(NC)"

## check: Run all checks (format, lint, test)
check: fmt lint test
	@echo "$(GREEN)✓ All checks passed$(NC)"

## watch: Watch for changes and rebuild (requires entr)
watch:
	@if command -v entr >/dev/null 2>&1; then \
		echo "$(BLUE)Watching for changes...$(NC)"; \
		find . -name '*.go' | entr -r make build; \
	else \
		echo "$(RED)✗ entr not found. Install with your package manager (e.g., apt install entr)$(NC)"; \
		exit 1; \
	fi

## help: Show this help message
help:
	@echo "$(BLUE)cqlai Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Usage:$(NC)"
	@echo "  make [target]"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@grep -E '^## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make build        # Build for current platform"
	@echo "  make install      # Build and install to /usr/local/bin"
	@echo "  make run          # Build and run immediately"
	@echo "  make release-all  # Build for all platforms"
	@echo "  make check        # Run all checks before committing"

# Development shortcuts
b: build
r: run
t: test
c: clean
i: install