# capture

[![CI](https://github.com/yhaliwaizman/capture/actions/workflows/ci.yml/badge.svg)](https://github.com/yhaliwaizman/capture/actions/workflows/ci.yml)
[![Release](https://github.com/yhaliwaizman/capture/actions/workflows/release.yml/badge.svg)](https://github.com/yhaliwaizman/capture/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yhaliwaizman/capture)](https://goreportcard.com/report/github.com/yhaliwaizman/capture)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A static analysis CLI tool that identifies mismatches between environment variables declared in .env files and those referenced in source code.

## Features

- 🔍 Detects environment variable usage in JavaScript, TypeScript, Go, and Python
- 🐳 Analyzes Dockerfiles for ENV/ARG declarations and variable usage
- 🔄 Cross-checks variables between .env, Dockerfile, and source code
- 🎯 Pattern-based detection without AST parsing for simplicity and speed
- 🔄 Deterministic output for reliable CI/CD integration
- ⚡ Memory-efficient streaming file processing
- 🚫 Configurable directory ignore patterns
- 📊 Clear reporting of unused and missing variables with file locations

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/yhaliwaizman/capture/releases).

**Linux (x86_64):**
```bash
curl -L https://github.com/yhaliwaizman/capture/releases/latest/download/capture_Linux_x86_64.tar.gz | tar xz
sudo mv capture /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/yhaliwaizman/capture/releases/latest/download/capture_Darwin_x86_64.tar.gz | tar xz
sudo mv capture /usr/local/bin/
```

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/yhaliwaizman/capture/releases/latest/download/capture_Darwin_arm64.tar.gz | tar xz
sudo mv capture /usr/local/bin/
```

**Windows:**
Download `capture_Windows_x86_64.zip` from the releases page and extract.

### Using Go Install

```bash
go install github.com/yhaliwaizman/capture/cmd/capture@latest
```

### From Source

```bash
# Clone the repository
git clone https://github.com/yhaliwaizman/capture.git
cd capture

# Build the binary
go build -o capture ./cmd/capture

# Or use make
make build

# Install to GOPATH/bin
make install
```

## Usage

### Basic Scan

```bash
capture scan --root ./project --env-file .env
```

### With Ignore Patterns

```bash
capture scan --root ./project --env-file .env --ignore vendor,tmp,cache
```

### Example Output

**No mismatches:**
```
No environment mismatches found.
```

**With mismatches:**
```
Declared but unused:
- OLD_API_KEY
- DEPRECATED_URL

Used but not declared:
- DATABASE_URL (src/db.go:15)
- REDIS_HOST (src/cache.js:8)

Code uses variables not in Dockerfile or .env:
- MISSING_VAR (src/app.js:10)

Dockerfile declares but code doesn't use:
- BUILD_VERSION
- NODE_ENV

Dockerfile uses undeclared variables:
- UNDEFINED_VAR (Dockerfile:15)
```

## Command-Line Options

- `--root` (required): Root directory to scan for source files
- `--env-file` (required): Path to the .env file
- `--ignore` (optional): Comma-separated list of directories to ignore

## Exit Codes

- **0**: No mismatches detected (success)
- **1**: Mismatches found (unused or missing variables)
- **2**: Configuration error (missing file, invalid flags, permissions)

## Supported Languages

| Language   | Patterns Detected |
|------------|-------------------|
| JavaScript | `process.env.VAR`, `process.env["VAR"]`, `process.env['VAR']` |
| TypeScript | `process.env.VAR`, `process.env["VAR"]`, `process.env['VAR']` |
| Go         | `os.Getenv("VAR")`, `os.LookupEnv("VAR")` |
| Python     | `os.getenv("VAR")`, `os.environ["VAR"]`, `os.environ['VAR']` |
| Dockerfile | `ENV KEY=value`, `ARG KEY=default`, `$VAR`, `${VAR}` |

## Dockerfile Analysis

The tool automatically detects and analyzes Dockerfiles in your project:

- **Detected files**: `Dockerfile`, `Dockerfile.*`, `*.dockerfile`
- **Declarations**: Extracts `ENV` and `ARG` instructions
- **Usage**: Detects `$VAR` and `${VAR}` references in RUN, CMD, etc.
- **Cross-checks**: Compares Dockerfile variables with .env and source code

### Dockerfile Patterns

**Declarations:**
```dockerfile
ENV API_KEY=default_value
ENV DATABASE_URL postgres://localhost
ENV A=1 B=2 C=3
ARG BUILD_VERSION
ARG NODE_ENV=production
```

**Usage:**
```dockerfile
RUN echo "Version: $BUILD_VERSION"
RUN echo "API: ${API_KEY}"
```

### Multi-Stage Dockerfiles

The tool analyzes all stages in multi-stage builds:

```dockerfile
FROM node:18 AS builder
ENV BUILD_ENV=production
ARG BUILD_VERSION

FROM node:18-alpine AS runtime
ENV RUNTIME_ENV=production
RUN echo "Build: $BUILD_ENV"
```

All `ENV` and `ARG` declarations from all stages are collected and cross-checked.

## Development

### Build

```bash
make build
```

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Clean Build Artifacts

```bash
make clean
```

## CI/CD Integration

The tool is designed for CI/CD pipelines with deterministic output and standard exit codes:

```yaml
# GitHub Actions example
- name: Check environment variables
  run: |
    ./capture scan --root . --env-file .env
```

```bash
# GitLab CI example
check-env:
  script:
    - ./capture scan --root . --env-file .env
```

## How It Works

1. **Parse .env file**: Extracts declared variable names matching `^[A-Z][A-Z0-9_]*# capture
2. **Walk directory tree**: Recursively finds source files (.js, .ts, .go, .py) and Dockerfiles
3. **Analyze Dockerfiles**: Extracts ENV/ARG declarations and variable usage
4. **Detect usage**: Applies regex patterns to find environment variable references in source code
5. **Compare sets**: Identifies mismatches:
   - Unused: declared in .env but not used in code
   - Missing: used in code but not declared in .env
   - Code uses variables not in Dockerfile or .env
   - Dockerfile declares variables unused in code
   - Dockerfile uses undeclared variables
6. **Generate report**: Outputs deterministic, sorted results

## License

MIT
