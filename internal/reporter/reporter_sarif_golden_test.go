package reporter

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/yhaliwaizman/capture/internal/types"
)

// TestReportSARIF_GoldenFull compares ReportSARIF output for a full scan
// (all 5 mismatch categories) against the golden file for regression detection.
func TestReportSARIF_GoldenFull(t *testing.T) {
	data := types.ReportData{
		Unused: []string{"UNUSED_VAR"},
		Missing: map[string]types.Location{
			"MISSING_VAR": {FilePath: "src/app.go", LineNumber: 10},
		},
		CodeUsesNotInDocker: map[string][]types.Location{
			"CODE_VAR": {{FilePath: "src/config.go", LineNumber: 20}},
		},
		DockerDeclaresUnused: []string{"DOCKER_UNUSED"},
		DockerUsesUndeclared: map[string]types.Location{
			"DOCKER_UNDECL": {FilePath: "Dockerfile", LineNumber: 5},
		},
	}

	var buf bytes.Buffer
	r := &ReporterImpl{out: &buf}
	if err := r.ReportSARIF(data); err != nil {
		t.Fatalf("ReportSARIF: %v", err)
	}

	golden, err := os.ReadFile(filepath.Join("testdata", "sarif_expected_full.json"))
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), golden) {
		t.Errorf("output does not match golden file testdata/sarif_expected_full.json\n--- got ---\n%s\n--- want ---\n%s", buf.String(), string(golden))
	}
}

// TestReportSARIF_GoldenEmpty compares ReportSARIF output for an empty scan
// (no mismatches) against the golden file for regression detection.
func TestReportSARIF_GoldenEmpty(t *testing.T) {
	data := types.ReportData{}

	var buf bytes.Buffer
	r := &ReporterImpl{out: &buf}
	if err := r.ReportSARIF(data); err != nil {
		t.Fatalf("ReportSARIF: %v", err)
	}

	golden, err := os.ReadFile(filepath.Join("testdata", "sarif_expected_empty.json"))
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), golden) {
		t.Errorf("output does not match golden file testdata/sarif_expected_empty.json\n--- got ---\n%s\n--- want ---\n%s", buf.String(), string(golden))
	}
}
