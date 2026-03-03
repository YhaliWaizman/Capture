# Development Guide

This guide covers the development setup and workflow for contributing to Capture.

## Quick Start

```bash
# Clone the repository
git clone https://github.com/yhaliwaizman/capture.git
cd capture

# Install dependencies
go mod download

# Setup git hooks (requires Python 3)
./scripts/setup-hooks.sh

# Build
make build

# Run tests
make test
```

## Git Hooks Setup

We use pre-commit hooks to ensure code quality and consistent commit messages.

### Installation

```bash
./scripts/setup-hooks.sh
```

This installs:
- **Commit message validation**: Ensures conventional commit format
- **Code formatting**: Runs `go fmt` automatically
- **Tests**: Runs tests before allowing commits
- **File checks**: Validates YAML, checks for large files, etc.

### Manual Installation

If you prefer to install manually:

```bash
# Install pre-commit (requires Python 3)
pip3 install pre-commit

# Install hooks
pre-commit install --hook-type pre-commit
pre-commit install --hook-type commit-msg
```

### Running Hooks Manually

```bash
# Run all hooks on all files
pre-commit run --all-files

# Run specific hook
pre-commit run go-fmt --all-files
pre-commit run go-test --all-files
```

### Bypassing Hooks (Not Recommended)

```bash
# Skip pre-commit hooks
git commit --no-verify -m "message"

# Skip commit-msg hook
git commit -n -m "message"
```

## Conventional Commits

All commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

### Examples

```bash
# Feature
git commit -m "feat: add JSON output format"
git commit -m "feat(detector): add Ruby language support"

# Bug fix
git commit -m "fix: correct variable detection in Python files"
git commit -m "fix(parser): handle quoted values in .env files"

# Documentation
git commit -m "docs: update README with examples"

# Performance
git commit -m "perf: implement parallel file processing"

# Breaking change
git commit -m "feat!: change CLI flag names"
# or
git commit -m "feat: change CLI flag names

BREAKING CHANGE: --env-file is now --env"
```

### Commit Types

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat` | New feature | Minor (1.0.0 → 1.1.0) |
| `fix` | Bug fix | Patch (1.0.0 → 1.0.1) |
| `docs` | Documentation only | None |
| `style` | Code style (formatting) | None |
| `refactor` | Code refactoring | None |
| `perf` | Performance improvement | Patch |
| `test` | Adding tests | None |
| `build` | Build system changes | None |
| `ci` | CI configuration | None |
| `chore` | Other changes | None |
| `feat!` or `BREAKING CHANGE:` | Breaking change | Major (1.0.0 → 2.0.0) |

## Release Process

We use [release-please](https://github.com/googleapis/release-please) for automated releases.

### How It Works

1. **Commit with conventional commits**
   ```bash
   git commit -m "feat: add new feature"
   ```

2. **Push to main** (via merged PR)
   ```bash
   git push origin main
   ```

3. **Release-please creates a PR**
   - Automatically updates CHANGELOG.md
   - Bumps version based on commit types
   - Creates release notes

4. **Merge the release PR**
   - GitHub release is created
   - GoReleaser builds binaries
   - Binaries are attached to release

### Version Bumping Rules

- `feat:` commits → Minor version (1.0.0 → 1.1.0)
- `fix:` commits → Patch version (1.0.0 → 1.0.1)
- `feat!:` or `BREAKING CHANGE:` → Major version (1.0.0 → 2.0.0)
- Other types → No version bump

### Manual Release (Emergency)

If automated release fails:

```bash
# Create and push tag
git tag v1.0.1
git push origin v1.0.1

# GoReleaser will run automatically
```

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/detector/...

# Run specific test
go test -run TestJSDetector ./internal/detector/
```

### Integration Tests

```bash
# Run integration tests
go test -tags=integration ./...
```

### Test Coverage

We aim for >80% test coverage. Check coverage:

```bash
make test-coverage
# Opens coverage.html in browser
```

## Code Quality

### Formatting

```bash
# Format all code
go fmt ./...

# Check formatting
gofmt -l .
```

### Linting

```bash
# Run go vet
go vet ./...

# Run golangci-lint (if installed)
golangci-lint run
```

### Static Analysis

```bash
# Check for common mistakes
go vet ./...

# Check for race conditions
go test -race ./...
```

## Debugging

### Debug Build

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o capture-debug ./cmd/capture

# Use with delve
dlv exec ./capture-debug -- scan --root . --env-file .env
```

### Verbose Output

```bash
# Run with verbose logging
./capture scan --root . --env-file .env --verbose
```

## Troubleshooting

### Pre-commit Hook Issues

**Problem**: Hooks not running

```bash
# Reinstall hooks
pre-commit uninstall
pre-commit install --hook-type pre-commit
pre-commit install --hook-type commit-msg
```

**Problem**: Python not found

```bash
# Install Python 3
# macOS
brew install python3

# Ubuntu/Debian
sudo apt-get install python3 python3-pip

# Then reinstall pre-commit
pip3 install pre-commit
```

### Test Failures

**Problem**: Tests fail locally but pass in CI

```bash
# Clean test cache
go clean -testcache

# Run tests with verbose output
go test -v ./...
```

### Build Issues

**Problem**: Build fails with dependency errors

```bash
# Clean and reinstall dependencies
go clean -modcache
go mod download
go mod tidy
```

## Additional Resources

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Release Please](https://github.com/googleapis/release-please)
- [Pre-commit](https://pre-commit.com/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Testing](https://golang.org/pkg/testing/)
