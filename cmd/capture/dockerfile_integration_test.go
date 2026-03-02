package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLI_DockerfileIntegration tests Dockerfile analysis integration
func TestCLI_DockerfileIntegration(t *testing.T) {
	binary := buildBinary(t)

	// Use the testdata directory which has Dockerfiles
	rootDir := "../../testdata"
	envFile := filepath.Join(rootDir, ".env")

	cmd := exec.Command(binary, "scan", "--root", rootDir, "--env-file", envFile)
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
	}

	output := stdout.String()

	// Verify Dockerfile-specific sections are present
	if !strings.Contains(output, "Code uses variables not in Dockerfile or .env:") {
		t.Error("Expected 'Code uses variables not in Dockerfile or .env:' section")
	}

	if !strings.Contains(output, "Dockerfile declares but code doesn't use:") {
		t.Error("Expected 'Dockerfile declares but code doesn't use:' section")
	}

	if !strings.Contains(output, "Dockerfile uses undeclared variables:") {
		t.Error("Expected 'Dockerfile uses undeclared variables:' section")
	}

	// Verify specific variables are reported
	// API_KEY is declared in Dockerfile and used in code
	// NODE_ENV is declared in Dockerfile but not in .env
	if !strings.Contains(output, "NODE_ENV") {
		t.Error("Expected NODE_ENV to be reported (used in code, declared in Dockerfile)")
	}
}

// TestCLI_DockerfileOnly tests scanning with only Dockerfile, no source files
func TestCLI_DockerfileOnly(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	// Create .env file
	envContent := `API_KEY=secret
DATABASE_URL=postgres://localhost
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Create Dockerfile
	dockerfileContent := `FROM alpine
ENV API_KEY=default
ENV EXTRA_VAR=value
RUN echo "$API_KEY"
`
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--root", tmpDir, "--env-file", envFile)
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
	}

	output := stdout.String()

	// DATABASE_URL is in .env but not used anywhere
	if !strings.Contains(output, "DATABASE_URL") {
		t.Error("Expected DATABASE_URL to be reported as unused")
	}

	// EXTRA_VAR is in Dockerfile but not in .env and not used in code
	if !strings.Contains(output, "EXTRA_VAR") {
		t.Error("Expected EXTRA_VAR to be reported")
	}
}

// TestCLI_MultiStageDockerfile tests multi-stage Dockerfile analysis
func TestCLI_MultiStageDockerfile(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	// Create .env file
	envContent := `BUILD_ENV=production
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Create multi-stage Dockerfile
	dockerfileContent := `FROM node:18 AS builder
ENV BUILD_ENV=production
ARG BUILD_VERSION

FROM node:18-alpine AS runtime
ENV RUNTIME_ENV=production

RUN echo "Build: $BUILD_ENV"
RUN echo "Runtime: $RUNTIME_ENV"
`
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--root", tmpDir, "--env-file", envFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run() // Ignore exit code, just check output

	output := stdout.String()

	// Both BUILD_ENV and RUNTIME_ENV should be detected
	// RUNTIME_ENV is declared in Dockerfile but not in .env
	if !strings.Contains(output, "RUNTIME_ENV") {
		t.Error("Expected RUNTIME_ENV to be detected from multi-stage Dockerfile")
	}

	// BUILD_VERSION is declared but not used in code
	if !strings.Contains(output, "BUILD_VERSION") {
		t.Error("Expected BUILD_VERSION to be detected")
	}
}

// TestCLI_DockerfileLineContinuation tests line continuation handling
func TestCLI_DockerfileLineContinuation(t *testing.T) {
	binary := buildBinary(t)

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	// Create .env file
	envContent := `KEY1=value1
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Create Dockerfile with line continuations
	dockerfileContent := `FROM alpine
ENV KEY1=value1 \
    KEY2=value2 \
    KEY3=value3

RUN echo "$KEY1 $KEY2 $KEY3"
`
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	cmd := exec.Command(binary, "scan", "--root", tmpDir, "--env-file", envFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run()

	output := stdout.String()

	// KEY2 and KEY3 should be detected from multi-line ENV
	if !strings.Contains(output, "KEY2") || !strings.Contains(output, "KEY3") {
		t.Error("Expected KEY2 and KEY3 to be detected from line continuation")
	}
}
