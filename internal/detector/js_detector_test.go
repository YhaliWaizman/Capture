package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSDetector_ProcessEnvDotPattern(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	content := `const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewJSDetector()
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

func TestJSDetector_ProcessEnvBracketDoubleQuote(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	content := `const apiKey = process.env["API_KEY"];
const dbUrl = process.env["DATABASE_URL"];`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewJSDetector()
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

func TestJSDetector_ProcessEnvBracketSingleQuote(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	content := `const apiKey = process.env['API_KEY'];
const dbUrl = process.env['DATABASE_URL'];`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewJSDetector()
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

func TestJSDetector_DynamicExpressionRejection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	content := "const key1 = process.env[varName];\n" +
		"const key2 = process.env[`VAR_${x}`];\n" +
		"const key3 = process.env[getKey()];\n" +
		"const valid = process.env.VALID_KEY;"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewJSDetector()
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

	// Ensure dynamic expressions are not detected
	if _, ok := result["varName"]; ok {
		t.Error("Should not detect variable reference")
	}
}

func TestJSDetector_LineNumberRecording(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	content := `// Line 1
const key1 = process.env.API_KEY; // Line 2
const key2 = process.env.DATABASE_URL; // Line 3
const key3 = process.env.API_KEY; // Line 4`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewJSDetector()
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

func TestJSDetector_MultipleOccurrencesSameVariable(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")
	content := `const key1 = process.env.API_KEY;
const key2 = process.env["API_KEY"];
const key3 = process.env['API_KEY'];`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewJSDetector()
	result, err := detector.Detect(testFile)

	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	// Should detect API_KEY 3 times
	apiKeyLocs := result["API_KEY"]
	if len(apiKeyLocs) != 3 {
		t.Errorf("Expected API_KEY to appear 3 times, got %d", len(apiKeyLocs))
	}

	// Verify all three line numbers
	expectedLines := []int{1, 2, 3}
	for i, loc := range apiKeyLocs {
		if loc.LineNumber != expectedLines[i] {
			t.Errorf("Expected line %d, got %d", expectedLines[i], loc.LineNumber)
		}
	}
}
