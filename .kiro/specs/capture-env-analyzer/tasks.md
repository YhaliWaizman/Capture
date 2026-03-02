# Implementation Plan: capture-env-analyzer

## Overview

This plan implements a static analysis CLI tool that detects environment variable mismatches between .env files and source code. The implementation follows a pipeline architecture with six main components: CLI, Env_Parser, File_Walker, Language_Detector, Diff_Engine, and Reporter. The tool supports JavaScript, TypeScript, Go, and Python using pattern-based regex detection.

## Tasks

- [x] 1. Set up project structure and core types
  - Create Go project with go.mod
  - Define core types and interfaces (Location, DiffResult, ReportData)
  - Set up Go testing package
  - Create directory structure (cmd/, internal/, testdata/)
  - _Requirements: 12.1, 12.5, 12.6_

- [x] 2. Implement Env_Parser component
  - [x] 2.1 Create EnvParser interface and implementation
    - Implement line-by-line file reading
    - Extract KEY from KEY=VALUE pattern with regex /^([A-Z][A-Z0-9_]*)=/
    - Skip empty lines and comments starting with #
    - Filter keys matching ^[A-Z][A-Z0-9_]*$
    - Return unique map[string]bool of declared variables
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8_
  
  - [x] 2.2 Write unit tests for Env_Parser
    - Test valid .env files with various KEY=VALUE formats
    - Test empty lines and comment handling
    - Test invalid key patterns (lowercase, special chars)
    - Test duplicate key handling
    - Test whitespace trimming
    - _Requirements: 13.1_

- [x] 3. Implement File_Walker component
  - [x] 3.1 Create FileWalker interface and implementation
    - Implement recursive directory traversal
    - Filter by extensions: .js, .ts, .go, .py (case-sensitive)
    - Skip symbolic links
    - Skip default ignored directories: .git, node_modules, vendor
    - Support custom ignore list from --ignore flag
    - Yield file paths as channel or slice
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 4.9, 4.10, 4.11, 4.12_
  
  - [x] 3.2 Write unit tests for File_Walker
    - Test directory traversal with fixture project
    - Test extension filtering
    - Test ignore patterns
    - Test symbolic link handling
    - _Requirements: 13.7_

- [x] 4. Implement Language_Detector components
  - [x] 4.1 Create LanguageDetector interface
    - Define Detect(filePath string) map[string][]Location method
    - Define Location struct with FilePath and LineNumber fields
    - _Requirements: 12.4_
  
  - [x] 4.2 Implement JSDetector for JavaScript/TypeScript
    - Compile regex patterns: process.env.VAR, process.env["VAR"], process.env['VAR']
    - Read files line-by-line for memory efficiency
    - Extract variable names matching ^[A-Z][A-Z0-9_]*$
    - Record file path (relative to root) and line number (1-indexed)
    - Skip dynamic expressions (template literals, variable references)
    - Return map[string][]Location with sorted locations
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 8.1, 8.2, 8.3, 8.4, 8.5, 11.1, 11.2, 11.3_
  
  - [x] 4.3 Write unit tests for JSDetector
    - Test process.env.VAR_NAME pattern
    - Test process.env["VAR_NAME"] pattern
    - Test process.env['VAR_NAME'] pattern
    - Test dynamic expression rejection (template literals, variables)
    - Test line number recording
    - Test multiple occurrences of same variable
    - _Requirements: 13.3, 13.4, 13.8_
  
  - [x] 4.4 Implement GoDetector for Go
    - Compile regex patterns: os.Getenv("VAR"), os.LookupEnv("VAR")
    - Read files line-by-line for memory efficiency
    - Extract variable names matching ^[A-Z][A-Z0-9_]*$
    - Record file path and line number
    - Skip dynamic expressions (variable references)
    - Return map[string][]Location with sorted locations
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 8.1, 8.2, 8.3, 8.4, 8.5, 11.1, 11.2, 11.3_
  
  - [x] 4.5 Write unit tests for GoDetector
    - Test os.Getenv("VAR_NAME") pattern
    - Test os.LookupEnv("VAR_NAME") pattern
    - Test dynamic expression rejection
    - Test line number recording
    - _Requirements: 13.5, 13.8_
  
  - [x] 4.6 Implement PythonDetector for Python
    - Compile regex patterns: os.getenv("VAR"), os.environ["VAR"], os.environ['VAR']
    - Read files line-by-line for memory efficiency
    - Extract variable names matching ^[A-Z][A-Z0-9_]*$
    - Record file path and line number
    - Skip dynamic expressions (variable references)
    - Return map[string][]Location with sorted locations
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 8.1, 8.2, 8.3, 8.4, 8.5, 11.1, 11.2, 11.3_
  
  - [x] 4.7 Write unit tests for PythonDetector
    - Test os.getenv("VAR_NAME") pattern
    - Test os.environ["VAR_NAME"] pattern
    - Test os.environ['VAR_NAME'] pattern
    - Test dynamic expression rejection
    - Test line number recording
    - _Requirements: 13.6, 13.8_
  
  - [x] 4.8 Create DetectorFactory
    - Map file extensions to detector instances: .js/.ts → JSDetector, .go → GoDetector, .py → PythonDetector
    - Return appropriate detector based on file extension
    - _Requirements: 12.4_

