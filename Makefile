BINARY_NAME=xks
VERSION?=1.0.0
BUILD_DIR=dist
LDFLAGS=-ldflags "-X main.version=$(VERSION) -s -w"

# Detect OS for default build
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

.PHONY: all build build-local build-linux build-windows build-darwin build-arm build-all clean test deps lint install help release dev security

# Default target
all: clean deps test build-all

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

lint: ## Run linter
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found, skipping..."; \
	fi

build-local: deps ## Build for current OS/ARCH
	@echo "🔧 Building for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

build-linux: deps ## Build for Linux x86_64
	@echo "🔧 Building for Linux (x86_64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

build-windows: deps ## Build for Windows x86_64
	@echo "🔧 Building for Windows (x86_64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

build-darwin: deps ## Build for macOS x86_64
	@echo "🔧 Building for macOS (x86_64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .

build-arm: deps ## Build ARM versions
	@echo "🔧 Building for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "🔧 Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .

build-all: build-linux build-windows build-darwin build-arm ## Build for all platforms

install: build-local ## Install locally
	@echo "📥 Installing $(BINARY_NAME)..."
	@if [ "$(GOOS)" = "windows" ]; then \
		cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/ 2>/dev/null || echo "⚠️  Cannot install to /usr/local/bin, copy manually"; \
	else \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/ || echo "⚠️  Cannot install to /usr/local/bin, copy manually"; \
	fi

release: clean test build-all ## Create release builds
	@echo "📦 Creating release archive..."
	@cd $(BUILD_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [ -f "$$binary" ]; then \
			echo "Creating archive for $$binary..."; \
			if echo "$$binary" | grep -q ".exe$$"; then \
				zip "$$binary-$(VERSION).zip" "$$binary"; \
			else \
				tar -czf "$$binary-$(VERSION).tar.gz" "$$binary"; \
			fi \
		fi \
	done
	@echo "✅ Release archives created in $(BUILD_DIR)/"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)/

dev: ## Run in development mode
	@echo "🚀 Running in development mode..."
	go run . $(ARGS)

# Security scan
security: ## Run security scan
	@echo "🔒 Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not found, install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi