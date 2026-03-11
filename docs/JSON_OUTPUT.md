# JSON Output Format

The `--format json` flag produces machine-readable JSON output for CI/CD integration and programmatic access.

## Usage

```bash
capture scan --dir . --env-file .env --format json
```

## Output Structure

The JSON output follows this schema:

```json
{
  "summary": {
    "files_scanned": 42,
    "variables_declared": 15,
    "variables_used": 18,
    "mismatches_found": 3
  },
  "unused": ["OLD_API_KEY", "DEPRECATED_URL"],
  "missing": [
    {
      "variable": "DATABASE_URL",
      "locations": [
        {"FilePath": "src/db.go", "LineNumber": 15},
        {"FilePath": "src/db.go", "LineNumber": 20}
      ]
    }
  ],
  "dockerfile_issues": {
    "code_uses_not_in_docker": [
      {
        "variable": "MISSING_VAR",
        "locations": [
          {"FilePath": "src/app.js", "LineNumber": 10}
        ]
      }
    ],
    "docker_declares_unused": ["BUILD_VERSION"],
    "docker_uses_undeclared": [
      {
        "variable": "UNDEFINED_VAR",
        "location": {"FilePath": "Dockerfile", "LineNumber": 15}
      }
    ]
  }
}
```

## Schema Details

### Summary Object

Contains scan statistics:

- `files_scanned` (int): Total number of files analyzed (source files + Dockerfiles)
- `variables_declared` (int): Number of variables declared in .env file
- `variables_used` (int): Number of unique variables referenced in source code
- `mismatches_found` (int): Total count of all mismatches (unused + missing + dockerfile issues)

### Unused Array

List of variable names declared in .env but not used in source code.

```json
"unused": ["OLD_API_KEY", "DEPRECATED_URL"]
```

### Missing Array

Variables used in code but not declared in .env. Each entry includes:

- `variable` (string): Variable name
- `locations` (array): All locations where the variable is used
  - `FilePath` (string): Relative path to the file
  - `LineNumber` (int): Line number (1-indexed)

```json
"missing": [
  {
    "variable": "DATABASE_URL",
    "locations": [
      {"FilePath": "src/db.go", "LineNumber": 15},
      {"FilePath": "src/db.go", "LineNumber": 20}
    ]
  }
]
```

### Dockerfile Issues Object

Contains three types of Dockerfile-specific mismatches:

#### code_uses_not_in_docker

Variables used in source code but not declared in Dockerfile or .env:

```json
"code_uses_not_in_docker": [
  {
    "variable": "MISSING_VAR",
    "locations": [
      {"FilePath": "src/app.js", "LineNumber": 10}
    ]
  }
]
```

#### docker_declares_unused

Variables declared in Dockerfile but not used in source code:

```json
"docker_declares_unused": ["BUILD_VERSION", "NODE_ENV"]
```

#### docker_uses_undeclared

Variables referenced in Dockerfile but not declared via ENV or ARG:

```json
"docker_uses_undeclared": [
  {
    "variable": "UNDEFINED_VAR",
    "location": {"FilePath": "Dockerfile", "LineNumber": 15}
  }
]
```

## Exit Codes

Exit codes remain the same regardless of output format:

- **0**: No mismatches detected
- **1**: Mismatches found
- **2**: Configuration error

## Examples

### Parse with jq

```bash
# Get summary statistics
capture scan --dir . --env-file .env --format json | jq '.summary'

# List all missing variables
capture scan --dir . --env-file .env --format json | jq '.missing[].variable'

# Count mismatches
capture scan --dir . --env-file .env --format json | jq '.summary.mismatches_found'

# Get all locations for a specific variable
capture scan --dir . --env-file .env --format json | \
  jq '.missing[] | select(.variable == "DATABASE_URL") | .locations'
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

print(f"Files scanned: {data['summary']['files_scanned']}")
print(f"Mismatches found: {data['summary']['mismatches_found']}")

for item in data['missing']:
    print(f"Missing: {item['variable']}")
    for loc in item['locations']:
        print(f"  - {loc['FilePath']}:{loc['LineNumber']}")
```

### CI/CD Integration

#### GitHub Actions

```yaml
name: Environment Variable Check

on: [push, pull_request]

jobs:
  check-env:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Download capture
        run: |
          curl -L https://github.com/yhaliwaizman/capture/releases/latest/download/capture_Linux_x86_64.tar.gz | tar xz
          chmod +x capture
      
      - name: Scan environment variables
        run: |
          ./capture scan --dir . --env-file .env --format json > results.json
          cat results.json | jq '.'
      
      - name: Check for mismatches
        run: |
          MISMATCHES=$(cat results.json | jq '.summary.mismatches_found')
          if [ "$MISMATCHES" -gt 0 ]; then
            echo "Found $MISMATCHES mismatches"
            exit 1
          fi
      
      - name: Upload results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: env-scan-results
          path: results.json
```

#### GitLab CI

```yaml
check-env:
  stage: test
  image: alpine:latest
  before_script:
    - apk add --no-cache curl jq
    - curl -L https://github.com/yhaliwaizman/capture/releases/latest/download/capture_Linux_x86_64.tar.gz | tar xz
    - chmod +x capture
  script:
    - ./capture scan --dir . --env-file .env --format json > results.json
    - cat results.json | jq '.summary'
    - |
      MISMATCHES=$(cat results.json | jq '.summary.mismatches_found')
      if [ "$MISMATCHES" -gt 0 ]; then
        echo "Found $MISMATCHES mismatches"
        exit 1
      fi
  artifacts:
    paths:
      - results.json
    when: always
```

## Comparison with Text Format

| Feature | Text Format | JSON Format |
|---------|-------------|-------------|
| Human-readable | ✅ Yes | ❌ No |
| Machine-parseable | ❌ Difficult | ✅ Easy |
| All locations | ❌ First only | ✅ All |
| Summary stats | ❌ No | ✅ Yes |
| CI/CD integration | ⚠️ Basic | ✅ Advanced |
| Default | ✅ Yes | ❌ No |

## Notes

- JSON output goes to stdout
- Errors and warnings still go to stderr
- Exit codes work the same way for both formats
- JSON is formatted with indentation for readability
- All arrays are sorted alphabetically for deterministic output
