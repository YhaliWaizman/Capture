package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
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
	if !strings.Contains(stderrOutput, "no readable env files found") {
		t.Errorf("Expected error message about unreadable env files in stderr, got: %s", stderrOutput)
	}
}

// TestCLI_MultipleEnvFiles tests repeated --env-file handling.
func TestCLI_MultipleEnvFiles(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	baseEnv := filepath.Join(tmpDir, ".env")
	localEnv := filepath.Join(tmpDir, ".env.local")
	srcFile := filepath.Join(tmpDir, "app.js")

	if err := os.WriteFile(baseEnv, []byte("API_KEY=base\n"), 0644); err != nil {
		t.Fatalf("Failed to write base .env file: %v", err)
	}
	if err := os.WriteFile(localEnv, []byte("LOCAL_ONLY=1\n"), 0644); err != nil {
		t.Fatalf("Failed to write local .env file: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY, process.env.LOCAL_ONLY);\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--dir", tmpDir, "--env-file", baseEnv, "--env-file", localEnv)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Errorf("Expected exit code 0, got error: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), "No environment mismatches found.") {
		t.Errorf("Expected success output, got: %s", stdout.String())
	}
}

// TestCLI_MissingEnvFile_WarnAndContinue ensures a missing env file does not fail the scan
// if at least one env file is readable.
func TestCLI_MissingEnvFile_WarnAndContinue(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	existingEnv := filepath.Join(tmpDir, ".env")
	missingEnv := filepath.Join(tmpDir, ".env.local")
	srcFile := filepath.Join(tmpDir, "app.js")

	if err := os.WriteFile(existingEnv, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("const key = process.env.API_KEY;\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--dir", tmpDir, "--env-file", existingEnv, "--env-file", missingEnv)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Errorf("Expected exit code 0, got error: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	if !strings.Contains(stderr.String(), "Warning: failed to parse .env file") {
		t.Errorf("Expected warning for missing env file, got stderr: %s", stderr.String())
	}
}

func TestCLI_ConfigFile_AutoDiscovery_AllFormats(t *testing.T) {
	binary := buildBinary(t)

	tests := []struct {
		name       string
		configName string
		configBody string
	}{
		{
			name:       "yaml",
			configName: ".capture.yaml",
			configBody: "root: .\nenv_files:\n  - .env\nformat: text\n",
		},
		{
			name:       "yml",
			configName: ".capture.yml",
			configBody: "root: .\nenv_files:\n  - .env\nformat: text\n",
		},
		{
			name:       "json",
			configName: ".capture.json",
			configBody: "{\n  \"root\": \".\",\n  \"env_files\": [\".env\"],\n  \"format\": \"text\"\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			srcFile := filepath.Join(tmpDir, "app.js")
			configFile := filepath.Join(tmpDir, tt.configName)

			if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
				t.Fatalf("Failed to write .env file: %v", err)
			}
			if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY);\n"), 0644); err != nil {
				t.Fatalf("Failed to write source file: %v", err)
			}
			if err := os.WriteFile(configFile, []byte(tt.configBody), 0644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			cmd := exec.Command(binary, "scan")
			cmd.Dir = tmpDir
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("Expected exit code 0, got: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
			}

			if !strings.Contains(stdout.String(), "No environment mismatches found.") {
				t.Fatalf("Expected success output, got: %s", stdout.String())
			}
		})
	}
}

func TestCLI_ConfigFile_FlagOverridesConfigValues(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	srcFile := filepath.Join(tmpDir, "app.js")
	configFile := filepath.Join(tmpDir, ".capture.yaml")

	if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY);\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}
	if err := os.WriteFile(configFile, []byte("root: .\nenv_files:\n  - missing.env\nformat: json\n"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--env-file", ".env", "--format", "text")
	cmd.Dir = tmpDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected exit code 0, got: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	if strings.Contains(stdout.String(), "{") {
		t.Fatalf("Expected text output override, got JSON-like output: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "No environment mismatches found.") {
		t.Fatalf("Expected success output, got: %s", stdout.String())
	}
}

func TestCLI_ConfigFile_ExplicitConfigPath(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	srcFile := filepath.Join(tmpDir, "app.js")
	configFile := filepath.Join(tmpDir, "capture.config.json")

	if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY);\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}
	if err := os.WriteFile(configFile, []byte("{\n  \"root\": \".\",\n  \"env_files\": [\".env\"],\n  \"format\": \"text\"\n}\n"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--config", configFile)
	cmd.Dir = tmpDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected exit code 0, got: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), "No environment mismatches found.") {
		t.Fatalf("Expected success output, got: %s", stdout.String())
	}
}

