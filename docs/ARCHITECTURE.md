# Architecture

## Overview

Capture is a CLI tool for detecting environment variable mismatches between `.env` files, source code, and Dockerfiles.

## Design Principles

1. **Pattern-based detection** - No AST parsing for simplicity and speed
2. **Streaming processing** - Memory-efficient for large codebases
3. **Deterministic output** - Consistent results for CI/CD
4. **Extensible** - Easy to add new language detectors

## Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                           │
│                    (cmd/capture/cmd/)                       │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Orchestration                            │
│                  (cmd/capture/cmd/scan.go)                  │
│  • Coordinates all components                               │
│  • Manages workflow                                         │
│  • Handles errors                                           │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼               ▼
┌──────────────┐ ┌──────────┐ ┌──────────────┐
│   Parser     │ │  Walker  │ │  Dockerfile  │
│              │ │          │ │   Analyzer   │
│ • .env files │ │ • Find   │ │ • Parse ENV  │
│ • Extract    │ │   files  │ │ • Parse ARG  │
│   variables  │ │ • Filter │ │ • Find usage │
└──────┬───────┘ └────┬─────┘ └───────┬──────┘
       │              │               │
       │              ▼               │
       │      ┌──────────────┐        │
       │      │  Detectors   │        │
       │      │              │        │
       │      │ • JS/TS      │        │
       │      │ • Go         │        │
       │      │ • Python     │        │
       │      └──────┬───────┘        │
       │             │                │
       └─────────────┼────────────────┘
                     ▼
            ┌─────────────────┐
            │   Diff Engine   │
            │                 │
            │ • Compare sets  │
            │ • Find unused   │
            │ • Find missing  │
            └────────┬────────┘
                     ▼
            ┌─────────────────┐
            │    Reporter     │
            │                 │
            │ • Format output │
            │ • Sort results  │
            │ • Display       │
            └─────────────────┘
```

## Package Structure

### `cmd/capture/`
Main application entry point. Contains CLI command definitions using Cobra.

**Files:**
- `main.go` - Entry point
- `cmd/root.go` - Root command setup
- `cmd/scan.go` - Scan command implementation
- `cmd/version.go` - Version command

### `internal/types/`
Shared types and interfaces used across packages.

**Key types:**
- `Location` - File path and line number
- `DiffResult` - Comparison results
- `ReportData` - Report formatting data

**Interfaces:**
- `EnvParser` - .env file parsing
- `FileWalker` - Directory traversal
- `LanguageDetector` - Variable detection
- `DiffEngine` - Set comparison
- `Reporter` - Output formatting

### `internal/parser/`
Parses `.env` files and extracts variable names.

**Pattern:** `^[A-Z][A-Z0-9_]*\s*=`

**Features:**
- Skips comments
- Validates uppercase naming
- Handles whitespace

### `internal/walker/`
Recursively traverses directories to find source files.

**Features:**
- Filters by extension (`.js`, `.ts`, `.go`, `.py`)
- Respects ignore patterns
- Skips symbolic links
- Detects Dockerfiles

### `internal/detector/`
Language-specific detectors for finding environment variable usage.

**Factory pattern:**
- `DetectorFactory` - Creates appropriate detector for file extension
- Each detector implements `LanguageDetector` interface

**Detectors:**
- `JSDetector` - JavaScript/TypeScript (`process.env.VAR`)
- `GoDetector` - Go (`os.Getenv("VAR")`)
- `PythonDetector` - Python (`os.getenv("VAR")`)

**Pattern matching:**
- Regex-based detection
- Rejects dynamic expressions
- Records line numbers
- Sorts results deterministically

### `internal/dockerfile/`
Analyzes Dockerfiles for ENV/ARG declarations and variable usage.

**Features:**
- Parses ENV and ARG instructions
- Detects `$VAR` and `${VAR}` usage
- Handles line continuations
- Validates FROM instruction
- Supports multi-stage builds

### `internal/diff/`
Compares declared and used variable sets.

**Operations:**
- Unused: `declared - used`
- Missing: `used - declared`
- Sorts results alphabetically

### `internal/reporter/`
Formats and outputs analysis results.

**Output sections:**
1. Declared but unused
2. Used but not declared
3. Docker-specific mismatches

**Features:**
- Deterministic output
- File locations for missing variables
- Sorted results

## Data Flow

```
1. Parse .env file
   ↓
2. Walk directory tree
   ↓
3. For each file:
   a. Detect file type
   b. Apply appropriate detector
   c. Collect variable usage
   ↓
4. Analyze Dockerfiles
   ↓
5. Compare sets (diff)
   ↓
6. Generate report
   ↓
7. Exit with appropriate code
```

## Error Handling

- **Configuration errors** → Exit code 2
- **File not found** → Exit code 2
- **Permission errors** → Exit code 2
- **Mismatches found** → Exit code 1
- **No mismatches** → Exit code 0

Soft errors (file processing failures) are logged but don't stop execution.

## Testing Strategy

### Unit Tests
- Each package has `*_test.go` files
- Table-driven tests for detectors
- Mock interfaces for integration

### Integration Tests
- `cmd/capture/*_test.go` for CLI testing
- `testdata/` for fixtures

### Coverage
- Target: >80% coverage
- Run: `make test-coverage`

## Extension Points

### Adding a New Language

1. Create detector in `internal/detector/`
2. Implement `LanguageDetector` interface
3. Add to `DetectorFactory`
4. Add file extension to `FileWalker`
5. Write tests
6. Update documentation

### Adding New Output Format

1. Create formatter in `internal/reporter/`
2. Implement formatting logic
3. Add CLI flag
4. Update tests

### Adding New Analysis Type

1. Create package in `internal/`
2. Define interfaces in `internal/types/`
3. Integrate into scan workflow
4. Add tests

## Performance Considerations

- **Streaming** - Files processed one at a time
- **Regex compilation** - Patterns compiled once
- **Memory** - No full file buffering
- **Determinism** - Sorted output for consistency

## Future Improvements

See [ROADMAP.md](../ROADMAP.md) for planned features.
