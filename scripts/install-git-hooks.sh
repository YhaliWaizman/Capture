#!/bin/bash

# Simple git hooks installation without pre-commit dependency
# This creates basic hooks for commit message validation

set -e

echo "Installing git hooks..."
echo ""

# Create commit-msg hook
cat > .git/hooks/commit-msg << 'EOF'
#!/bin/bash

# Validate commit message format
COMMIT_MSG_FILE=$1
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# Skip merge commits
if echo "$COMMIT_MSG" | grep -qE "^Merge "; then
    exit 0
fi

# Regex for conventional commits
PATTERN="^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?: .{1,}"

if ! echo "$COMMIT_MSG" | grep -qE "$PATTERN"; then
    echo "❌ Invalid commit message format!"
    echo ""
    echo "Commit message must follow Conventional Commits format:"
    echo "  <type>(<scope>): <subject>"
    echo ""
    echo "Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert"
    echo ""
    echo "Examples:"
    echo "  feat: add JSON output format"
    echo "  fix(parser): handle quoted values"
    echo "  docs: update README"
    echo ""
    echo "Your message:"
    echo "  $COMMIT_MSG"
    echo ""
    exit 1
fi

exit 0
EOF

chmod +x .git/hooks/commit-msg

# Create pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

# Run go fmt
echo "Running go fmt..."
if ! go fmt ./...; then
    echo "❌ go fmt failed"
    exit 1
fi

# Run tests
echo "Running tests..."
if ! go test ./...; then
    echo "❌ Tests failed"
    exit 1
fi

echo "✓ All checks passed"
exit 0
EOF

chmod +x .git/hooks/pre-commit

echo "✓ Git hooks installed successfully!"
echo ""
echo "Hooks installed:"
echo "  - commit-msg: Validates conventional commit format"
echo "  - pre-commit: Runs go fmt and tests"
echo ""
echo "Commit message format:"
echo "  <type>(<scope>): <subject>"
echo ""
echo "Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore"
echo ""
echo "Examples:"
echo "  feat: add JSON output format"
echo "  fix: correct variable detection in Python files"
echo "  docs: update README with new examples"
echo ""
echo "To skip hooks (not recommended): git commit --no-verify"
