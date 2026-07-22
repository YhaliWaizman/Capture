package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRubyDetector_ENVBracketPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.rb")
	content := "api_key = ENV['API_KEY']\n" +
		`db_url = ENV["DATABASE_URL"]`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewRubyDetector()
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

func TestRubyDetector_ENVFetchPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.rb")
	content := `api_key = ENV.fetch("API_KEY")
db_url = ENV.fetch('DATABASE_URL', nil)`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewRubyDetector()
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

func TestRubyDetector_DynamicExpressionRejection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.rb")
	content := `key1 = ENV[var_name]
key2 = ENV.fetch(get_key())
valid = ENV.fetch("VALID_KEY")`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewRubyDetector()
	result, err := detector.Detect(testFile)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(result))
	}
	if _, ok := result["VALID_KEY"]; !ok {
		t.Error("Expected VALID_KEY to be detected")
	}
}

func TestRubyDetector_LineNumberRecording(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.rb")
	content := `key1 = ENV['API_KEY']
key2 = ENV.fetch("DATABASE_URL")
key3 = ENV["API_KEY"]`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewRubyDetector()
	result, err := detector.Detect(testFile)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	apiKeyLocs := result["API_KEY"]
	if len(apiKeyLocs) != 2 {
		t.Errorf("Expected API_KEY to appear 2 times, got %d", len(apiKeyLocs))
	}
	if apiKeyLocs[0].LineNumber != 1 {
		t.Errorf("Expected first API_KEY on line 1, got %d", apiKeyLocs[0].LineNumber)
	}
	if apiKeyLocs[1].LineNumber != 3 {
		t.Errorf("Expected second API_KEY on line 3, got %d", apiKeyLocs[1].LineNumber)
	}

	dbURLLocs := result["DATABASE_URL"]
	if len(dbURLLocs) != 1 {
		t.Errorf("Expected DATABASE_URL to appear 1 time, got %d", len(dbURLLocs))
	}
	if dbURLLocs[0].LineNumber != 2 {
		t.Errorf("Expected DATABASE_URL on line 2, got %d", dbURLLocs[0].LineNumber)
	}
}
