package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPHPDetector_ENVSuperglobalPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.php")
	content := `<?php
$db = $_ENV['DATABASE_URL'];
$key = $_ENV["API_KEY"];`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewPHPDetector()
	result, err := detector.Detect(testFile)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result))
	}
	if _, ok := result["DATABASE_URL"]; !ok {
		t.Error("Expected DATABASE_URL to be detected")
	}
	if _, ok := result["API_KEY"]; !ok {
		t.Error("Expected API_KEY to be detected")
	}
}

func TestPHPDetector_SERVERSuperglobalPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.php")
	content := `<?php
$db = $_SERVER['DATABASE_URL'];
$key = $_SERVER["API_KEY"];`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewPHPDetector()
	result, err := detector.Detect(testFile)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result))
	}
	if _, ok := result["DATABASE_URL"]; !ok {
		t.Error("Expected DATABASE_URL to be detected")
	}
	if _, ok := result["API_KEY"]; !ok {
		t.Error("Expected API_KEY to be detected")
	}
}

func TestPHPDetector_GetenvPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.php")
	content := `<?php
$db = getenv('DATABASE_URL');
$key = getenv("API_KEY");`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewPHPDetector()
	result, err := detector.Detect(testFile)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(result))
	}
	if _, ok := result["DATABASE_URL"]; !ok {
		t.Error("Expected DATABASE_URL to be detected")
	}
	if _, ok := result["API_KEY"]; !ok {
		t.Error("Expected API_KEY to be detected")
	}
}

func TestPHPDetector_DynamicExpressionRejection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.php")
	content := `<?php
$key1 = $_ENV[$name];
$key2 = $_SERVER[getKey()];
$key3 = getenv($dynamic);
$valid = $_ENV["VALID_KEY"];`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewPHPDetector()
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
