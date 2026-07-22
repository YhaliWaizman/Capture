package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJVMDetector_JavaGetenvPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "Test.java")
	content := `class Test {
  void run() {
    var db = System.getenv("DATABASE_URL");
    var key = System.getenv().get("API_KEY");
  }
}`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewJVMDetector()
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

func TestJVMDetector_KotlinGetenvMapPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.kt")
	content := `fun main() {
  val key = System.getenv()["API_KEY"]
  val db = System.getenv("DATABASE_URL")
}`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewJVMDetector()
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

func TestJVMDetector_DynamicExpressionRejection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.kts")
	content := `val key1 = System.getenv(name)
val key2 = System.getenv().get(dynamicKey)
val key3 = System.getenv()[dynamicIndex]
val valid = System.getenv("VALID_KEY")`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewJVMDetector()
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