func TestCLI_ConfigFile_InvalidConfig(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	srcFile := filepath.Join(tmpDir, "app.js")
	configFile := filepath.Join(tmpDir, ".capture.yaml")

	if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY);\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}
	if err := os.WriteFile(configFile, []byte("root: .\nunknown_field: true\n"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cmd := exec.Command(binary, "scan")
	cmd.Dir = tmpDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 2 {
			t.Fatalf("Expected exit code 2, got %d\nStdout: %s\nStderr: %s", exitErr.ExitCode(), stdout.String(), stderr.String())
		}
	} else if err == nil {
		t.Fatal("Expected config validation error, got exit code 0")
	} else {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(stderr.String(), "invalid YAML config") {
		t.Fatalf("Expected invalid config error, got stderr: %s", stderr.String())
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

// TestCLI_DefaultFlagBehavior tests behavior when flags use default values
// Requirement: 13.7
func TestCLI_DefaultFlagBehavior(t *testing.T) {
	binary := buildBinary(t)

	tests := []struct {
		name          string
		args          []string
		expectedInErr string
		setupFunc     func(t *testing.T) string // Returns working directory
	}{
		{
			name:          "missing dir flag uses default and fails if not exists",
			args:          []string{"scan", "--env-file", ".env"},
			expectedInErr: "no readable env files found",
			setupFunc: func(t *testing.T) string {
				// Create temp dir without .env file
				tmpDir := t.TempDir()
				return tmpDir
			},
		},
		{
			name:          "missing env-file flag uses default and fails if not exists",
			args:          []string{"scan", "--dir", "."},
			expectedInErr: "no readable env files found",
			setupFunc: func(t *testing.T) string {
				// Create temp dir without .env file
				tmpDir := t.TempDir()
				return tmpDir
			},
		},
		{
			name:          "missing both flags uses defaults and fails",
			args:          []string{"scan"},
			expectedInErr: "no readable env files found",
			setupFunc: func(t *testing.T) string {
				// Create temp dir without .env file
				tmpDir := t.TempDir()
				return tmpDir
			},
		},
		{
			name:          "missing scan command shows help",
			args:          []string{"--dir", ".", "--env-file", ".env"},
			expectedInErr: "", // This will show help or unknown command
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := tt.setupFunc(t)

			cmd := exec.Command(binary, tt.args...)
			cmd.Dir = workDir // Run in the temp directory
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Should exit with non-zero code
			if err == nil {
				t.Error("Expected non-zero exit code, got 0")
			}

			// Check for expected error message if specified
			if tt.expectedInErr != "" {
				stderrOutput := stderr.String()
				if !strings.Contains(stderrOutput, tt.expectedInErr) {
					t.Errorf("Expected '%s' in stderr, got: %s", tt.expectedInErr, stderrOutput)
				}
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

func TestCLI_WorkersFlag_DeterministicParity(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("DECLARED_ONLY=1\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	sourceFiles := map[string]string{
		"app.js":      "console.log(process.env.MISSING_C)\nconsole.log(process.env.MISSING_A)\n",
		"main.go":     "package main\nimport \"os\"\nfunc main(){_ = os.Getenv(\"MISSING_B\")}\n",
		"service.py":  "import os\nprint(os.getenv(\"MISSING_A\"))\n",
		"worker.rb":   "puts ENV['MISSING_B']\n",
		"feature.php": "<?php echo getenv('MISSING_C');\n",
	}
	for name, content := range sourceFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write source file %s: %v", name, err)
		}
	}

	runScan := func(workers string) (string, int, string) {
		cmd := exec.Command(binary, "scan", "--dir", tmpDir, "--env-file", envFile, "--workers", workers)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout.String(), exitErr.ExitCode(), stderr.String()
		}
		if err != nil {
			t.Fatalf("Unexpected command error with workers=%s: %v", workers, err)
		}
		return stdout.String(), 0, stderr.String()
	}

	out1, code1, err1 := runScan("1")
	out4, code4, err4 := runScan("4")

	if code1 != 1 || code4 != 1 {
		t.Fatalf("Expected exit code 1 for mismatches, got workers=1:%d workers=4:%d", code1, code4)
	}
	if err1 != "" || err4 != "" {
		t.Fatalf("Expected empty stderr, got workers=1:%q workers=4:%q", err1, err4)
	}
	if out1 != out4 {
		t.Fatalf("Expected identical output across worker counts.\nworkers=1:\n%s\nworkers=4:\n%s", out1, out4)
	}
}

func TestCLI_IncrementalScan_CreatesCacheAndDetectsGitChanges(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	srcFile := filepath.Join(tmpDir, "app.js")

	if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY)\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	initCmd := exec.Command("git", "init")
	initCmd.Dir = tmpDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init git repo: %v\nOutput: %s", err, string(out))
	}
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	if out, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add files: %v\nOutput: %s", err, string(out))
	}
	commitCmd := exec.Command("git", "-c", "user.name=Capture Test", "-c", "user.email=capture@example.com", "commit", "-m", "init")
	commitCmd.Dir = tmpDir
	if out, err := commitCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit files: %v\nOutput: %s", err, string(out))
	}

	firstRun := exec.Command(binary, "scan", "--dir", tmpDir, "--env-file", envFile, "--incremental")
	var firstOut, firstErr bytes.Buffer
	firstRun.Stdout = &firstOut
	firstRun.Stderr = &firstErr
	if err := firstRun.Run(); err != nil {
		t.Fatalf("Expected first incremental run to succeed, got: %v\nStdout: %s\nStderr: %s", err, firstOut.String(), firstErr.String())
	}

	cacheFile := filepath.Join(tmpDir, ".capture", "cache.json")
	if _, err := os.Stat(cacheFile); err != nil {
		t.Fatalf("Expected cache file to be created at %s: %v", cacheFile, err)
	}

	if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY)\nconsole.log(process.env.NEW_VAR)\n"), 0644); err != nil {
		t.Fatalf("Failed to update source file: %v", err)
	}

	secondRun := exec.Command(binary, "scan", "--dir", tmpDir, "--env-file", envFile, "--incremental")
	var secondOut, secondErr bytes.Buffer
	secondRun.Stdout = &secondOut
	secondRun.Stderr = &secondErr
	err := secondRun.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("Expected exit code 1 after changed file scan, got %d\nStdout: %s\nStderr: %s", exitErr.ExitCode(), secondOut.String(), secondErr.String())
		}
	} else if err == nil {
		t.Fatal("Expected mismatches (exit code 1) after introducing NEW_VAR")
	} else {
		t.Fatalf("Unexpected command error: %v\nStdout: %s\nStderr: %s", err, secondOut.String(), secondErr.String())
	}

	if !strings.Contains(secondOut.String(), "NEW_VAR") {
		t.Fatalf("Expected NEW_VAR mismatch in output, got: %s", secondOut.String())
	}
}

