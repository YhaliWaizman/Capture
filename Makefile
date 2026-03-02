.PHONY: build test clean install help

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
	@./capture scan --root . --env-file .env

# Display help
help:
	@echo "Available targets:"
	@echo "  build          - Build the capture executable"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  run            - Run the tool (example usage)"
	@echo "  help           - Display this help message"
