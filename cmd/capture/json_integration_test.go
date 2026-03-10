package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/yhaliwaizman/capture/internal/types"
)

func TestCLI_JSONFormat(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "capture-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("capture-test")

	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create test .env file
	envContent := `API_KEY=secret
DATABASE_URL=postgres://localhost
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Create test source file
	jsContent := `const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;
const missing = process.env.MISSING_VAR;
`
	jsPath := filepath.Join(tmpDir, "app.js")
	if err := os.WriteFile(jsPath, []byte(jsContent), 0644); err != nil {
		t.Fatalf("Failed to create app.js file: %v", err)
	}

	// Run scan with JSON format
	cmd := exec.Command("./capture-test", "scan",
		"--dir", tmpDir,
		"--env-file", envPath,
		"--format", "json")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Error("Expected non-zero exit code for mismatches")
	}

	// Parse JSON output
	var output types.JSONOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, stdout.String())
	}

	// Verify summary
	if output.Summary.FilesScanned != 1 {
		t.Errorf("FilesScanned = %d, want 1", output.Summary.FilesScanned)
	}
	if output.Summary.VariablesDeclared != 2 {
		t.Errorf("VariablesDeclared = %d, want 2", output.Summary.VariablesDeclared)
	}
	if output.Summary.VariablesUsed != 3 {
		t.Errorf("VariablesUsed = %d, want 3", output.Summary.VariablesUsed)
	}
	if output.Summary.MismatchesFound < 1 {
		t.Errorf("MismatchesFound = %d, want at least 1", output.Summary.MismatchesFound)
	}

	// Verify missing variables
	if len(output.Missing) != 1 {
		t.Fatalf("Missing count = %d, want 1", len(output.Missing))
	}
	if output.Missing[0].Variable != "MISSING_VAR" {
		t.Errorf("Missing variable = %s, want MISSING_VAR", output.Missing[0].Variable)
	}
	if len(output.Missing[0].Locations) != 1 {
		t.Errorf("Missing locations count = %d, want 1", len(output.Missing[0].Locations))
	}

	// Verify unused is empty
	if len(output.Unused) != 0 {
		t.Errorf("Unused count = %d, want 0", len(output.Unused))
	}
}

func TestCLI_JSONFormatNoMismatches(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "capture-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("capture-test")

	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create test .env file
	envContent := `API_KEY=secret
DATABASE_URL=postgres://localhost
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Create test source file with no mismatches
	jsContent := `const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;
`
	jsPath := filepath.Join(tmpDir, "app.js")
	if err := os.WriteFile(jsPath, []byte(jsContent), 0644); err != nil {
		t.Fatalf("Failed to create app.js file: %v", err)
	}

	// Run scan with JSON format
	cmd := exec.Command("./capture-test", "scan",
		"--dir", tmpDir,
		"--env-file", envPath,
		"--format", "json")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Errorf("Expected zero exit code for no mismatches, got error: %v", err)
	}

	// Parse JSON output
	var output types.JSONOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, stdout.String())
	}

	// Verify no mismatches
	if output.Summary.MismatchesFound != 0 {
		t.Errorf("MismatchesFound = %d, want 0", output.Summary.MismatchesFound)
	}
	if len(output.Missing) != 0 {
		t.Errorf("Missing count = %d, want 0", len(output.Missing))
	}
	if len(output.Unused) != 0 {
		t.Errorf("Unused count = %d, want 0", len(output.Unused))
	}
}

func TestCLI_InvalidFormat(t *testing.T) {
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

	// Check error message
	if !bytes.Contains(stderr.Bytes(), []byte("invalid format")) {
		t.Errorf("Expected 'invalid format' error message, got: %s", stderr.String())
	}
}

func TestCLI_DefaultFormatIsText(t *testing.T) {
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

	// Create test source file
	jsPath := filepath.Join(tmpDir, "app.js")
	if err := os.WriteFile(jsPath, []byte("const key = process.env.API_KEY;\n"), 0644); err != nil {
		t.Fatalf("Failed to create app.js file: %v", err)
	}

	// Run scan without format flag (should default to text)
	cmd := exec.Command("./capture-test", "scan",
		"--dir", tmpDir,
		"--env-file", envPath)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Errorf("Expected zero exit code, got error: %v", err)
	}

	// Verify output is text format (not JSON)
	output := stdout.String()
	if bytes.Contains(stdout.Bytes(), []byte("{")) {
		t.Error("Expected text format output, got JSON")
	}
	if !bytes.Contains(stdout.Bytes(), []byte("No environment mismatches found")) {
		t.Errorf("Expected text format message, got: %s", output)
	}
}
