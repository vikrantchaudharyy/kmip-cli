# Makefile for kmip-cli

# Variables
BINARY_NAME=kmip-cli
BUILD_DIR=./build
SRC=./...

# Default target
.DEFAULT_GOAL := build

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy

# Build for the current platform
build: deps
	@echo "Building for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@go build -o $(BINARY_NAME) $(SRC)

# Build for all supported platforms
all: clean build-linux build-windows build-macos-amd64 build-macos-arm64

# Build for Linux (amd64)
build-linux: deps
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SRC)

# Build for Windows (amd64)
build-windows: deps
	@echo "Building for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SRC)

# Build for macOS (amd64)
build-macos-amd64: deps
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(SRC)

# Build for macOS (arm64)
build-macos-arm64: deps
	@echo "Building for macOS (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(SRC)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)

# Phony targets
.PHONY: build all clean build-linux build-windows build-macos-amd64 build-macos-arm64 deps
