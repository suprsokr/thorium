# Thorium CLI Makefile

BINARY_NAME=thorium
VERSION=1.0.0
BUILD_DIR=build
STORMLIB_DIR=stormlib

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Detect OS and architecture
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Build targets
.PHONY: all build build-pure clean test deps stormlib help

# Default: build with StormLib (all-in-one)
all: build

# Build with cgo + StormLib (all-in-one binary) - this is the default
build: stormlib
	@echo "Building thorium CLI with StormLib..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(STORMLIB_DIR)/src" \
	CGO_LDFLAGS="-L$(STORMLIB_DIR)/build -lstorm -lz -lbz2 -lstdc++" \
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/thorium
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME) (with built-in MPQ support)"

# Build without cgo (pure Go, requires external mpqbuilder)
build-pure:
	@echo "Building thorium CLI (pure Go)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/thorium
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Note: This build requires external 'mpqbuilder' tool for MPQ operations"

# Build StormLib as static library
stormlib:
	@echo "Building StormLib..."
	@mkdir -p $(STORMLIB_DIR)/build
	@cd $(STORMLIB_DIR)/build && cmake .. -DBUILD_SHARED_LIBS=OFF && make -j4
	@echo "StormLib built: $(STORMLIB_DIR)/build/libstorm.a"

# Build for multiple platforms (pure Go only - no cgo cross-compile)
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/thorium
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/thorium

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/thorium
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/thorium

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/thorium

# Install to system
install: build
	@echo "Installing thorium to /usr/local/bin..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installed: /usr/local/bin/$(BINARY_NAME)"

# Run tests
test:
	$(GOTEST) -v ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(STORMLIB_DIR)/build

# Development helpers
dev: build
	@cp $(BUILD_DIR)/$(BINARY_NAME) ./$(BINARY_NAME)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Show help
help:
	@echo "Thorium CLI Makefile"
	@echo ""
	@echo "Build Targets:"
	@echo "  build          Build all-in-one binary with StormLib (default)"
	@echo "  build-pure     Build pure Go binary (requires external mpqbuilder)"
	@echo "  build-all      Build pure Go for Linux/macOS/Windows"
	@echo "  stormlib       Build StormLib static library"
	@echo ""
	@echo "Install Targets:"
	@echo "  install        Install to /usr/local/bin"
	@echo ""
	@echo "Other Targets:"
	@echo "  test           Run tests"
	@echo "  deps           Download Go dependencies"
	@echo "  clean          Remove build artifacts"
	@echo "  help           Show this help"
