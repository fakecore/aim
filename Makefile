.PHONY: build install test clean lint fmt vet help build-all package-all dev-setup dev-install dev-test dev-clean dev-rebuild fish-setup fish-functions fish-info dev-quick-setup dev-env version rc release push

# Build variables
BINARY_NAME=aim
VERSION?=$(shell cat VERSION 2>/dev/null || echo "0.1.3-rc1")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/fakecore/aim/internal/cmd.Version=$(VERSION) \
	-X github.com/fakecore/aim/internal/cmd.GitCommit=$(GIT_COMMIT) \
	-X github.com/fakecore/aim/internal/cmd.BuildDate=$(BUILD_DATE)"

# Directories
BIN_DIR=bin
CMD_DIR=cmd

# Development environment
DEV_HOME=/tmp/aim-dev
DEV_BIN=$(DEV_HOME)/bin
DEV_CONFIG=$(DEV_HOME)/config
DEV_CACHE=$(DEV_HOME)/cache

# Default target
all: build

## help: Display this help message
help:
	@echo "AIM - AI Interface Manager"
	@echo ""
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
	@echo ""

## build: Build the binaries
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./$(CMD_DIR)/aim
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

## install: Install to /usr/local/bin (system-wide, from local build)
install: build
	@./scripts/setup-tool.sh install --local --force

## uninstall: Uninstall from /usr/local/bin
uninstall:
	@./scripts/setup-tool.sh uninstall --force

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out
	@go clean
	@echo "Clean complete"

## coverage: Show test coverage
coverage: test
	@go tool cover -html=coverage.out

## lint: Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete"


## mod: Download and tidy dependencies
mod:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

## build-all: Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BIN_DIR)
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)/aim
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)/aim
	@echo "Building for macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)/aim
	@echo "Building for macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)/aim
	@echo "Cross-platform build complete"

