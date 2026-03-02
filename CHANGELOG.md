# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - =2026-03-02

### Added
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

### Features
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

[Unreleased]: https://github.com/yhaliwaizman/capture/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yhaliwaizman/capture/releases/tag/v1.0.0
