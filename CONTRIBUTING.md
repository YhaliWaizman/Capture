# Contributing to Capture

Thank you for your interest in contributing to Capture! This document provides guidelines and instructions for contributing.

## Configuration Files

Configuration files are organized in `.github/config/`:

- `commitlint.json` - Commit message validation rules
- `pre-commit.yaml` - Pre-commit hooks configuration
- `release-please.json` - Release automation configuration
- `release-please-manifest.json` - Version tracking

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/capture.git`
3. Add upstream remote: `git remote add upstream https://github.com/yhaliwaizman/capture.git`
4. Create a feature branch: `git checkout -b feature/your-feature-name`

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Python 3 (for pre-commit hooks)

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Setup git hooks
./scripts/setup-hooks.sh
```

This will install pre-commit hooks that:
- Validate commit message format
- Run `go fmt` on your code
- Run tests before committing
- Check for common issues

## Commit Message Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) for automated changelog generation and semantic versioning.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `build`: Changes to build system or dependencies
- `ci`: Changes to CI configuration files and scripts
- `chore`: Other changes that don't modify src or test files

### Examples

```bash
# Feature
feat: add JSON output format
feat(detector): add Ruby language support

# Bug fix
fix: correct variable detection in Python files
fix(parser): handle quoted values in .env files

# Documentation
docs: update README with Docker Compose examples
docs(api): add JSDoc comments to detector interface

# Performance
perf: implement parallel file processing

# Tests
test: add integration tests for Dockerfile analysis
```

### Scope (Optional)

Common scopes:
- `detector`: Language detectors
- `parser`: File parsers
- `reporter`: Output formatting
- `docker`: Docker-related features
- `cli`: Command-line interface

## Development Workflow

### 1. Make Your Changes

```bash
# Create a feature branch
git checkout -b feat/json-output

# Make your changes
# ...

# Run tests
make test

# Run linter
go vet ./...
```

### 2. Commit Your Changes

```bash
# Stage your changes
git add .

# Commit with conventional commit message
git commit -m "feat: add JSON output format"

# Pre-commit hooks will run automatically
```

### 3. Push and Create PR

```bash
# Push to your fork
git push origin feat/json-output

# Create a pull request on GitHub
```

## Testing

### Run All Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make test-coverage
```

### Run Specific Tests

```bash
go test -v ./internal/detector/...
```

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing (pre-commit hook does this automatically)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small

## Pull Request Guidelines

1. **Title**: Use conventional commit format
   - Good: `feat: add SARIF output format`
   - Bad: `Added new feature`

2. **Description**: Explain what and why
   - What changes were made
   - Why the changes were necessary
   - Any breaking changes
   - Related issues

3. **Tests**: Include tests for new features
   - Unit tests for new functions
   - Integration tests for new features
   - Update existing tests if behavior changes

4. **Documentation**: Update relevant documentation
   - README.md for user-facing changes
   - Code comments for implementation details
   - CHANGELOG.md (handled automatically by release-please)

## Release Process

We use [release-please](https://github.com/googleapis/release-please) for automated releases.

### How It Works

1. Commit changes using conventional commits
2. Push to main branch (via merged PR)
3. Release-please creates/updates a release PR
4. When release PR is merged:
   - Version is bumped automatically
   - CHANGELOG is updated
   - GitHub release is created
   - Binaries are built and attached

### Version Bumping

- `feat:` → Minor version bump (1.0.0 → 1.1.0)
- `fix:` → Patch version bump (1.0.0 → 1.0.1)
- `feat!:` or `BREAKING CHANGE:` → Major version bump (1.0.0 → 2.0.0)

## Project Structure

```
capture/
├── cmd/capture/          # Main application entry point
│   └── cmd/             # CLI commands
├── internal/            # Internal packages
│   ├── detector/        # Language-specific detectors
│   ├── diff/           # Comparison logic
│   ├── dockerfile/     # Dockerfile analysis
│   ├── parser/         # File parsers
│   ├── reporter/       # Output formatting
│   ├── types/          # Shared types and interfaces
│   └── walker/         # File system traversal
├── testdata/           # Test fixtures
└── scripts/            # Development scripts
```

## Getting Help

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Join discussions in pull requests
- Read the [README](README.md) for usage information

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help others learn and grow

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
