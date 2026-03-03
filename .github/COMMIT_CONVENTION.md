# Commit Message Convention

Quick reference for conventional commits.

## Format

```
<type>(<scope>): <subject>
```

## Types

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat` | New feature | Minor (1.0.0 → 1.1.0) |
| `fix` | Bug fix | Patch (1.0.0 → 1.0.1) |
| `docs` | Documentation | None |
| `style` | Formatting | None |
| `refactor` | Code refactoring | None |
| `perf` | Performance | Patch |
| `test` | Tests | None |
| `build` | Build system | None |
| `ci` | CI config | None |
| `chore` | Maintenance | None |

## Breaking Changes

Add `!` after type or `BREAKING CHANGE:` in footer → Major version bump (1.0.0 → 2.0.0)

## Examples

```bash
# Features
feat: add JSON output format
feat(detector): add Ruby language support
feat(cli): add --format flag

# Bug fixes
fix: correct variable detection in Python
fix(parser): handle quoted values in .env
fix(docker): parse multi-stage Dockerfiles

# Documentation
docs: update README with examples
docs(api): add JSDoc comments

# Performance
perf: implement parallel file processing
perf(walker): optimize directory traversal

# Breaking changes
feat!: change CLI flag names
feat: rename --env-file to --env

BREAKING CHANGE: --env-file flag renamed to --env
```

## Scopes (Optional)

- `detector` - Language detectors
- `parser` - File parsers
- `reporter` - Output formatting
- `docker` - Docker features
- `cli` - Command-line interface
- `ci` - CI/CD
- `test` - Testing

## Tips

✅ **Do:**
- Use lowercase for type
- Use imperative mood ("add" not "added")
- Keep subject under 50 characters
- Separate subject from body with blank line

❌ **Don't:**
- End subject with period
- Use past tense
- Include issue numbers in subject (use footer)

## Full Example

```
feat(detector): add Ruby language support

Implement Ruby detector to identify ENV['VAR'] and ENV.fetch('VAR')
patterns in Ruby files. Supports .rb, .rake, and Gemfile extensions.

Closes #123
```
