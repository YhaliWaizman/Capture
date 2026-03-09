# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.1/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1](https://github.com/YhaliWaizman/Capture/compare/v1.0.0...v1.0.1) (2026-03-09)


### Bug Fixes

* add sensible defaults to flags for easier user interface ([31542e2](https://github.com/YhaliWaizman/Capture/commit/31542e2197d7d80b6a9c35c1e0e5b2ea5d70c218))
* **chore but need release tests:** consolidate release workflows and … ([37adeb8](https://github.com/YhaliWaizman/Capture/commit/37adeb863bff7e316edf8b86c6dafe0cf9bd58c3))
* **chore but need release tests:** consolidate release workflows and simplify CI pipeline ([d39efdc](https://github.com/YhaliWaizman/Capture/commit/d39efdc1a5c08e01760b506c1cee4d1042ac3f59))
* remove root flag and switch with dir flag for better readability ([dad29c9](https://github.com/YhaliWaizman/Capture/commit/dad29c94fe034f3abb2ad64a83ae147e44c7cec8))

## [1.0.1] - 2026-03-02

### Features

- Initial release of capture CLI tool
- Environment variable detection in JavaScript, TypeScript, Go, and Python
- .env file parsing and validation
- Dockerfile analysis (ENV/ARG declarations and usage)
- Cross-checking between .env, Dockerfile, and source code
- Multi-stage Dockerfile support
- Line continuation handling in Dockerfiles
- Deterministic output for CI/CD integration
- Memory-efficient streaming file processing
- Configurable directory ignore patterns
- Clear reporting with file locations
- Exit codes for CI/CD integration (0=success, 1=mismatches, 2=error)
- Pattern-based detection without AST parsing
- Support for multiple Dockerfile naming conventions (Dockerfile, Dockerfile.*, *.dockerfile)
- Case-insensitive Dockerfile instruction matching
- Uppercase-only variable name validation
- Comprehensive test coverage (75 tests)

### Documentation

- Complete README with usage examples
- Dockerfile analysis feature documentation
- Integration guides
- CI/CD integration examples

[1.0.1]: https://github.com/yhaliwaizman/capture/releases/tag/v1.0.1
