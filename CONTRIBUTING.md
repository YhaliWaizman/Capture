# Contributing to capture

Thank you for your interest in contributing to capture! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, but recommended)

### Setting Up Development Environment

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/yhaliwaizman/capture.git
   cd capture
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the project:**
   ```bash
   go build -o capture ./cmd/capture
   # or
   make build
   ```

4. **Run tests:**
   ```bash
   go test ./...
   # or
   make test
   ```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Changes

- Write clean, readable code
- Follow Go conventions and best practices
- Add tests for new functionality
- Update documentation as needed

### 3. Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

### 4. Check Code Quality

```bash
# Format code
go fmt ./...

# Run linter (if installed)
golangci-lint run

# Check for common issues
go vet ./...
```

### 5. Commit Changes

Follow conventional commit format:

```bash
git commit -m "feat: add new feature"
git commit -m "fix: resolve bug in parser"
git commit -m "docs: update README"
git commit -m "test: add tests for detector"
git commit -m "chore: update dependencies"
```

**Commit types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or changes
- `chore`: Maintenance tasks
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `ci`: CI/CD changes

### 6. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Code Guidelines

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Keep functions small and focused
- Write descriptive variable names
- Add comments for exported functions

### Testing

- Write unit tests for all new code
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Test edge cases and error conditions
- Use descriptive test names

**Example:**
```go
func TestFeature_SpecificCase(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"
    
    // Act
    result := Feature(input)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

### Documentation

- Add godoc comments for exported functions
- Update README.md for user-facing changes
- Add examples for new features
- Update CHANGELOG.md

## Project Structure

```
capture/
├── cmd/capture/          # Main application entry point
├── internal/             # Internal packages
│   ├── detector/        # Language-specific detectors
│   ├── diff/            # Comparison logic
│   ├── dockerfile/      # Dockerfile analysis
│   ├── parser/          # .env file parser
│   ├── reporter/        # Output formatting
│   ├── types/           # Shared types
│   └── walker/          # File system traversal
├── testdata/            # Test fixtures
├── .github/workflows/   # CI/CD workflows
└── docs/                # Additional documentation
```

## Adding New Features

### Adding a New Language Detector

1. Create detector in `internal/detector/`
2. Implement `LanguageDetector` interface
3. Add to `DetectorFactory`
4. Write comprehensive tests
5. Update documentation

### Adding New Analysis Type

1. Create new package in `internal/`
2. Define clear interfaces
3. Integrate into CLI pipeline
4. Add tests and documentation
5. Update README with examples

## Testing

### Running Specific Tests

```bash
# Run tests for specific package
go test ./internal/detector

# Run specific test
go test -run TestJSDetector ./internal/detector

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...
```

### Writing Integration Tests

Place integration tests in `cmd/capture/*_test.go`:

```go
func TestCLI_NewFeature(t *testing.T) {
    binary := buildBinary(t)
    // Test CLI behavior
}
```

## Pull Request Process

1. **Ensure all tests pass**
2. **Update documentation**
3. **Add entry to CHANGELOG.md**
4. **Request review from maintainers**
5. **Address review feedback**
6. **Wait for CI checks to pass**
7. **Maintainer will merge**

### PR Checklist

- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Code formatted (`go fmt`)
- [ ] All tests passing
- [ ] No linter warnings
- [ ] Commit messages follow convention

## Release Process

Releases are automated via GitHub Actions:

1. Update CHANGELOG.md
2. Commit changes
3. Create and push tag: `git tag v1.0.0 && git push origin v1.0.0`
4. GitHub Actions builds and publishes release

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/yhaliwaizman/capture/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yhaliwaizman/capture/discussions)
- **Documentation**: See README.md and docs/

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Help others learn and grow

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in:
- GitHub contributors page
- Release notes
- CHANGELOG.md (for significant contributions)

Thank you for contributing to capture! 🎉
