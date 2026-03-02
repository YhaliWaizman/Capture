package walker

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestFileWalker_Walk(t *testing.T) {
	// Create a temporary test directory structure
	tmpDir := t.TempDir()

	// Create test directory structure
	dirs := []string{
		"src",
		"src/components",
		".git",
		"node_modules",
		"vendor",
		"custom_ignore",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create test files
	files := map[string]string{
		"src/app.js":               "// JavaScript file",
		"src/app.ts":               "// TypeScript file",
		"src/main.go":              "// Go file",
		"src/script.py":            "// Python file",
		"src/components/button.js": "// Component file",
		"src/components/button.ts": "// Component file",
		"src/readme.md":            "# Markdown file",
		"src/config.json":          "{}",
		".git/config":              "git config",
		"node_modules/package.js":  "// Should be ignored",
		"vendor/lib.go":            "// Should be ignored",
		"custom_ignore/test.js":    "// Should be ignored with custom flag",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		err := os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	t.Run("Basic directory traversal", func(t *testing.T) {
		walker := NewFileWalker()
		result, err := walker.Walk(tmpDir, nil)
		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Sort for consistent comparison
		sort.Strings(result)

		// Expected files (should include .js, .ts, .go, .py but not files in ignored dirs)
		expected := []string{
			filepath.Join(tmpDir, "src/app.js"),
			filepath.Join(tmpDir, "src/app.ts"),
			filepath.Join(tmpDir, "src/components/button.js"),
			filepath.Join(tmpDir, "src/components/button.ts"),
			filepath.Join(tmpDir, "src/main.go"),
			filepath.Join(tmpDir, "src/script.py"),
			filepath.Join(tmpDir, "custom_ignore/test.js"),
		}
		sort.Strings(expected)

		if len(result) != len(expected) {
			t.Errorf("Expected %d files, got %d", len(expected), len(result))
			t.Logf("Expected: %v", expected)
			t.Logf("Got: %v", result)
		}

		for i, exp := range expected {
			if i >= len(result) || result[i] != exp {
				t.Errorf("Expected file %s, got %s", exp, result[i])
			}
		}
	})

	t.Run("Extension filtering", func(t *testing.T) {
		walker := NewFileWalker()
		result, err := walker.Walk(tmpDir, nil)
		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Verify only valid extensions are included
		validExts := map[string]bool{".js": true, ".ts": true, ".go": true, ".py": true}
		for _, file := range result {
			ext := filepath.Ext(file)
			if !validExts[ext] {
				t.Errorf("File with invalid extension found: %s", file)
			}
		}

		// Verify .md and .json files are not included
		for _, file := range result {
			if filepath.Ext(file) == ".md" || filepath.Ext(file) == ".json" {
				t.Errorf("File with wrong extension should not be included: %s", file)
			}
		}
	})

	t.Run("Default ignore patterns", func(t *testing.T) {
		walker := NewFileWalker()
		result, err := walker.Walk(tmpDir, nil)
		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Verify files in .git, node_modules, vendor are not included
		for _, file := range result {
			if strings.Contains(file, ".git") ||
				strings.Contains(file, "node_modules") ||
				strings.Contains(file, "vendor") {
				t.Errorf("File in ignored directory should not be included: %s", file)
			}
		}
	})

	t.Run("Custom ignore patterns", func(t *testing.T) {
		walker := NewFileWalker()
		result, err := walker.Walk(tmpDir, []string{"custom_ignore"})
		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Verify files in custom_ignore are not included
		for _, file := range result {
			if strings.Contains(file, "custom_ignore") {
				t.Errorf("File in custom ignored directory should not be included: %s", file)
			}
		}

		// Verify we still get other files
		if len(result) == 0 {
			t.Error("Expected to find files outside custom_ignore directory")
		}
	})

	t.Run("Symbolic link handling", func(t *testing.T) {
		// Create a symbolic link
		linkTarget := filepath.Join(tmpDir, "src/app.js")
		linkPath := filepath.Join(tmpDir, "src/link.js")

		err := os.Symlink(linkTarget, linkPath)
		if err != nil {
			t.Skipf("Cannot create symbolic link (may require permissions): %v", err)
		}

		walker := NewFileWalker()
		result, err := walker.Walk(tmpDir, nil)
		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Verify symbolic link is not included
		for _, file := range result {
			if file == linkPath {
				t.Errorf("Symbolic link should not be included: %s", file)
			}
		}
	})

	t.Run("Case-sensitive extension matching", func(t *testing.T) {
		// Create files with uppercase extensions
		upperFiles := map[string]string{
			"src/test.JS": "// Should not match",
			"src/test.TS": "// Should not match",
			"src/test.GO": "// Should not match",
			"src/test.PY": "// Should not match",
		}

		for path, content := range upperFiles {
			fullPath := filepath.Join(tmpDir, path)
			err := os.WriteFile(fullPath, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", path, err)
			}
		}

		walker := NewFileWalker()
		result, err := walker.Walk(tmpDir, nil)
		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Verify uppercase extensions are not included
		for _, file := range result {
			ext := filepath.Ext(file)
			if ext == ".JS" || ext == ".TS" || ext == ".GO" || ext == ".PY" {
				t.Errorf("File with uppercase extension should not be included: %s", file)
			}
		}
	})
}

func TestFileWalker_NonExistentDirectory(t *testing.T) {
	walker := NewFileWalker()
	_, err := walker.Walk("/nonexistent/directory", nil)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestFileWalker_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	walker := NewFileWalker()
	result, err := walker.Walk(tmpDir, nil)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(result))
	}
}
