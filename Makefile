# Makefile

# Define default Go build flags
LDFLAGS = -s -w
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
all: build-linux build-darwin build-windows

build: | $(OUTPUT_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)

debug: | $(OUTPUT_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(DEBUG_FLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)

test:
	go test ./...

# TODO: include setting release global variables and versioning
# Build and create tarball for Linux
build-linux: | $(OUTPUT_DIR)
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64
	@echo "Creating tarball for Linux..."
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)-linux-amd64
	@echo "Deleting binary for Linux..."
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64

# Build and create tarball for macOS (Darwin)
build-darwin: | $(OUTPUT_DIR)
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64
	@echo "Creating tarball for macOS..."
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)-darwin-amd64
	@echo "Deleting binary for macOS..."
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64

# Build and create zip file for Windows
build-windows: | $(OUTPUT_DIR)
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@echo "Creating zip file for Windows..."
	zip -r $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.zip $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@echo "Deleting binary for Windows..."
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe

# Clean up build artifacts
clean:
	rm -rf $(OUTPUT_DIR)
