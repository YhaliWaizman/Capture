# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.1/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0](https://github.com/YhaliWaizman/Capture/compare/v1.1.1...v2.0.0) (2026-07-23)


### ⚠ BREAKING CHANGES

* **release:** YIPPEE

### Features

* **ci:** add official GitHub Action for capture scan ([5829a8b](https://github.com/YhaliWaizman/Capture/commit/5829a8b895e0b17f693cb8948233dd5840bfda71))
* **detector:** add Java and Kotlin env detection ([7024703](https://github.com/YhaliWaizman/Capture/commit/702470385378aad6f6bc5054787e3b51c8324d41))
* **detector:** add PHP language support ([dd2539e](https://github.com/YhaliWaizman/Capture/commit/dd2539e3d725032663c0d9527b759550ae53efa7)), closes [#9](https://github.com/YhaliWaizman/Capture/issues/9)
* **detector:** add Ruby language support ([fc36058](https://github.com/YhaliWaizman/Capture/commit/fc3605890e5703706dba2c00be851a9be4ced294))
* **docker:** add Docker Compose support for env validation ([73358bd](https://github.com/YhaliWaizman/Capture/commit/73358bdea563112bdc253b089e9fbf0c1d1726e8))
* **scan:** add auto-fix mode for missing env vars ([f57e140](https://github.com/YhaliWaizman/Capture/commit/f57e1407d55d26088f0806ff09736f74981fb4c2))
* **scan:** add configuration file support ([eb1d8e5](https://github.com/YhaliWaizman/Capture/commit/eb1d8e577e9cce8ac9be992f457102ff96c256ae))
* **scan:** add incremental scanning with cache ([f0f6e63](https://github.com/YhaliWaizman/Capture/commit/f0f6e630ad9807b4ac074e390c50b2f7fc7af4e0))
* **scan:** add parallel file processing with workers ([a663855](https://github.com/YhaliWaizman/Capture/commit/a66385589ae8dceccf18a52b8d7a9ca0bbc0fabf))
* **scan:** add watch mode for automatic re-scans ([6f1ab1f](https://github.com/YhaliWaizman/Capture/commit/6f1ab1f3b6ac0a0fe9868a7a3982688cbfe372d5))
* **scan:** support multiple env files with last file wins precedence ([c8468fa](https://github.com/YhaliWaizman/Capture/commit/c8468fa509bfa52ff497fd355c0e9ef52764316b))
* **scan:** support multiple env files with last file wins precedence ([5d93057](https://github.com/YhaliWaizman/Capture/commit/5d93057c831837db8bba8bf43134317d603a5643))
* **security:** detect hardcoded secrets in source files ([0362069](https://github.com/YhaliWaizman/Capture/commit/036206932c2f60682d74424356ffa97852669dc0))
* **template:** generate .env.example from detected env usage ([fc42052](https://github.com/YhaliWaizman/Capture/commit/fc420525b43f8930d3ec08b2518cb335a0b5ce12))


### Miscellaneous Chores

* **release:** first major good release ([66d296f](https://github.com/YhaliWaizman/Capture/commit/66d296f23d4884c3dc8f74d74fb5a9312f89e658))

## [1.7.0](https://github.com/YhaliWaizman/Capture/compare/v1.6.0...v1.7.0) (2026-07-22)

### Features

* **developer-experience:** add `template` command to generate grouped, sorted `.env.example` files from detected environment variable usage with configurable `--root`, `--output`, and `--ignore`
* **security:** detect possible hardcoded secrets (API keys, tokens, private keys, and hardcoded credentials) with file/line reporting and environment-variable migration suggestions

## [1.6.0](https://github.com/YhaliWaizman/Capture/compare/v1.5.0...v1.6.0) (2026-07-22)

### Features

* **developer-experience:** add `--fix` mode to append missing env vars to `.env` with backup creation, optional confirmation bypass (`--yes`), and preview mode (`--dry-run`)
* **developer-experience:** add `--watch` mode to continuously re-run scans on file changes with 500ms debounce, terminal clear between runs, and timestamped watch status messages

## [1.5.0](https://github.com/YhaliWaizman/Capture/compare/v1.4.0...v1.5.0) (2026-07-22)

### Features

* **performance:** add parallel source-file processing with configurable `--workers` for faster scans on multi-core systems
* **performance:** add incremental scanning with `--incremental` and git-aware cache at `.capture/cache.json`, plus `--no-cache` to force full scans

## [1.4.0](https://github.com/YhaliWaizman/Capture/compare/v1.3.0...v1.4.0) (2026-07-22)

### Features

* **docker:** add Docker Compose support for `docker-compose*.yml` and `compose.yml` with environment/substitution checks and `env_file` validation

## [1.3.0](https://github.com/YhaliWaizman/Capture/compare/v1.1.1...v1.3.0) (2026-07-22)

### Features

* **detector:** add PHP language support (`$_ENV`, `$_SERVER`, `getenv`) for `.php` files
* **detector:** add Java/Kotlin language support (`System.getenv`, `System.getenv().get`, `System.getenv()[...]`) for `.java`, `.kt`, and `.kts` files

## [1.1.1](https://github.com/YhaliWaizman/Capture/compare/v1.1.0...v1.1.1) (2026-03-15)


### Bug Fixes

* **scan:** add SARIF 2.1.0 output format support ([5ce3881](https://github.com/YhaliWaizman/Capture/commit/5ce3881761b624863fac5ff53934f1e24ce277b5))
* **scan:** add SARIF 2.1.0 output format support ([c24ab6c](https://github.com/YhaliWaizman/Capture/commit/c24ab6c6668f9d23fe72ab740036ae58cb03508b))

## [1.1.0](https://github.com/YhaliWaizman/Capture/compare/v1.0.4...v1.1.0) (2026-03-11)


### Features

* **scan:** add JSON output format support ([5f7c2da](https://github.com/YhaliWaizman/Capture/commit/5f7c2dabda3b8d63d9447343568339b07e32d314))
* **scan:** add JSON output format support ([feb938e](https://github.com/YhaliWaizman/Capture/commit/feb938e34043978ddd3a04681dbc5270ec1d957f))


### Bug Fixes

* correct -dir flag to --dir in GitLab CI README example ([2556edf](https://github.com/YhaliWaizman/Capture/commit/2556edf1a3a835adc23a5aa4c46a92da5be3d284))

## [1.0.4](https://github.com/YhaliWaizman/Capture/compare/v1.0.3...v1.0.4) (2026-03-10)


### Bug Fixes

* typo ([b03fc88](https://github.com/YhaliWaizman/Capture/commit/b03fc8823a936591767071b285442abd5991cd58))
* typo ([07407b6](https://github.com/YhaliWaizman/Capture/commit/07407b624e5b50f31e945298ed140305f21e9298))

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
