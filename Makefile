.PHONY: build test clean install help setup dev-setup

# Setup development environment
setup: dev-setup

dev-setup:
	@echo "Setting up development environment..."
	@echo ""
	@echo "1. Installing Go dependencies..."
	@go mod download
	@echo ""
	@echo "2. Setting up git hooks..."
	@./scripts/install-git-hooks.sh
	@echo ""
	@echo "✓ Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  - Run 'make test' to verify setup"
	@echo "  - Read CONTRIBUTING.md for guidelines"
	@echo "  - Use conventional commits (feat:, fix:, etc.)"
	@echo ""
	@echo "Optional: Install pre-commit for advanced hooks"
	@echo "  sudo pacman -S python-pre-commit  # Arch Linux"
	@echo "  brew install pre-commit            # macOS"
	@echo "  Then run: ./scripts/setup-hooks.sh"

# Build the capture executable
build:
	@echo "Building capture..."
	@go build -o capture ./cmd/capture
	@echo "Build complete: ./capture"

# Run all tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f capture coverage.out coverage.html
	@echo "Clean complete"

# Install the binary to GOPATH/bin
install:
	@echo "Installing capture..."
	@go install ./cmd/capture
	@echo "Install complete"

# Run the tool (example usage)
run:
	@./capture scan --dir . --env-file .env

# Display help
help:
	@echo "Available targets:"
	@echo "  setup          - Setup development environment (install deps + git hooks)"
	@echo "  dev-setup      - Alias for setup"
	@echo "  build          - Build the capture executable"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  run            - Run the tool (example usage)"
	@echo "  help           - Display this help message"