- [x] 5. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Implement Diff_Engine component
  - [x] 6.1 Create DiffEngine interface and implementation
    - Implement Compare(declared map[string]bool, used map[string]bool) DiffResult
    - Compute unused as declared minus used
    - Compute missing as used minus declared
    - Sort both arrays alphabetically
    - Ensure pure computation with no I/O operations
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_
  
  - [x] 6.2 Write unit tests for Diff_Engine
    - Test empty sets
    - Test identical sets (no mismatches)
    - Test only unused variables
    - Test only missing variables
    - Test both unused and missing variables
    - Test alphabetical sorting
    - _Requirements: 13.2_

- [x] 7. Implement Reporter component
  - [x] 7.1 Create Reporter interface and implementation
    - Format "Declared but unused:" section with "- VAR_NAME" lines
    - Format "Used but not declared:" section with "- VAR_NAME (path:line)" lines
    - Add blank line separator between sections when both exist
    - Output "No environment mismatches found." when no mismatches
    - Write results to stdout
    - Write warnings/errors to stderr
    - Sort missing variables alphabetically for determinism
    - Use first location for each missing variable
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 10.7, 10.8_
  
  - [x] 7.2 Write unit tests for Reporter
    - Test output with only unused variables
    - Test output with only missing variables
    - Test output with both unused and missing variables
    - Test output with no mismatches
    - Test deterministic output (same input produces same output)
    - Test blank line separator
    - _Requirements: 10.9_

- [x] 8. Implement CLI component
  - [x] 8.1 Create CLI flag parser
    - Parse "scan" command
    - Parse --root flag (required)
    - Parse --env-file flag (required)
    - Parse --ignore flag (optional, comma-separated)
    - Validate required flags are present
    - Exit with code 2 and stderr message if flags missing
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_
  
  - [x] 8.2 Implement main execution flow
    - Validate .env file exists (exit code 2 if not)
    - Validate root directory exists (exit code 2 if not)
    - Handle permission errors (exit code 2)
    - Coordinate pipeline: Env_Parser → File_Walker → Language_Detector → Diff_Engine → Reporter
    - Exit with code 0 if no mismatches
    - Exit with code 1 if mismatches found
    - Exit with code 2 on configuration errors
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 11.5, 12.1_
  
  - [x] 8.3 Write integration tests for CLI
    - Create testdata/ fixture project with .env file and source files
    - Test successful scan with no mismatches (exit code 0)
    - Test scan with unused variables (exit code 1)
    - Test scan with missing variables (exit code 1)
    - Test missing .env file (exit code 2)
    - Test missing root directory (exit code 2)
    - Test --ignore flag functionality
    - Verify deterministic output
    - _Requirements: 13.7_

- [x] 9. Wire components together and create entry point
  - [x] 9.1 Create main entry point
    - Import all components
    - Wire CLI → Env_Parser → File_Walker → Language_Detector → Diff_Engine → Reporter
    - Handle os.Args parsing
    - Set up proper error handling and exit codes
    - _Requirements: 11.4, 12.1_
  
  - [x] 9.2 Add build configuration
    - Configure Go build target
    - Set up executable entry point in cmd/
    - Add Makefile or build scripts for build and test
    - _Requirements: 11.5_

- [x] 10. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- The implementation uses Go as specified in the design document
- Pattern-based regex detection is used instead of AST parsing for simplicity
- Line-by-line file reading ensures O(n) space complexity
- All output is deterministic for CI/CD reliability
