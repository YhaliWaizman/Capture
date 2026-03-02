package types

import "testing"

func TestLocationStruct(t *testing.T) {
	loc := Location{
		FilePath:   "src/main.go",
		LineNumber: 42,
	}

	if loc.FilePath != "src/main.go" {
		t.Errorf("Expected FilePath to be 'src/main.go', got '%s'", loc.FilePath)
	}

	if loc.LineNumber != 42 {
		t.Errorf("Expected LineNumber to be 42, got %d", loc.LineNumber)
	}
}

func TestDiffResultStruct(t *testing.T) {
	result := DiffResult{
		Unused:  []string{"VAR1", "VAR2"},
		Missing: []string{"VAR3"},
	}

	if len(result.Unused) != 2 {
		t.Errorf("Expected 2 unused variables, got %d", len(result.Unused))
	}

	if len(result.Missing) != 1 {
		t.Errorf("Expected 1 missing variable, got %d", len(result.Missing))
	}
}

func TestReportDataStruct(t *testing.T) {
	data := ReportData{
		Unused: []string{"VAR1"},
		Missing: map[string]Location{
			"VAR2": {FilePath: "src/app.go", LineNumber: 10},
		},
	}

	if len(data.Unused) != 1 {
		t.Errorf("Expected 1 unused variable, got %d", len(data.Unused))
	}

	if len(data.Missing) != 1 {
		t.Errorf("Expected 1 missing variable, got %d", len(data.Missing))
	}

	loc, exists := data.Missing["VAR2"]
	if !exists {
		t.Error("Expected VAR2 to exist in Missing map")
	}

	if loc.FilePath != "src/app.go" {
		t.Errorf("Expected FilePath to be 'src/app.go', got '%s'", loc.FilePath)
	}
}
