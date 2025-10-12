# Makefile

# Define default Go build flags
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0-dev)
LDFLAGS = -s -w -X 'parm/parmver.StringVersion=$(VERSION)'
DEBUG_FLAGS =

# Set GoOS and GoARCH (can override in command line)
GOOS ?= linux
GOARCH ?= amd64

# The binary name and output location
BINARY_NAME = parm
OUTPUT_DIR = ./bin

# Make sure the output directory exists
.PHONY: $(OUTPUT_DIR)
$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

# Default target (build all platforms and create tarballs/zips)
all: build
release: release-linux release-darwin release-windows

build: | $(OUTPUT_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)

debug: | $(OUTPUT_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(DEBUG_FLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)

.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Build and create tarball for Linux
release-linux: | $(OUTPUT_DIR)
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)
	@echo "Creating tarball for Linux..."
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)
	@echo "Deleting binary for Linux..."
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)

# Build and create tarball for macOS (Darwin)
release-darwin: | $(OUTPUT_DIR)
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)
	@echo "Creating tarball for macOS..."
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)
	@echo "Deleting binary for macOS..."
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)

# Build and create zip file for Windows
release-windows: | $(OUTPUT_DIR)
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME).exe
	@echo "Creating zip file for Windows..."
	zip -r $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.zip $(OUTPUT_DIR)/$(BINARY_NAME).exe
	@echo "Deleting binary for Windows..."
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME).exe

# Clean up build artifacts
clean:
	rm -rf $(OUTPUT_DIR)

format:
	gofmt -w .
