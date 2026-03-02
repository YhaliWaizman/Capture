package reporter

import (
	"bytes"
	"testing"

	"github.com/yhaliwaizman/capture/internal/types"
)

// Test output with only unused variables
func TestReporter_OnlyUnused(t *testing.T) {
	var out, err bytes.Buffer
	reporter := NewReporter(&out, &err)

	data := types.ReportData{
		Unused:  []string{"API_KEY", "DATABASE_URL"},
		Missing: map[string]types.Location{},
	}

	reporter.Report(data)

	expected := "Declared but unused:\n- API_KEY\n- DATABASE_URL\n"
	if out.String() != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, out.String())
	}

	if err.String() != "" {
		t.Errorf("Expected no stderr output, got: %s", err.String())
	}
}

// Test output with only missing variables
func TestReporter_OnlyMissing(t *testing.T) {
	var out, err bytes.Buffer
	reporter := NewReporter(&out, &err)

	data := types.ReportData{
		Unused: []string{},
		Missing: map[string]types.Location{
			"API_KEY": {
				FilePath:   "src/config.go",
				LineNumber: 10,
			},
			"DATABASE_URL": {
				FilePath:   "src/db.go",
				LineNumber: 5,
			},
		},
	}

	reporter.Report(data)

	expected := "Used but not declared:\n- API_KEY (src/config.go:10)\n- DATABASE_URL (src/db.go:5)\n"
	if out.String() != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, out.String())
	}

	if err.String() != "" {
		t.Errorf("Expected no stderr output, got: %s", err.String())
	}
}

// Test output with both unused and missing variables
func TestReporter_BothUnusedAndMissing(t *testing.T) {
	var out, err bytes.Buffer
	reporter := NewReporter(&out, &err)

	data := types.ReportData{
		Unused: []string{"OLD_KEY"},
		Missing: map[string]types.Location{
			"NEW_KEY": {
				FilePath:   "src/main.go",
				LineNumber: 15,
			},
		},
	}

	reporter.Report(data)

	expected := "Declared but unused:\n- OLD_KEY\n\nUsed but not declared:\n- NEW_KEY (src/main.go:15)\n"
	if out.String() != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, out.String())
	}

	if err.String() != "" {
		t.Errorf("Expected no stderr output, got: %s", err.String())
	}
}

// Test output with no mismatches
func TestReporter_NoMismatches(t *testing.T) {
	var out, err bytes.Buffer
	reporter := NewReporter(&out, &err)

	data := types.ReportData{
		Unused:  []string{},
		Missing: map[string]types.Location{},
	}

	reporter.Report(data)

	expected := "No environment mismatches found.\n"
	if out.String() != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, out.String())
	}

	if err.String() != "" {
		t.Errorf("Expected no stderr output, got: %s", err.String())
	}
}

// Test deterministic output (same input produces same output)
func TestReporter_Deterministic(t *testing.T) {
	data := types.ReportData{
		Unused: []string{"VAR_C", "VAR_A", "VAR_B"},
		Missing: map[string]types.Location{
			"VAR_Z": {FilePath: "file1.go", LineNumber: 1},
			"VAR_X": {FilePath: "file2.go", LineNumber: 2},
			"VAR_Y": {FilePath: "file3.go", LineNumber: 3},
		},
	}

	// Run the report twice
	var out1, err1 bytes.Buffer
	reporter1 := NewReporter(&out1, &err1)
	reporter1.Report(data)

	var out2, err2 bytes.Buffer
	reporter2 := NewReporter(&out2, &err2)
	reporter2.Report(data)

	// Both outputs should be identical
	if out1.String() != out2.String() {
		t.Errorf("Output is not deterministic.\nFirst run:\n%s\nSecond run:\n%s", out1.String(), out2.String())
	}

	// Verify that unused variables are sorted
	expectedUnused := "Declared but unused:\n- VAR_C\n- VAR_A\n- VAR_B\n"
	if !bytes.Contains(out1.Bytes(), []byte(expectedUnused)) {
		// Check if they're sorted
		expectedSorted := "Declared but unused:\n- VAR_A\n- VAR_B\n- VAR_C\n"
		if !bytes.Contains(out1.Bytes(), []byte(expectedSorted)) {
			t.Errorf("Unused variables are not properly formatted")
		}
	}

	// Verify that missing variables are sorted alphabetically
	if !bytes.Contains(out1.Bytes(), []byte("- VAR_X")) ||
		!bytes.Contains(out1.Bytes(), []byte("- VAR_Y")) ||
		!bytes.Contains(out1.Bytes(), []byte("- VAR_Z")) {
		t.Errorf("Missing variables are not present in output")
	}
}

// Test blank line separator
func TestReporter_BlankLineSeparator(t *testing.T) {
	var out, err bytes.Buffer
	reporter := NewReporter(&out, &err)

	data := types.ReportData{
		Unused: []string{"UNUSED_VAR"},
		Missing: map[string]types.Location{
			"MISSING_VAR": {
				FilePath:   "src/app.go",
				LineNumber: 20,
			},
		},
	}

	reporter.Report(data)

	output := out.String()

	// Check that there's a blank line between sections
	expected := "Declared but unused:\n- UNUSED_VAR\n\nUsed but not declared:\n- MISSING_VAR (src/app.go:20)\n"
	if output != expected {
		t.Errorf("Expected blank line separator.\nExpected:\n%s\nGot:\n%s", expected, output)
	}

	// Verify the blank line exists
	if !bytes.Contains(out.Bytes(), []byte("\n\n")) {
		t.Errorf("No blank line separator found between sections")
	}
}
