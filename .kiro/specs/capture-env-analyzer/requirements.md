# Requirements Document

## Introduction

capture is a static analysis CLI tool that identifies mismatches between environment variables declared in .env files and those referenced in source code. The tool performs pattern-based detection across multiple programming languages and produces deterministic, CI-compatible reports.

## Glossary

- **CLI**: The command-line interface component that accepts user input and coordinates execution
- **Env_Parser**: The component that reads and extracts variable names from .env files
- **File_Walker**: The component that recursively traverses the project directory
- **Language_Detector**: The component that identifies environment variable usage patterns in source code
- **Diff_Engine**: The component that compares declared and used variable sets
- **Reporter**: The component that formats and outputs the analysis results
- **Declared_Variable**: An environment variable name extracted from the .env file
- **Used_Variable**: An environment variable name detected in source code
- **Location**: A file path and line number where a variable is used
- **Mismatch**: A situation where declared and used variable sets differ

## Requirements

### Requirement 1: CLI Command Interface

**User Story:** As a developer, I want to invoke the tool with clear parameters, so that I can analyze my project's environment variable usage.

#### Acceptance Criteria

1. THE CLI SHALL accept a command named "scan"
2. WHEN the scan command is invoked, THE CLI SHALL require a --dir flag with a directory path value
3. WHEN the scan command is invoked, THE CLI SHALL require an --env-file flag with a file path value
4. THE CLI SHALL accept an optional --ignore flag with comma-separated directory names
5. WHEN required flags are missing, THE CLI SHALL exit with code 2 and output an error message to stderr

### Requirement 2: Exit Code Signaling

**User Story:** As a CI/CD engineer, I want the tool to use standard exit codes, so that I can integrate it into automated pipelines.

#### Acceptance Criteria

1. WHEN no mismatches are detected, THE CLI SHALL exit with code 0
2. WHEN declared and used variables differ, THE CLI SHALL exit with code 1
3. WHEN the .env file does not exist, THE CLI SHALL exit with code 2
4. WHEN the scan directory does not exist, THE CLI SHALL exit with code 2
5. WHEN a permission error occurs, THE CLI SHALL exit with code 2

### Requirement 3: Parse Environment Variable Declarations

**User Story:** As a developer, I want the tool to extract variable names from my .env file, so that I can compare them against code usage.

#### Acceptance Criteria

1. WHEN a .env file is provided, THE Env_Parser SHALL read the file line-by-line
2. WHEN a line is empty, THE Env_Parser SHALL skip it
3. WHEN a line starts with #, THE Env_Parser SHALL skip it
4. WHEN a line matches the pattern KEY=VALUE, THE Env_Parser SHALL extract KEY
5. THE Env_Parser SHALL trim whitespace from extracted keys
6. WHEN a key matches the pattern ^[A-Z][A-Z0-9_]*$, THE Env_Parser SHALL include it in the declared set
7. WHEN a key does not match the pattern ^[A-Z][A-Z0-9_]*$, THE Env_Parser SHALL skip it silently
8. THE Env_Parser SHALL produce a unique set of declared variable names
9. FOR ALL valid .env files, parsing the file twice SHALL produce identical sets (idempotence property)

### Requirement 4: Traverse Project Directory

**User Story:** As a developer, I want the tool to scan my entire project, so that all environment variable usage is detected.

#### Acceptance Criteria

1. WHEN a scan directory is provided, THE File_Walker SHALL recursively traverse all subdirectories
2. THE File_Walker SHALL NOT follow symbolic links
3. THE File_Walker SHALL skip directories named .git
4. THE File_Walker SHALL skip directories named node_modules
5. THE File_Walker SHALL skip directories named vendor
6. WHEN the --ignore flag contains directory names, THE File_Walker SHALL skip directories matching those names
7. THE File_Walker SHALL match directory names exactly without glob expansion
8. WHEN a file has extension .js, THE File_Walker SHALL include it for analysis
9. WHEN a file has extension .ts, THE File_Walker SHALL include it for analysis
10. WHEN a file has extension .go, THE File_Walker SHALL include it for analysis
11. WHEN a file has extension .py, THE File_Walker SHALL include it for analysis
12. THE File_Walker SHALL perform case-sensitive extension matching

### Requirement 5: Detect JavaScript and TypeScript Usage

**User Story:** As a JavaScript developer, I want the tool to detect process.env usage, so that I can identify missing variables.

#### Acceptance Criteria

