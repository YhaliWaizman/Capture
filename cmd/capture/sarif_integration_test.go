package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yhaliwaizman/capture/internal/types"
)

// TestCLI_SARIFFormatAccepted tests that --format sarif is accepted without error
// Requirements: 1.1
func TestCLI_SARIFFormatAccepted(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "capture-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("capture-test")

	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create test .env file
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("API_KEY=secret\n"), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Create test source file that uses the variable (no mismatches)
	jsPath := filepath.Join(tmpDir, "app.js")
	if err := os.WriteFile(jsPath, []byte("const key = process.env.API_KEY;\n"), 0644); err != nil {
		t.Fatalf("Failed to create app.js file: %v", err)
	}

	// Run scan with SARIF format
	cmd := exec.Command("./capture-test", "scan",
		"--dir", tmpDir,
		"--env-file", envPath,
		"--format", "sarif")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Errorf("Expected zero exit code for --format sarif with no mismatches, got error: %v\nStderr: %s", err, stderr.String())
	}

	// Verify output is valid SARIF JSON
	var doc types.SARIFDocument
	if err := json.Unmarshal(stdout.Bytes(), &doc); err != nil {
		t.Fatalf("Failed to parse SARIF output: %v\nOutput: %s", err, stdout.String())
	}

	if doc.Version != "2.1.0" {
		t.Errorf("SARIF version = %q, want %q", doc.Version, "2.1.0")
	}
}

// TestCLI_InvalidFormatListsSARIF tests that invalid format error message includes sarif
// Requirements: 1.4
func TestCLI_InvalidFormatListsSARIF(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "capture-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("capture-test")

	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create test .env file
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("API_KEY=secret\n"), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Run scan with invalid format
	cmd := exec.Command("./capture-test", "scan",
		"--dir", tmpDir,
		"--env-file", envPath,
		"--format", "xml")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Error("Expected non-zero exit code for invalid format")
	}

	stderrOutput := stderr.String()

	// Check that error message mentions sarif as a valid format
	if !strings.Contains(stderrOutput, "sarif") {
		t.Errorf("Expected 'sarif' in error message, got: %s", stderrOutput)
	}

	// Check that error message also mentions text and json
	if !strings.Contains(stderrOutput, "text") {
		t.Errorf("Expected 'text' in error message, got: %s", stderrOutput)
	}
	if !strings.Contains(stderrOutput, "json") {
		t.Errorf("Expected 'json' in error message, got: %s", stderrOutput)
	}

	// Verify exit code is 2
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 2 {
			t.Errorf("Expected exit code 2, got %d", exitErr.ExitCode())
		}
	}
}
