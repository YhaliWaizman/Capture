package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary builds the capture binary for testing
func buildBinary(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "capture")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	return binaryPath
}

// TestCLI_SuccessfulScanNoMismatches tests a scan with no mismatches (exit code 0)
// Requirement: 13.7
func TestCLI_SuccessfulScanNoMismatches(t *testing.T) {
	binary := buildBinary(t)

	// Create a temporary test environment
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	srcFile := filepath.Join(tmpDir, "app.js")

	// Write .env file with one variable
	if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Write source file that uses the same variable
	if err := os.WriteFile(srcFile, []byte("const key = process.env.API_KEY;\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Run the CLI
	cmd := exec.Command(binary, "scan", "--dir", tmpDir, "--env-file", envFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit with code 0 (no mismatches)
	if err != nil {
		t.Errorf("Expected exit code 0, got error: %v\nStderr: %s", err, stderr.String())
	}

	// Should output "No environment mismatches found."
	output := stdout.String()
	if !strings.Contains(output, "No environment mismatches found.") {
		t.Errorf("Expected 'No environment mismatches found.' in output, got: %s", output)
	}
}

// TestCLI_ScanWithUnusedVariables tests a scan with unused variables (exit code 1)
// Requirement: 13.7
func TestCLI_ScanWithUnusedVariables(t *testing.T) {
	binary := buildBinary(t)

	// Use the testdata directory
	rootDir := "../../testdata"
	envFile := filepath.Join(rootDir, ".env")

	cmd := exec.Command(binary, "scan", "--dir", rootDir, "--env-file", envFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit with code 1 (mismatches found)
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d\nStdout: %s\nStderr: %s",
				exitErr.ExitCode(), stdout.String(), stderr.String())
		}
	} else if err == nil {
		t.Error("Expected exit code 1, got 0")
	} else {
		t.Errorf("Unexpected error: %v", err)
	}

	output := stdout.String()

	// Should report UNUSED_VAR as declared but unused
	if !strings.Contains(output, "Declared but unused:") {
		t.Errorf("Expected 'Declared but unused:' in output, got: %s", output)
	}
	if !strings.Contains(output, "UNUSED_VAR") {
		t.Errorf("Expected 'UNUSED_VAR' in output, got: %s", output)
	}

	// Should report MISSING_VAR as used but not declared
	if !strings.Contains(output, "Used but not declared:") {
		t.Errorf("Expected 'Used but not declared:' in output, got: %s", output)
	}
	if !strings.Contains(output, "MISSING_VAR") {
		t.Errorf("Expected 'MISSING_VAR' in output, got: %s", output)
	}
}

// TestCLI_MissingEnvFile tests behavior when .env file doesn't exist (exit code 2)
// Requirement: 13.7
func TestCLI_MissingEnvFile(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	nonExistentEnv := filepath.Join(tmpDir, "nonexistent.env")

	cmd := exec.Command(binary, "scan", "--dir", tmpDir, "--env-file", nonExistentEnv)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit with code 2 (configuration error)
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 2 {
			t.Errorf("Expected exit code 2, got %d\nStderr: %s", exitErr.ExitCode(), stderr.String())
		}
	} else if err == nil {
		t.Error("Expected exit code 2, got 0")
	}

	// Should output error message to stderr
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "does not exist") {
		t.Errorf("Expected error message about missing file in stderr, got: %s", stderrOutput)
	}
}

// TestCLI_MissingDirectory tests behavior when directory doesn't exist (exit code 2)
// Requirement: 13.7
func TestCLI_MissingDirectory(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	nonExistentRoot := filepath.Join(tmpDir, "nonexistent")

	// Create .env file
	if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--dir", nonExistentRoot, "--env-file", envFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit with code 2 (configuration error)
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 2 {
			t.Errorf("Expected exit code 2, got %d\nStderr: %s", exitErr.ExitCode(), stderr.String())
		}
	} else if err == nil {
		t.Error("Expected exit code 2, got 0")
	}

	// Should output error message to stderr
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "does not exist") {
		t.Errorf("Expected error message about missing directory in stderr, got: %s", stderrOutput)
	}
}

// TestCLI_MissingRequiredFlags tests behavior when required flags are missing (exit code 2)
// Requirement: 13.7
func TestCLI_MissingRequiredFlags(t *testing.T) {
	binary := buildBinary(t)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "missing dir flag",
			args: []string{"scan", "--env-file", ".env"},
		},
		{
			name: "missing env-file flag",
			args: []string{"scan", "--dir", "."},
		},
		{
			name: "missing both flags",
			args: []string{"scan"},
		},
		{
			name: "missing scan command",
			args: []string{"--dir", ".", "--env-file", ".env"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Should exit with code 2 (configuration error)
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() != 2 {
					t.Errorf("Expected exit code 2, got %d\nStderr: %s", exitErr.ExitCode(), stderr.String())
				}
			} else if err == nil {
				t.Error("Expected exit code 2, got 0")
			}

			// Should output error message to stderr
			stderrOutput := stderr.String()
			if stderrOutput == "" {
				t.Error("Expected error message in stderr, got empty output")
			}
		})
	}
}

// TestCLI_IgnoreFlag tests the --ignore flag functionality
// Requirement: 13.7
func TestCLI_IgnoreFlag(t *testing.T) {
	binary := buildBinary(t)

	// Use the testdata directory which has an "ignored" subdirectory
	rootDir := "../../testdata"
	envFile := filepath.Join(rootDir, ".env")

	// Run with --ignore flag
	cmd := exec.Command(binary, "scan", "--dir", rootDir, "--env-file", envFile, "--ignore", "ignored")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should exit with code 1 (mismatches found, but not from ignored directory)
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d\nStdout: %s\nStderr: %s",
				exitErr.ExitCode(), stdout.String(), stderr.String())
		}
	} else if err == nil {
		t.Error("Expected exit code 1, got 0")
	}

	output := stdout.String()

	// Should NOT report IGNORED_VAR (from ignored directory)
	if strings.Contains(output, "IGNORED_VAR") {
		t.Errorf("Expected IGNORED_VAR to be ignored, but found in output: %s", output)
	}

	// Should still report MISSING_VAR (from non-ignored files)
	if !strings.Contains(output, "MISSING_VAR") {
		t.Errorf("Expected MISSING_VAR in output, got: %s", output)
	}
}

// TestCLI_DeterministicOutput tests that output is deterministic
// Requirement: 13.7
func TestCLI_DeterministicOutput(t *testing.T) {
	binary := buildBinary(t)

	rootDir := "../../testdata"
	envFile := filepath.Join(rootDir, ".env")

	// Run the CLI twice
	var outputs []string
	for i := 0; i < 2; i++ {
		cmd := exec.Command(binary, "scan", "--dir", rootDir, "--env-file", envFile)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		_ = cmd.Run() // Ignore error, we just want the output
		outputs = append(outputs, stdout.String())
	}

	// Outputs should be identical
	if outputs[0] != outputs[1] {
		t.Errorf("Output is not deterministic:\nRun 1:\n%s\n\nRun 2:\n%s", outputs[0], outputs[1])
	}
}
