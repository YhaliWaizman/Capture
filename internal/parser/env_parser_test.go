package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvParser_Parse_ValidFile(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `# This is a comment
API_KEY=secret123
DATABASE_URL=postgres://localhost:5432/db
MAX_RETRIES=5
DEBUG_MODE=true
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()
	result, err := parser.Parse(envFile)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := map[string]bool{
		"API_KEY":      true,
		"DATABASE_URL": true,
		"MAX_RETRIES":  true,
		"DEBUG_MODE":   true,
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d variables, got %d", len(expected), len(result))
	}

	for key := range expected {
		if !result[key] {
			t.Errorf("Expected variable %s not found", key)
		}
	}
}

func TestEnvParser_Parse_EmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `
API_KEY=secret

DATABASE_URL=postgres

`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()
	result, err := parser.Parse(envFile)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result))
	}

	if !result["API_KEY"] || !result["DATABASE_URL"] {
		t.Error("Expected variables not found")
	}
}

func TestEnvParser_Parse_Comments(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `# Comment at start
API_KEY=secret
# Another comment
# Yet another comment
DATABASE_URL=postgres
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()
	result, err := parser.Parse(envFile)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result))
	}
}

func TestEnvParser_Parse_InvalidKeys(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `VALID_KEY=value
lowercase_key=value
mixedCase_Key=value
123_STARTS_WITH_NUMBER=value
KEY-WITH-DASH=value
KEY.WITH.DOT=value
VALID_KEY_2=value
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()
	result, err := parser.Parse(envFile)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Only VALID_KEY and VALID_KEY_2 should be included
	if len(result) != 2 {
		t.Errorf("Expected 2 valid variables, got %d", len(result))
	}

	if !result["VALID_KEY"] || !result["VALID_KEY_2"] {
		t.Error("Expected valid variables not found")
	}

	// Verify invalid keys are not included
	invalidKeys := []string{"lowercase_key", "mixedCase_Key", "123_STARTS_WITH_NUMBER", "KEY-WITH-DASH", "KEY.WITH.DOT"}
	for _, key := range invalidKeys {
		if result[key] {
			t.Errorf("Invalid key %s should not be included", key)
		}
	}
}

func TestEnvParser_Parse_DuplicateKeys(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `API_KEY=first_value
API_KEY=second_value
DATABASE_URL=postgres
API_KEY=third_value
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()
	result, err := parser.Parse(envFile)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have unique set of keys
	if len(result) != 2 {
		t.Errorf("Expected 2 unique variables, got %d", len(result))
	}

	if !result["API_KEY"] || !result["DATABASE_URL"] {
		t.Error("Expected variables not found")
	}
}

func TestEnvParser_Parse_WhitespaceTrimming(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `  API_KEY=value
	DATABASE_URL=value
MAX_RETRIES  =value
  DEBUG_MODE  =value  
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()
	result, err := parser.Parse(envFile)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := []string{"API_KEY", "DATABASE_URL", "MAX_RETRIES", "DEBUG_MODE"}

	if len(result) != len(expected) {
		t.Errorf("Expected %d variables, got %d", len(expected), len(result))
	}

	for _, key := range expected {
		if !result[key] {
			t.Errorf("Expected variable %s not found", key)
		}
	}
}

func TestEnvParser_Parse_ValueFormats(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `SIMPLE=value
WITH_SPACES=value with spaces
WITH_QUOTES="quoted value"
WITH_EQUALS=key=value=more
EMPTY_VALUE=
URL=https://example.com/path?query=param
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()
	result, err := parser.Parse(envFile)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := []string{"SIMPLE", "WITH_SPACES", "WITH_QUOTES", "WITH_EQUALS", "EMPTY_VALUE", "URL"}

	if len(result) != len(expected) {
		t.Errorf("Expected %d variables, got %d", len(expected), len(result))
	}

	for _, key := range expected {
		if !result[key] {
			t.Errorf("Expected variable %s not found", key)
		}
	}
}

func TestEnvParser_Parse_FileNotFound(t *testing.T) {
	parser := NewEnvParser()
	_, err := parser.Parse("/nonexistent/path/.env")

	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestEnvParser_Parse_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(envFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()
	result, err := parser.Parse(envFile)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 variables for empty file, got %d", len(result))
	}
}

func TestEnvParser_Parse_OnlyComments(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `# Comment 1
# Comment 2
# Comment 3
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()
	result, err := parser.Parse(envFile)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 variables for comments-only file, got %d", len(result))
	}
}

func TestEnvParser_Parse_Idempotence(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `API_KEY=secret
DATABASE_URL=postgres
MAX_RETRIES=5
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewEnvParser()

	// Parse twice
	result1, err1 := parser.Parse(envFile)
	if err1 != nil {
		t.Fatalf("First parse failed: %v", err1)
	}

	result2, err2 := parser.Parse(envFile)
	if err2 != nil {
		t.Fatalf("Second parse failed: %v", err2)
	}

	// Results should be identical
	if len(result1) != len(result2) {
		t.Errorf("Idempotence violated: different lengths %d vs %d", len(result1), len(result2))
	}

	for key := range result1 {
		if !result2[key] {
			t.Errorf("Idempotence violated: key %s missing in second parse", key)
		}
	}

	for key := range result2 {
		if !result1[key] {
			t.Errorf("Idempotence violated: key %s missing in first parse", key)
		}
	}
}
