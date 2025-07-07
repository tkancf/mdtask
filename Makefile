.PHONY: all build clean deps css js watch test release

# Variables
BINARY_NAME=mdtask
GO_FILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
CSS_INPUT=internal/web/static/css/input.css
CSS_OUTPUT=internal/web/static/css/style.css

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

# Default target
all: deps css js build

# Install dependencies
deps:
	@echo "Installing npm dependencies..."
	@npm install

# Build CSS
css:
	@echo "Building CSS..."
	@npm run build-css

# Build JavaScript
js:
	@echo "Building JavaScript..."
	@npm run build-js

# Watch CSS (for development)
watch:
	@echo "Watching CSS..."
	@npm run watch-css

# Build binary
build: css js
	@echo "Building binary..."
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Build for release (all platforms)
release: css js
	@echo "Building release binaries for version $(VERSION)..."
	@mkdir -p dist
	
	# macOS
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 .
	
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 .
	
	# Linux
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 .
	
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 .
	
	# Windows
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "Release builds complete!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf dist/
	@rm -f $(CSS_OUTPUT)

# Development mode - build and run
dev: css js
	@echo "Starting in development mode..."
	@go run main.go web

# Install locally
install: build
	@echo "Installing to /usr/local/bin..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/