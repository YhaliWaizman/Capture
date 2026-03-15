# SARIF Output Format

The `--format sarif` flag produces output in [SARIF](https://sarifweb.azurewebsites.net/) (Static Analysis Results Interchange Format) 2.1.0, an OASIS standard for static analysis tool output. SARIF enables integration with GitHub Code Scanning, Azure DevOps, and other security platforms that consume standardized findings.

## Usage

```bash
capture scan --dir . --env-file .env --format sarif
```

To save results to a file:

```bash
capture scan --dir . --env-file .env --format sarif > results.sarif
```

## Rule Definitions

Each category of environment variable mismatch maps to a distinct SARIF rule. Rules are only included in the output when at least one finding exists for that category.

| Rule ID | Name | Description | Level |
|---------|------|-------------|-------|
| ENV001 | `unused-variable` | Variable is declared in .env but not used in code | `warning` |
| ENV002 | `missing-variable` | Variable is used in code but not declared in .env | `error` |
| ENV003 | `code-uses-not-in-docker` | Variable is used in code but not declared in Dockerfile or .env | `warning` |
| ENV004 | `docker-declares-unused` | Variable is declared in Dockerfile but not used in code | `warning` |
| ENV005 | `docker-uses-undeclared` | Variable is used in Dockerfile but not declared | `error` |

### Severity Levels

- **error**: The variable mismatch is likely to cause runtime failures (missing declarations)
- **warning**: The variable mismatch indicates unused or orphaned configuration

## Output Structure

The SARIF output conforms to the [SARIF 2.1.0 schema](https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json). It contains a single run with tool metadata, rule definitions, and results.

### Complete Example

The following example shows SARIF output from a scan that found all five categories of mismatches:

```json
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "capture",
          "informationUri": "https://github.com/syncd-one/syncd",
          "rules": [
            {
              "id": "ENV001",
              "name": "unused-variable",
              "shortDescription": {
                "text": "Variable is declared in .env but not used in code"
              },
              "helpUri": "https://github.com/syncd-one/syncd#unused-variable"
            },
            {
              "id": "ENV002",
              "name": "missing-variable",
              "shortDescription": {
                "text": "Variable is used in code but not declared in .env"
              },
              "helpUri": "https://github.com/syncd-one/syncd#missing-variable"
            },
            {
              "id": "ENV003",
              "name": "code-uses-not-in-docker",
              "shortDescription": {
                "text": "Variable is used in code but not declared in Dockerfile or .env"
              },
              "helpUri": "https://github.com/syncd-one/syncd#code-uses-not-in-docker"
            },
            {
              "id": "ENV004",
              "name": "docker-declares-unused",
              "shortDescription": {
                "text": "Variable is declared in Dockerfile but not used in code"
              },
              "helpUri": "https://github.com/syncd-one/syncd#docker-declares-unused"
            },
            {
              "id": "ENV005",
              "name": "docker-uses-undeclared",
              "shortDescription": {
                "text": "Variable is used in Dockerfile but not declared"
              },
              "helpUri": "https://github.com/syncd-one/syncd#docker-uses-undeclared"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "ENV001",
          "ruleIndex": 0,
          "level": "warning",
          "message": {
            "text": "UNUSED_VAR: Variable is declared in .env but not used in code"
          }
        },
        {
          "ruleId": "ENV002",
          "ruleIndex": 1,
          "level": "error",
          "message": {
            "text": "MISSING_VAR: Variable is used in code but not declared in .env"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "src/app.go"
                },
                "region": {
                  "startLine": 10
                }
              }
            }
          ]
        },
        {
          "ruleId": "ENV003",
          "ruleIndex": 2,
          "level": "warning",
          "message": {
            "text": "CODE_VAR: Variable is used in code but not declared in Dockerfile or .env"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "src/config.go"
                },
                "region": {
                  "startLine": 20
                }
              }
            }
          ]
        },
        {
          "ruleId": "ENV004",
          "ruleIndex": 3,
          "level": "warning",
          "message": {
            "text": "DOCKER_UNUSED: Variable is declared in Dockerfile but not used in code"
          }
        },
        {
          "ruleId": "ENV005",
          "ruleIndex": 4,
          "level": "error",
          "message": {
            "text": "DOCKER_UNDECL: Variable is used in Dockerfile but not declared"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "Dockerfile"
                },
                "region": {
                  "startLine": 5
                }
              }
            }
          ]
        }
      ]
    }
  ]
}
```


### Empty Results

When no mismatches are found, the output is a valid SARIF document with empty `results` and `rules` arrays:

```json
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "capture",
          "informationUri": "https://github.com/syncd-one/syncd",
          "rules": []
        }
      },
      "results": []
    }
  ]
}
```

### Key Fields

- **version**: Always `"2.1.0"`
- **$schema**: Points to the official SARIF 2.1.0 JSON schema
- **runs**: Contains exactly one run object
- **tool.driver.name**: Always `"capture"`
- **tool.driver.rules**: Only rules with at least one result are included
- **results**: Sorted by `ruleId`, then alphabetically by variable name
- **locations**: Included for ENV002, ENV003, and ENV005 results when location data is available; omitted for ENV001 and ENV004

## Location Data

Results for rules that have source location information include a `locations` array with physical location details:

- **ENV002** (missing-variable): File path and line number where the variable is used in code
- **ENV003** (code-uses-not-in-docker): File path and line number of the first usage in code
- **ENV005** (docker-uses-undeclared): Dockerfile path and line number

Results for **ENV001** (unused-variable) and **ENV004** (docker-declares-unused) omit the `locations` array since these findings relate to declarations without specific source code references.

All file paths use forward-slash separators for cross-platform compatibility.

## Deterministic Output

SARIF output is fully deterministic — identical scan inputs always produce byte-identical output. This makes it safe to use in CI/CD pipelines for diffing and caching.

## Exit Codes

Exit codes are the same regardless of output format:

- **0**: No mismatches detected
- **1**: Mismatches found
- **2**: Configuration error

## GitHub Actions Integration

Upload SARIF results to GitHub Code Scanning to see findings directly in pull requests and the Security tab:

```yaml
name: Environment Variable Scan

on: [push, pull_request]

jobs:
  capture-scan:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4

      - name: Download capture
        run: |
          curl -L https://github.com/yhaliwaizman/capture/releases/latest/download/capture_Linux_x86_64.tar.gz | tar xz
          chmod +x capture

      - name: Run capture scan
        run: ./capture scan --dir . --env-file .env --format sarif > results.sarif

      - name: Upload SARIF results
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

This workflow:
1. Checks out the repository
2. Downloads the `capture` binary
3. Runs a scan with `--format sarif` and writes output to `results.sarif`
4. Uploads the SARIF file to GitHub Code Scanning using `github/codeql-action/upload-sarif`

The `security-events: write` permission is required for the SARIF upload step.

## Comparison with Other Formats

| Feature | Text | JSON | SARIF |
|---------|------|------|-------|
| Human-readable | ✅ Yes | ❌ No | ❌ No |
| Machine-parseable | ❌ Difficult | ✅ Easy | ✅ Easy |
| GitHub Code Scanning | ❌ No | ❌ No | ✅ Yes |
| Summary statistics | ❌ No | ✅ Yes | ❌ No |
| Location data | ⚠️ First only | ✅ All | ✅ First |
| Deterministic | ✅ Yes | ✅ Yes | ✅ Yes |
| Default | ✅ Yes | ❌ No | ❌ No |

## Notes

- SARIF output goes to stdout; errors go to stderr
- The output conforms to SARIF 2.1.0 and is accepted by GitHub Code Scanning, Azure DevOps, and other SARIF-compatible tools
- No external dependencies are required — the output is generated using Go's standard `encoding/json` package
