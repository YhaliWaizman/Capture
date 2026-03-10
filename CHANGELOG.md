# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.1/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.3](https://github.com/YhaliWaizman/Capture/compare/v1.0.2...v1.0.3) (2026-03-10)


### Bug Fixes

* correct key name from 'folder' to 'directory' in goreleaser conf… ([eb1f5b0](https://github.com/YhaliWaizman/Capture/commit/eb1f5b07278f2a7326044f824cb64964476b114c))
* correct key name from 'folder' to 'directory' in goreleaser configuration ([9a3ae1a](https://github.com/YhaliWaizman/Capture/commit/9a3ae1aa69899991ba60e438b1a13489580d845a))

## [1.0.2](https://github.com/YhaliWaizman/Capture/compare/v1.0.1...v1.0.2) (2026-03-09)


### Bug Fixes

* add version to releaser ([216a14a](https://github.com/YhaliWaizman/Capture/commit/216a14a1dde558ee0f950032efa153920ec295af))
* add version to releaser ([9c50427](https://github.com/YhaliWaizman/Capture/commit/9c50427480d0f0e5e78a93fcb84457c3c71d1701))

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
