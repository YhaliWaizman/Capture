package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonDetector_OsGetenvPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	content := `import os

api_key = os.getenv("API_KEY")
db_url = os.getenv("DATABASE_URL")`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewPythonDetector()
	result, err := detector.Detect(testFile)

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result))
	}

	if _, ok := result["API_KEY"]; !ok {
		t.Error("Expected API_KEY to be detected")
	}

	if _, ok := result["DATABASE_URL"]; !ok {
		t.Error("Expected DATABASE_URL to be detected")
	}
}

func TestPythonDetector_OsEnvironBracketDoubleQuote(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	content := `import os

api_key = os.environ["API_KEY"]
db_url = os.environ["DATABASE_URL"]`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewPythonDetector()
	result, err := detector.Detect(testFile)

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result))
	}

	if _, ok := result["API_KEY"]; !ok {
		t.Error("Expected API_KEY to be detected")
	}

	if _, ok := result["DATABASE_URL"]; !ok {
		t.Error("Expected DATABASE_URL to be detected")
	}
}

func TestPythonDetector_OsEnvironBracketSingleQuote(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	content := `import os

api_key = os.environ['API_KEY']
db_url = os.environ['DATABASE_URL']`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewPythonDetector()
	result, err := detector.Detect(testFile)

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result))
	}

	if _, ok := result["API_KEY"]; !ok {
		t.Error("Expected API_KEY to be detected")
	}

	if _, ok := result["DATABASE_URL"]; !ok {
		t.Error("Expected DATABASE_URL to be detected")
	}
}

func TestPythonDetector_DynamicExpressionRejection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	content := `import os

key1 = os.getenv(var_name)
key2 = os.environ[get_key()]
valid = os.getenv("VALID_KEY")`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewPythonDetector()
	result, err := detector.Detect(testFile)

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	// Should only detect VALID_KEY, not the dynamic expressions
	if len(result) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(result))
	}

	if _, ok := result["VALID_KEY"]; !ok {
		t.Error("Expected VALID_KEY to be detected")
	}
}

func TestPythonDetector_LineNumberRecording(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.py")
	content := `import os
key1 = os.getenv("API_KEY")
key2 = os.environ["DATABASE_URL"]
key3 = os.environ['API_KEY']`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewPythonDetector()
	result, err := detector.Detect(testFile)

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	// Check API_KEY appears on lines 2 and 4
	apiKeyLocs := result["API_KEY"]
	if len(apiKeyLocs) != 2 {
		t.Errorf("Expected API_KEY to appear 2 times, got %d", len(apiKeyLocs))
	}

	if apiKeyLocs[0].LineNumber != 2 {
		t.Errorf("Expected first API_KEY on line 2, got %d", apiKeyLocs[0].LineNumber)
	}

	if apiKeyLocs[1].LineNumber != 4 {
		t.Errorf("Expected second API_KEY on line 4, got %d", apiKeyLocs[1].LineNumber)
	}

	// Check DATABASE_URL appears on line 3
	dbUrlLocs := result["DATABASE_URL"]
	if len(dbUrlLocs) != 1 {
		t.Errorf("Expected DATABASE_URL to appear 1 time, got %d", len(dbUrlLocs))
	}

	if dbUrlLocs[0].LineNumber != 3 {
		t.Errorf("Expected DATABASE_URL on line 3, got %d", dbUrlLocs[0].LineNumber)
	}
}
