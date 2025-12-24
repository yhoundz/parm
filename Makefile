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

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Make sure the output directory exists
.PHONY: $(OUTPUT_DIR)
$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

all: build ## Build the project (default build)

release: release-linux release-darwin release-windows ## Build releases for all supported platforms

build: | $(OUTPUT_DIR) ## Build the binary for the current OS and architecture
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)

debug: | $(OUTPUT_DIR) ## Build the binary with debug flags
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(DEBUG_FLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)

.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	go test ./...

release-linux: | $(OUTPUT_DIR) ## Build and package for Linux (amd64 and arm64)
	@echo "Building for Linux amd64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)
	@echo "Creating tarball for Linux amd64..."
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)

	@echo "Building for Linux arm64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)
	@echo "Creating tarball for Linux arm64..."
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)

release-darwin: | $(OUTPUT_DIR) ## Build and package for macOS (amd64 and arm64)
	@echo "Building for macOS amd64..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)
	@echo "Creating tarball for macOS amd64..."
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)

	@echo "Building for macOS arm64..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)
	@echo "Creating tarball for macOS arm64..."
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)

release-windows: | $(OUTPUT_DIR) ## Build and package for Windows (amd64)
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME).exe
	@echo "Creating zip file for Windows..."
	zip -r $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.zip $(OUTPUT_DIR)/$(BINARY_NAME).exe
	@echo "Deleting binary for Windows..."
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME).exe

clean: ## Remove build artifacts
	rm -rf $(OUTPUT_DIR)

format: ## Format Go source code
	gofmt -w .