func TestCLI_IncrementalScan_NoCacheFlagForcesFullScan(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	srcFile := filepath.Join(tmpDir, "app.js")

	if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY)\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--dir", tmpDir, "--env-file", envFile, "--incremental", "--no-cache")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected scan to succeed, got: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	cacheFile := filepath.Join(tmpDir, ".capture", "cache.json")
	if _, err := os.Stat(cacheFile); !os.IsNotExist(err) {
		t.Fatalf("Expected no cache file when --no-cache is set, got stat err=%v", err)
	}
}

func TestCLI_IncrementalScan_FallsBackWithoutGit(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	srcFile := filepath.Join(tmpDir, "app.js")

	if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("console.log(process.env.MISSING_VAR)\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--dir", tmpDir, "--env-file", envFile, "--incremental")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("Expected exit code 1 for mismatch fallback scan, got %d\nStdout: %s\nStderr: %s", exitErr.ExitCode(), stdout.String(), stderr.String())
		}
	} else if err == nil {
		t.Fatal("Expected mismatch exit code 1, got 0")
	} else {
		t.Fatalf("Unexpected command error: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), "MISSING_VAR") {
		t.Fatalf("Expected MISSING_VAR in output, got: %s", stdout.String())
	}
}

func TestCLI_WatchMode_ReRunsOnChange(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	srcFile := filepath.Join(tmpDir, "app.js")

	if err := os.WriteFile(envFile, []byte("API_KEY=test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY)\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, "scan", "--dir", tmpDir, "--env-file", envFile, "--watch")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start watch command: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	if err := os.WriteFile(srcFile, []byte("console.log(process.env.API_KEY)\nconsole.log(process.env.MISSING_WATCH_MODE)\n"), 0644); err != nil {
		t.Fatalf("Failed to update source file: %v", err)
	}

	time.Sleep(2 * time.Second)

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("Failed to send interrupt: %v", err)
	}
	waitErr := cmd.Wait()

	if waitErr != nil {
		t.Fatalf("Expected graceful shutdown on interrupt, got: %v\nStdout: %s\nStderr: %s", waitErr, stdoutBuf.String(), stderrBuf.String())
	}
	if !strings.Contains(stderrBuf.String(), "Watching for changes... (Press Ctrl+C to stop)") {
		t.Fatalf("Expected watch mode status message, got stderr: %s", stderrBuf.String())
	}
	if !strings.Contains(stdoutBuf.String(), "MISSING_WATCH_MODE") {
		t.Fatalf("Expected re-scan output after file change.\nStdout: %s\nStderr: %s", stdoutBuf.String(), stderrBuf.String())
	}
	if matched := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`).MatchString(stderrBuf.String()); !matched {
		t.Fatalf("Expected timestamp in watch output, got stderr: %s", stderrBuf.String())
	}
}