## package-all: Create tar.gz packages for all platforms
package-all: build-all
	@echo "Creating tar.gz packages..."
	@mkdir -p $(BIN_DIR)/packages
	@echo "Packaging for Linux AMD64..."
	@mkdir -p $(BIN_DIR)/packages/$(BINARY_NAME)-linux-amd64
	@cp $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 $(BIN_DIR)/packages/$(BINARY_NAME)-linux-amd64/aim
	@cd $(BIN_DIR)/packages && tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	@echo "Packaging for Linux ARM64..."
	@mkdir -p $(BIN_DIR)/packages/$(BINARY_NAME)-linux-arm64
	@cp $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 $(BIN_DIR)/packages/$(BINARY_NAME)-linux-arm64/aim
	@cd $(BIN_DIR)/packages && tar -czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	@echo "Packaging for macOS AMD64..."
	@mkdir -p $(BIN_DIR)/packages/$(BINARY_NAME)-darwin-amd64
	@cp $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 $(BIN_DIR)/packages/$(BINARY_NAME)-darwin-amd64/aim
	@cd $(BIN_DIR)/packages && tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	@echo "Packaging for macOS ARM64..."
	@mkdir -p $(BIN_DIR)/packages/$(BINARY_NAME)-darwin-arm64
	@cp $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 $(BIN_DIR)/packages/$(BINARY_NAME)-darwin-arm64/aim
	@cd $(BIN_DIR)/packages && tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	@echo "Packaging complete"
	@ls -la $(BIN_DIR)/packages/*.tar.gz

## run: Build and run aim
run: build
	@./$(BIN_DIR)/$(BINARY_NAME)

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed"

## dev-setup: Setup development environment
dev-setup:
	@echo "Setting up dev environment at $(DEV_HOME)"
	@mkdir -p $(DEV_BIN) $(DEV_CONFIG) $(DEV_CACHE)
	@echo "✓ Created $(DEV_HOME)"
	@echo ""
	@echo "Add to your shell rc file:"
	@echo "  export PATH=\"$(DEV_BIN):\$$PATH\""
	@echo "  export AIM_HOME=$(DEV_HOME)"
	@echo ""
	@echo "Or run for current session:"
	@echo "  export PATH=\"$(DEV_BIN):\$$PATH\" && export AIM_HOME=$(DEV_HOME)"

## dev-install: Build and install to dev environment
dev-install: build
	@echo "Installing to dev environment..."
	@cp $(BIN_DIR)/aim $(DEV_BIN)/
	@chmod +x $(DEV_BIN)/aim
	@echo "✓ Installed to $(DEV_BIN)"
	@echo ""
	@echo "Binaries:"
	@ls -lh $(DEV_BIN)/

## dev-test: Run in dev environment
dev-test: dev-install
	@echo "Testing dev installation..."
	@echo ""
	@echo "==> aim version"
	@$(DEV_BIN)/aim version
	@echo ""
	@echo "==> Symlinks"
	@ls -la $(DEV_BIN)/ | grep -E '(aim)'

## dev-rebuild: Quick rebuild and install
dev-rebuild:
	@make -s clean
	@make -s build
	@make -s dev-install

## dev-clean: Clean dev environment
dev-clean:
	@echo "Cleaning dev environment..."
	@rm -rf $(DEV_HOME)
	@echo "✓ Cleaned $(DEV_HOME)"

## fish-setup: Complete setup for Fish shell users
fish-setup: dev-setup dev-install fish-functions
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✓ Fish shell setup complete!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run: source ~/.config/fish/config.fish"
	@echo "  2. Run: aim-dev"
	@echo "  3. Test: aim version"
	@echo ""
	@echo "Or just restart your Fish shell!"
	@echo ""

## fish-functions: Create Fish functions for development
fish-functions:
	@./scripts/setup-fish-functions.sh

## fish-info: Show Fish shell integration status
fish-info:
	@echo "Fish Shell Integration Status"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Functions:"
	@test -f ~/.config/fish/functions/aim-dev.fish && echo "  ✓ aim-dev" || echo "  ✗ aim-dev"
	@test -f ~/.config/fish/functions/aim-rebuild.fish && echo "  ✓ aim-rebuild" || echo "  ✗ aim-rebuild"
	@test -f ~/.config/fish/functions/aim-test.fish && echo "  ✓ aim-test" || echo "  ✗ aim-test"
	@test -f ~/.config/fish/functions/aim-clean.fish && echo "  ✓ aim-clean" || echo "  ✗ aim-clean"
	@test -f ~/.config/fish/functions/aim-info.fish && echo "  ✓ aim-info" || echo "  ✗ aim-info"
	@echo ""
	@echo "Dev Environment:"
	@test -d $(DEV_HOME) && echo "  ✓ $(DEV_HOME)" || echo "  ✗ $(DEV_HOME) (not created)"
	@test -f $(DEV_BIN)/aim && echo "  ✓ aim binary" || echo "  ✗ aim binary"
	@echo ""
	@echo "To setup: make fish-setup"

## version: Show current version information
version:
	@echo "Current version information:"
	@echo "=========================="
	@./scripts/release.sh current

## rc: Create a Release Candidate (RC) version (commits only)
rc:
	@echo "Creating Release Candidate..."
	@./scripts/release.sh rc

## release: Create a final release version (commits only)
release:
	@echo "Creating final release..."
	@./scripts/release.sh release

## push: Push commits and tags to remote repository
push:
	@echo "Pushing commits and tags to remote..."
	@./scripts/release.sh push

## rc-dry: Preview RC version creation without executing
rc-dry:
	@echo "Previewing Release Candidate creation..."
	@./scripts/release.sh rc --dry-run

## release-dry: Preview release version creation without executing
release-dry:
	@echo "Previewing final release creation..."
	@./scripts/release.sh release --dry-run

## push-dry: Preview push operation without executing
push-dry:
	@echo "Previewing push operation..."
	@./scripts/release.sh push --dry-run

## version-set: Set specific version (for manual version control)
version-set:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make version-set VERSION=1.2.3"; \
		exit 1; \
	fi
	@echo "Setting version to $(VERSION)..."
	@echo "$(VERSION)" > VERSION
	@echo "Version updated to $(VERSION)"

## version-patch: Increment patch version (1.2.3 -> 1.2.4)
version-patch:
	@echo "Incrementing patch version..."
	@./scripts/release.sh patch --dry-run

## version-minor: Increment minor version (1.2.3 -> 1.3.0)
version-minor:
	@echo "Incrementing minor version..."
	@./scripts/release.sh minor --dry-run

## version-major: Increment major version (1.2.3 -> 2.0.0)
version-major:
	@echo "Incrementing major version..."
	@./scripts/release.sh major --dry-run