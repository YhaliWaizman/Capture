package dockerfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDockerfileAnalyzer_SimpleDockerfile(t *testing.T) {
	analyzer := NewDockerfileAnalyzer()

	// Create a temporary Dockerfile
	tmpDir := t.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	content := `FROM node:18-alpine

ARG BUILD_VERSION
ENV API_KEY=secret
ENV DATABASE_URL postgres://localhost

RUN echo "Version: $BUILD_VERSION"
RUN echo "API: ${API_KEY}"
`

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test Dockerfile: %v", err)
	}

	result, err := analyzer.Analyze(dockerfilePath)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Check declared variables
	expectedDeclared := []string{"BUILD_VERSION", "API_KEY", "DATABASE_URL"}
	for _, varName := range expectedDeclared {
		if !result.Declared[varName] {
			t.Errorf("Expected %s to be declared", varName)
		}
	}

	// Check used variables
	expectedUsed := []string{"BUILD_VERSION", "API_KEY"}
	for _, varName := range expectedUsed {
		if _, found := result.Used[varName]; !found {
			t.Errorf("Expected %s to be used", varName)
		}
	}
}

func TestDockerfileAnalyzer_ENVFormats(t *testing.T) {
	analyzer := NewDockerfileAnalyzer()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "ENV with equals",
			content: `FROM alpine
ENV KEY=value`,
			expected: []string{"KEY"},
		},
		{
			name: "ENV with space",
			content: `FROM alpine
ENV KEY value`,
			expected: []string{"KEY"},
		},
		{
			name: "ENV multiple with equals",
			content: `FROM alpine
ENV A=1 B=2 C=3`,
			expected: []string{"A", "B", "C"},
		},
		{
			name: "ENV lowercase ignored",
			content: `FROM alpine
ENV lowercase=value
ENV UPPERCASE=value`,
			expected: []string{"UPPERCASE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

			if err := os.WriteFile(dockerfilePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test Dockerfile: %v", err)
			}

			result, err := analyzer.Analyze(dockerfilePath)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			if len(result.Declared) != len(tt.expected) {
				t.Errorf("Expected %d declared variables, got %d", len(tt.expected), len(result.Declared))
			}

			for _, varName := range tt.expected {
				if !result.Declared[varName] {
					t.Errorf("Expected %s to be declared", varName)
				}
			}
		})
	}
}

func TestDockerfileAnalyzer_ARGFormats(t *testing.T) {
	analyzer := NewDockerfileAnalyzer()

	tmpDir := t.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	content := `FROM alpine
ARG KEY1
ARG KEY2=default
ARG lowercase
ARG VALID_KEY=value
`

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test Dockerfile: %v", err)
	}

	result, err := analyzer.Analyze(dockerfilePath)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	expected := []string{"KEY1", "KEY2", "VALID_KEY"}
	if len(result.Declared) != len(expected) {
		t.Errorf("Expected %d declared variables, got %d", len(expected), len(result.Declared))
	}

	for _, varName := range expected {
		if !result.Declared[varName] {
			t.Errorf("Expected %s to be declared", varName)
		}
	}

	// Ensure lowercase is not included
	if result.Declared["lowercase"] {
		t.Error("Lowercase variable should not be declared")
	}
}

func TestDockerfileAnalyzer_LineContinuation(t *testing.T) {
	analyzer := NewDockerfileAnalyzer()

	tmpDir := t.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	content := `FROM alpine
ENV KEY1=value1 \
    KEY2=value2 \
    KEY3=value3

RUN echo "Test: $KEY1 $KEY2 $KEY3"
`

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test Dockerfile: %v", err)
	}

	result, err := analyzer.Analyze(dockerfilePath)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	expected := []string{"KEY1", "KEY2", "KEY3"}
	for _, varName := range expected {
		if !result.Declared[varName] {
			t.Errorf("Expected %s to be declared", varName)
		}
	}
}

func TestDockerfileAnalyzer_VariableUsage(t *testing.T) {
	analyzer := NewDockerfileAnalyzer()

	tmpDir := t.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	content := `FROM alpine
ENV DECLARED=value

RUN echo "$VAR1"
RUN echo "${VAR2}"
RUN echo "Mixed: $VAR1 and ${VAR2}"
RUN echo "$DECLARED"
`

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test Dockerfile: %v", err)
	}

	result, err := analyzer.Analyze(dockerfilePath)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	expectedUsed := []string{"VAR1", "VAR2", "DECLARED"}
	for _, varName := range expectedUsed {
		if _, found := result.Used[varName]; !found {
			t.Errorf("Expected %s to be used", varName)
		}
	}
}

func TestDockerfileAnalyzer_NoFROM(t *testing.T) {
	analyzer := NewDockerfileAnalyzer()

	tmpDir := t.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	// Invalid Dockerfile without FROM
	content := `ENV KEY=value
RUN echo "test"
`

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test Dockerfile: %v", err)
	}

	result, err := analyzer.Analyze(dockerfilePath)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should return empty result for invalid Dockerfile
	if len(result.Declared) != 0 {
		t.Error("Expected no declared variables for invalid Dockerfile")
	}
	if len(result.Used) != 0 {
		t.Error("Expected no used variables for invalid Dockerfile")
	}
}

func TestDockerfileAnalyzer_CaseInsensitiveInstructions(t *testing.T) {
	analyzer := NewDockerfileAnalyzer()

	tmpDir := t.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	content := `from alpine
env KEY1=value1
arg KEY2=value2
run echo "$KEY1 $KEY2"
`

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test Dockerfile: %v", err)
	}

	result, err := analyzer.Analyze(dockerfilePath)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should recognize lowercase instructions
	if !result.Declared["KEY1"] {
		t.Error("Expected KEY1 to be declared (lowercase env)")
	}
	if !result.Declared["KEY2"] {
		t.Error("Expected KEY2 to be declared (lowercase arg)")
	}
}

func TestDockerfileAnalyzer_MultiStage(t *testing.T) {
	analyzer := NewDockerfileAnalyzer()

	tmpDir := t.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")

	content := `FROM node:18 AS builder
ENV BUILD_ENV=production
ARG BUILD_VERSION

FROM node:18-alpine AS runtime
ENV RUNTIME_ENV=production

RUN echo "Build: $BUILD_ENV"
RUN echo "Runtime: $RUNTIME_ENV"
RUN echo "Version: $BUILD_VERSION"
`

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test Dockerfile: %v", err)
	}

	result, err := analyzer.Analyze(dockerfilePath)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// All declarations should be captured
	expectedDeclared := []string{"BUILD_ENV", "BUILD_VERSION", "RUNTIME_ENV"}
	for _, varName := range expectedDeclared {
		if !result.Declared[varName] {
			t.Errorf("Expected %s to be declared", varName)
		}
	}
}
