#!/bin/bash

# Validate commit message format
# This script checks if a commit message follows conventional commits format

COMMIT_MSG_FILE=$1
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

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

echo "✓ Commit message format is valid"
exit 0
