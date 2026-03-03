#!/bin/bash

# Setup script for git hooks and pre-commit
# This script installs pre-commit and sets up git hooks

set -e

echo "Setting up git hooks for capture..."
echo ""

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo "pre-commit is not installed. Installing..."
    echo ""
    
    # Detect OS and install accordingly
    if [ -f /etc/arch-release ]; then
        # Arch Linux
        echo "Detected Arch Linux"
        echo "Installing pre-commit via pacman..."
        echo ""
        echo "Please run:"
        echo "  sudo pacman -S python-pre-commit"
        echo ""
        echo "Then run this script again."
        exit 1
    elif command -v pipx &> /dev/null; then
        # pipx is available
        echo "Installing with pipx..."
        pipx install pre-commit
    elif command -v pip3 &> /dev/null; then
        # Try pip3 with --user
        echo "Installing with pip3..."
        if pip3 install --user pre-commit 2>/dev/null; then
            echo "✓ Installed with pip3"
        else
            # Externally managed environment
            echo ""
            echo "Your Python environment is externally managed."
            echo ""
            echo "Please install pre-commit using one of these methods:"
            echo ""
            echo "1. Using pipx (recommended):"
            echo "   pip3 install --user pipx"
            echo "   pipx install pre-commit"
            echo ""
            echo "2. Using system package manager:"
            echo "   # Arch Linux"
            echo "   sudo pacman -S python-pre-commit"
            echo ""
            echo "   # Ubuntu/Debian"
            echo "   sudo apt install pre-commit"
            echo ""
            echo "   # macOS"
            echo "   brew install pre-commit"
            echo ""
            echo "Then run this script again."
            exit 1
        fi
    elif command -v pip &> /dev/null; then
        # Try pip with --user
        echo "Installing with pip..."
        if pip install --user pre-commit 2>/dev/null; then
            echo "✓ Installed with pip"
        else
            echo "Failed to install with pip. See instructions above."
            exit 1
        fi
    else
        echo "Error: No Python package manager found."
        echo ""
        echo "Please install pre-commit manually:"
        echo "  https://pre-commit.com/#install"
        exit 1
    fi
    
    # Add user bin to PATH if not already there
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo ""
        echo "⚠️  Note: You may need to add ~/.local/bin to your PATH"
        echo "Add this to your ~/.bashrc or ~/.zshrc:"
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        echo ""
    fi
else
    echo "✓ pre-commit is already installed"
fi

# Verify pre-commit is accessible
if ! command -v pre-commit &> /dev/null; then
    echo ""
    echo "❌ pre-commit was installed but is not in PATH"
    echo ""
    echo "Please add the installation directory to your PATH"
    echo "Common locations:"
    echo "  ~/.local/bin (pip --user)"
    echo "  ~/.local/pipx/venvs/pre-commit/bin (pipx)"
    echo ""
    echo "Add to your ~/.bashrc or ~/.zshrc:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "Then run this script again."
    exit 1
fi

# Install pre-commit hooks using config from .github/config/
echo ""
echo "Installing pre-commit hooks..."
pre-commit install --hook-type pre-commit --config .github/config/pre-commit.yaml
pre-commit install --hook-type commit-msg --config .github/config/pre-commit.yaml

echo ""
echo "✓ Git hooks installed successfully!"
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
echo "To run hooks manually: pre-commit run --all-files --config .github/config/pre-commit.yaml"
