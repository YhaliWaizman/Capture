# Feature: JSON Output Format

## Overview

Added support for JSON output format to enable better CI/CD integration and programmatic access to scan results.

## Implementation

### Changes Made

1. **New Types** (`internal/types/types.go`)
   - `JSONOutput`: Complete JSON output structure
   - `Summary`: Scan statistics
   - `MissingVariable`: Variable with all locations
   - `DockerfileIssues`: Dockerfile-specific mismatches
   - `DockerUndeclaredVar`: Undeclared variable in Dockerfile
   - Extended `ReportData` to include additional fields for JSON output

2. **Reporter Enhancement** (`internal/reporter/reporter.go`)
   - Added `ReportJSON()` method to `Reporter` interface
   - Implemented JSON marshaling with proper formatting
   - Maintains alphabetical sorting for deterministic output
   - Includes all locations for each variable (not just first)

3. **Command-Line Interface** (`cmd/capture/cmd/scan.go`)
   - Added `--format` flag with `text` (default) and `json` options
   - Validation for format flag
   - Updated `executeScan()` to collect comprehensive data
   - Conditional output based on format selection

4. **Tests**
   - `internal/reporter/reporter_json_test.go`: Unit tests for JSON reporter
   - `cmd/capture/json_integration_test.go`: Integration tests for JSON output
   - All existing tests continue to pass

5. **Documentation**
   - Updated `README.md` with JSON format examples
   - Created `docs/JSON_OUTPUT.md` with comprehensive JSON documentation
   - Updated command help text

## Usage

### Basic Usage

```bash
# Text format (default)
capture scan --dir . --env-file .env

# JSON format
capture scan --dir . --env-file .env --format json
```

### JSON Output Structure

```json
{
  "summary": {
    "files_scanned": 42,
    "variables_declared": 15,
    "variables_used": 18,
    "mismatches_found": 3
  },
  "unused": ["OLD_API_KEY"],
  "missing": [
    {
      "variable": "DATABASE_URL",
      "locations": [
        {"FilePath": "src/db.go", "LineNumber": 15}
      ]
    }
  ],
  "dockerfile_issues": {
    "code_uses_not_in_docker": [],
    "docker_declares_unused": [],
    "docker_uses_undeclared": []
  }
}
```

## Features

### Summary Statistics

- `files_scanned`: Total files analyzed
- `variables_declared`: Variables in .env
- `variables_used`: Unique variables in code
- `mismatches_found`: Total mismatch count

### Complete Location Information

Unlike text format (which shows only first location), JSON includes all locations where each variable is used.

### Dockerfile Issues

Three categories of Dockerfile-specific mismatches:
1. Code uses variables not in Dockerfile or .env
2. Dockerfile declares variables unused in code
3. Dockerfile uses undeclared variables

### Deterministic Output

- All arrays sorted alphabetically
- Consistent structure
- Reliable for CI/CD parsing

## Exit Codes

Exit codes remain unchanged:
- **0**: No mismatches
- **1**: Mismatches found
- **2**: Configuration error

## CI/CD Integration Examples

### GitHub Actions

```yaml
- name: Scan environment variables
  run: |
    ./capture scan --dir . --env-file .env --format json > results.json
    MISMATCHES=$(cat results.json | jq '.summary.mismatches_found')
    if [ "$MISMATCHES" -gt 0 ]; then
      exit 1
    fi
```

### GitLab CI

```yaml
check-env:
  script:
    - ./capture scan --dir . --env-file .env --format json > results.json
    - cat results.json | jq '.summary'
  artifacts:
    paths:
      - results.json
```

### Parse with Python

```python
import json
import subprocess

result = subprocess.run(
    ['capture', 'scan', '--dir', '.', '--env-file', '.env', '--format', 'json'],
    capture_output=True,
    text=True
)

data = json.loads(result.stdout)
print(f"Mismatches: {data['summary']['mismatches_found']}")
```

## Backward Compatibility

- Default format remains `text`
- Existing scripts and workflows continue to work
- No breaking changes to existing functionality

## Testing

All tests pass:
- Unit tests for JSON reporter
- Integration tests for JSON output
- Validation of JSON structure
- Error handling for invalid format
- Default format behavior

## Future Enhancements

This JSON output format provides a foundation for:
- SARIF output format
- JUnit XML output format
- Custom report templates
- Integration with other tools

## Related Issues

- Part of v1.1.0 milestone
- Prerequisite for SARIF output
- Enables JUnit XML output