1. WHEN a .js or .ts file contains process.env.VAR_NAME, THE Language_Detector SHALL extract VAR_NAME
2. WHEN a .js or .ts file contains process.env["VAR_NAME"], THE Language_Detector SHALL extract VAR_NAME
3. WHEN a .js or .ts file contains process.env['VAR_NAME'], THE Language_Detector SHALL extract VAR_NAME
4. WHEN a variable name matches ^[A-Z][A-Z0-9_]*$, THE Language_Detector SHALL record it with file path and line number
5. WHEN a .js or .ts file contains process.env[varName] with a non-literal argument, THE Language_Detector SHALL skip it
6. WHEN a .js or .ts file contains process.env[\`VAR_${x}\`] with template interpolation, THE Language_Detector SHALL skip it

### Requirement 6: Detect Go Usage

**User Story:** As a Go developer, I want the tool to detect os.Getenv usage, so that I can identify missing variables.

#### Acceptance Criteria

1. WHEN a .go file contains os.Getenv("VAR_NAME"), THE Language_Detector SHALL extract VAR_NAME
2. WHEN a .go file contains os.LookupEnv("VAR_NAME"), THE Language_Detector SHALL extract VAR_NAME
3. WHEN a variable name matches ^[A-Z][A-Z0-9_]*$, THE Language_Detector SHALL record it with file path and line number
4. WHEN a .go file contains os.Getenv(varName) with a non-literal argument, THE Language_Detector SHALL skip it

### Requirement 7: Detect Python Usage

**User Story:** As a Python developer, I want the tool to detect os.getenv usage, so that I can identify missing variables.

#### Acceptance Criteria

1. WHEN a .py file contains os.getenv("VAR_NAME"), THE Language_Detector SHALL extract VAR_NAME
2. WHEN a .py file contains os.environ["VAR_NAME"], THE Language_Detector SHALL extract VAR_NAME
3. WHEN a .py file contains os.environ['VAR_NAME'], THE Language_Detector SHALL extract VAR_NAME
4. WHEN a variable name matches ^[A-Z][A-Z0-9_]*$, THE Language_Detector SHALL record it with file path and line number
5. WHEN a .py file contains os.getenv(var_name) with a non-literal argument, THE Language_Detector SHALL skip it

### Requirement 8: Record Usage Locations

**User Story:** As a developer, I want to see where each variable is used, so that I can quickly locate references.

#### Acceptance Criteria

1. WHEN a variable is detected, THE Language_Detector SHALL record the file path relative to the scan directory
2. WHEN a variable is detected, THE Language_Detector SHALL record the line number
3. WHEN a variable appears multiple times, THE Language_Detector SHALL record all locations
4. THE Language_Detector SHALL produce a map from variable names to location lists
5. FOR ALL location lists, THE Language_Detector SHALL sort locations by file path then by line number

### Requirement 9: Compare Variable Sets

**User Story:** As a developer, I want to know which variables are mismatched, so that I can fix configuration issues.

#### Acceptance Criteria

1. WHEN declared and used variable sets are provided, THE Diff_Engine SHALL compute unused variables as declared minus used
2. WHEN declared and used variable sets are provided, THE Diff_Engine SHALL compute missing variables as used minus declared
3. THE Diff_Engine SHALL sort unused variables alphabetically
4. THE Diff_Engine SHALL sort missing variables alphabetically
5. THE Diff_Engine SHALL perform no input/output operations
6. FOR ALL input sets, running the comparison twice SHALL produce identical results (idempotence property)

### Requirement 10: Generate Deterministic Reports

**User Story:** As a CI/CD engineer, I want consistent output format, so that I can reliably parse results.

#### Acceptance Criteria

1. WHEN unused variables exist, THE Reporter SHALL output a section titled "Declared but unused:"
2. WHEN unused variables exist, THE Reporter SHALL list each variable on a separate line prefixed with "- "
3. WHEN missing variables exist, THE Reporter SHALL output a section titled "Used but not declared:"
4. WHEN missing variables exist, THE Reporter SHALL list each variable with its first location in the format "- VAR_NAME (path/file.ext:line)"
5. WHEN both unused and missing variables exist, THE Reporter SHALL separate sections with one blank line
6. WHEN no mismatches exist, THE Reporter SHALL output "No environment mismatches found."
7. THE Reporter SHALL write analysis results to stdout
8. WHEN soft file read errors occur, THE Reporter SHALL write warnings to stderr
9. FOR ALL identical input data, THE Reporter SHALL produce identical output (determinism property)

### Requirement 11: Stream File Processing

**User Story:** As a developer working on large codebases, I want efficient memory usage, so that the tool can analyze large projects.

#### Acceptance Criteria

1. WHEN processing source files, THE Language_Detector SHALL read files line-by-line
2. THE Language_Detector SHALL NOT load entire file contents into memory simultaneously
3. THE Language_Detector SHALL compile regular expressions once per detector instance
4. THE File_Walker SHALL process files sequentially without parallelism
5. THE CLI SHALL execute with time complexity O(total file size)

### Requirement 12: Modular Architecture

**User Story:** As a maintainer, I want clear separation of concerns, so that I can extend and test the tool easily.

#### Acceptance Criteria

1. THE CLI SHALL coordinate execution without containing parsing logic
2. THE Env_Parser SHALL contain no file traversal logic
3. THE File_Walker SHALL contain no language detection logic
4. THE Language_Detector SHALL implement a common interface for extensibility
5. THE Diff_Engine SHALL contain no input/output logic
6. THE Reporter SHALL contain no comparison logic

### Requirement 13: Unit Test Coverage

**User Story:** As a maintainer, I want comprehensive tests, so that I can confidently modify the codebase.

#### Acceptance Criteria

1. THE test suite SHALL include unit tests for .env parsing with valid and malformed input
2. THE test suite SHALL include unit tests for diff logic with various set combinations
3. THE test suite SHALL include regex detection tests for JavaScript patterns
4. THE test suite SHALL include regex detection tests for TypeScript patterns
5. THE test suite SHALL include regex detection tests for Go patterns
6. THE test suite SHALL include regex detection tests for Python patterns
7. THE test suite SHALL include an integration test using a fixture project in testdata/
8. FOR ALL language detectors, THE test suite SHALL verify that dynamic expressions are ignored
