# capture

A static analysis CLI tool that identifies mismatches between environment variables declared in .env files and those referenced in source code.

## Features

- 🔍 Detects environment variable usage in JavaScript, TypeScript, Go, and Python
- 🎯 Pattern-based detection without AST parsing for simplicity and speed
- 🔄 Deterministic output for reliable CI/CD integration
- ⚡ Memory-efficient streaming file processing
- 🚫 Configurable directory ignore patterns
- 📊 Clear reporting of unused and missing variables with file locations

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yhaliwaizman/capture.git
cd capture

# Build the binary
make build

# Or install to GOPATH/bin
make install
```

### Using Go Install

```bash
go install github.com/yhaliwaizman/capture/cmd/capture@latest
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

1. **Parse .env file**: Extracts declared variable names matching `^[A-Z][A-Z0-9_]*$`
2. **Walk directory tree**: Recursively finds source files (.js, .ts, .go, .py)
3. **Detect usage**: Applies regex patterns to find environment variable references
4. **Compare sets**: Identifies unused (declared but not used) and missing (used but not declared)
5. **Generate report**: Outputs deterministic, sorted results

## License

MIT
