BINARY_NAME=parm
BIN_DIR=bin
BUILD_DIR=dist
INSTALL_PATH=$(HOME)/.local/bin
COMMIT=$(shell git rev-parse --short HEAD)
DATE=$(shell date +%Y-%m-%d)
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
REPO_FULL := $(shell git remote get-url origin 2>/dev/null | sed 's/.*github.com[\/:]//;s/\.git//' || echo "aleister1102/parm")
OWNER := $(shell echo $(REPO_FULL) | cut -d/ -f1)
REPO := $(shell echo $(REPO_FULL) | cut -d/ -f2)

VERSION_PARTS := $(subst ., ,$(subst v,,$(CURRENT_TAG)))
MAJOR := $(word 1,$(VERSION_PARTS))
MINOR := $(word 2,$(VERSION_PARTS))
PATCH := $(word 3,$(VERSION_PARTS))
NEXT_PATCH := v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1)))
NEXT_MINOR := v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0
NEXT_MAJOR := v$(shell echo $$(($(MAJOR)+1))).0.0

LDFLAGS = -X 'parm/parmver.StringVersion=$(CURRENT_TAG)' -X 'parm/parmver.Owner=$(OWNER)' -X 'parm/parmver.Repo=$(REPO)' -s -w

.DEFAULT_GOAL := help

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build       Build binary for current platform"
	@echo "  clean       Remove build artifacts"
	@echo "  test        Run tests"
	@echo "  install     Build and install to $(INSTALL_PATH)"
	@echo "  uninstall   Remove from $(INSTALL_PATH)"
	@echo "  release     Release new version (usage: make release TAG=v1.0.0)"
	@echo "  bump-patch  Release next patch version"
	@echo "  bump-minor  Release next minor version"
	@echo "  bump-major  Release next major version"

build:
	mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) main.go

clean:
	rm -rf $(BIN_DIR)
	rm -rf $(BUILD_DIR)


test:
	go test ./...

install: build
	mkdir -p $(INSTALL_PATH)
	cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_PATH)"

uninstall:
	rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Removed $(BINARY_NAME) from $(INSTALL_PATH)"

release: ## Release new version (usage: make release TAG=v1.0.0)
	@if [ -z "$(TAG)" ]; then echo "Usage: make release TAG=v1.0.0"; exit 1; fi
	@echo "Releasing $(TAG) to $(OWNER)/$(REPO)..."
	@sed -i.bak "s/{{OWNER}}/$(OWNER)/g" scripts/install.sh && sed -i.bak "s/{{REPO}}/$(REPO)/g" scripts/install.sh && rm scripts/install.sh.bak
	@go test ./...
	@rm -rf $(BUILD_DIR) && mkdir -p $(BUILD_DIR)
	@echo "Building binaries..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "-X 'parm/parmver.StringVersion=$(TAG)' -X 'parm/parmver.Owner=$(OWNER)' -X 'parm/parmver.Repo=$(REPO)' -s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	@GOOS=linux GOARCH=arm64 go build -ldflags "-X 'parm/parmver.StringVersion=$(TAG)' -X 'parm/parmver.Owner=$(OWNER)' -X 'parm/parmver.Repo=$(REPO)' -s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	@GOOS=darwin GOARCH=amd64 go build -ldflags "-X 'parm/parmver.StringVersion=$(TAG)' -X 'parm/parmver.Owner=$(OWNER)' -X 'parm/parmver.Repo=$(REPO)' -s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	@GOOS=darwin GOARCH=arm64 go build -ldflags "-X 'parm/parmver.StringVersion=$(TAG)' -X 'parm/parmver.Owner=$(OWNER)' -X 'parm/parmver.Repo=$(REPO)' -s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	@GOOS=windows GOARCH=amd64 go build -ldflags "-X 'parm/parmver.StringVersion=$(TAG)' -X 'parm/parmver.Owner=$(OWNER)' -X 'parm/parmver.Repo=$(REPO)' -s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@GOOS=windows GOARCH=arm64 go build -ldflags "-X 'parm/parmver.StringVersion=$(TAG)' -X 'parm/parmver.Owner=$(OWNER)' -X 'parm/parmver.Repo=$(REPO)' -s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe main.go
	@cd $(BUILD_DIR) && shasum -a 256 * > checksums.txt
	@echo "Creating GitHub release..."
	@if ! git rev-parse $(TAG) >/dev/null 2>&1; then \
		git tag -a $(TAG) -m "Release $(TAG)"; \
	fi
	@git push origin $(TAG)
	@echo "Done: $(TAG)"
	@echo "Binaries are available in $(BUILD_DIR)"

bump-patch: ## Release next patch version
	@$(MAKE) release TAG=$(NEXT_PATCH)

bump-minor: ## Release next minor version
	@$(MAKE) release TAG=$(NEXT_MINOR)

bump-major: ## Release next major version
	@$(MAKE) release TAG=$(NEXT_MAJOR)
