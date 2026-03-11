package reporter

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/yhaliwaizman/capture/internal/types"
)

func TestReporterImpl_ReportJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    types.ReportData
		wantErr bool
	}{
		{
			name: "no mismatches",
			data: types.ReportData{
				Unused:               []string{},
				Missing:              map[string]types.Location{},
				AllLocations:         map[string][]types.Location{},
				FilesScanned:         10,
				VariablesDeclared:    5,
				VariablesUsed:        5,
				CodeUsesNotInDocker:  map[string][]types.Location{},
				DockerDeclaresUnused: []string{},
				DockerUsesUndeclared: map[string]types.Location{},
			},
			wantErr: false,
		},
		{
			name: "with unused variables",
			data: types.ReportData{
				Unused:               []string{"OLD_API_KEY", "DEPRECATED_URL"},
				Missing:              map[string]types.Location{},
				AllLocations:         map[string][]types.Location{},
				FilesScanned:         10,
				VariablesDeclared:    7,
				VariablesUsed:        5,
				CodeUsesNotInDocker:  map[string][]types.Location{},
				DockerDeclaresUnused: []string{},
				DockerUsesUndeclared: map[string]types.Location{},
			},
			wantErr: false,
		},
		{
			name: "with missing variables",
			data: types.ReportData{
				Unused: []string{},
				Missing: map[string]types.Location{
					"DATABASE_URL": {FilePath: "src/db.go", LineNumber: 15},
				},
				AllLocations: map[string][]types.Location{
					"DATABASE_URL": {
						{FilePath: "src/db.go", LineNumber: 15},
						{FilePath: "src/db.go", LineNumber: 20},
					},
				},
				FilesScanned:         10,
				VariablesDeclared:    5,
				VariablesUsed:        6,
				CodeUsesNotInDocker:  map[string][]types.Location{},
				DockerDeclaresUnused: []string{},
				DockerUsesUndeclared: map[string]types.Location{},
			},
			wantErr: false,
		},
		{
			name: "with dockerfile issues",
			data: types.ReportData{
				Unused:  []string{},
				Missing: map[string]types.Location{},
				AllLocations: map[string][]types.Location{
					"MISSING_VAR": {
						{FilePath: "src/app.js", LineNumber: 10},
					},
				},
				FilesScanned:      10,
				VariablesDeclared: 5,
				VariablesUsed:     6,
				CodeUsesNotInDocker: map[string][]types.Location{
					"MISSING_VAR": {
						{FilePath: "src/app.js", LineNumber: 10},
					},
				},
				DockerDeclaresUnused: []string{"BUILD_VERSION"},
				DockerUsesUndeclared: map[string]types.Location{
					"UNDEFINED_VAR": {FilePath: "Dockerfile", LineNumber: 15},
				},
			},
			wantErr: false,
		},
		{
			name: "complete example",
			data: types.ReportData{
				Unused: []string{"OLD_API_KEY", "DEPRECATED_URL"},
				Missing: map[string]types.Location{
					"DATABASE_URL": {FilePath: "src/db.go", LineNumber: 15},
				},
				AllLocations: map[string][]types.Location{
					"DATABASE_URL": {
						{FilePath: "src/db.go", LineNumber: 15},
					},
					"MISSING_VAR": {
						{FilePath: "src/app.js", LineNumber: 10},
					},
				},
				FilesScanned:      42,
				VariablesDeclared: 15,
				VariablesUsed:     18,
				CodeUsesNotInDocker: map[string][]types.Location{
					"MISSING_VAR": {
						{FilePath: "src/app.js", LineNumber: 10},
					},
				},
				DockerDeclaresUnused: []string{"BUILD_VERSION"},
				DockerUsesUndeclared: map[string]types.Location{
					"UNDEFINED_VAR": {FilePath: "Dockerfile", LineNumber: 15},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var errOut bytes.Buffer
			r := NewReporter(&out, &errOut)

			err := r.ReportJSON(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify JSON is valid
			var result types.JSONOutput
			if err := json.Unmarshal(out.Bytes(), &result); err != nil {
				t.Errorf("ReportJSON() produced invalid JSON: %v", err)
				return
			}

			// Verify summary
			if result.Summary.FilesScanned != tt.data.FilesScanned {
				t.Errorf("FilesScanned = %d, want %d", result.Summary.FilesScanned, tt.data.FilesScanned)
			}
			if result.Summary.VariablesDeclared != tt.data.VariablesDeclared {
				t.Errorf("VariablesDeclared = %d, want %d", result.Summary.VariablesDeclared, tt.data.VariablesDeclared)
			}
			if result.Summary.VariablesUsed != tt.data.VariablesUsed {
				t.Errorf("VariablesUsed = %d, want %d", result.Summary.VariablesUsed, tt.data.VariablesUsed)
			}

			// Verify unused
			if len(result.Unused) != len(tt.data.Unused) {
				t.Errorf("Unused count = %d, want %d", len(result.Unused), len(tt.data.Unused))
			}

			// Verify missing
			if len(result.Missing) != len(tt.data.Missing) {
				t.Errorf("Missing count = %d, want %d", len(result.Missing), len(tt.data.Missing))
			}

			// Verify dockerfile issues
			if len(result.DockerfileIssues.CodeUsesNotInDocker) != len(tt.data.CodeUsesNotInDocker) {
				t.Errorf("CodeUsesNotInDocker count = %d, want %d",
					len(result.DockerfileIssues.CodeUsesNotInDocker), len(tt.data.CodeUsesNotInDocker))
			}
			if len(result.DockerfileIssues.DockerDeclaresUnused) != len(tt.data.DockerDeclaresUnused) {
				t.Errorf("DockerDeclaresUnused count = %d, want %d",
					len(result.DockerfileIssues.DockerDeclaresUnused), len(tt.data.DockerDeclaresUnused))
			}
			if len(result.DockerfileIssues.DockerUsesUndeclared) != len(tt.data.DockerUsesUndeclared) {
				t.Errorf("DockerUsesUndeclared count = %d, want %d",
					len(result.DockerfileIssues.DockerUsesUndeclared), len(tt.data.DockerUsesUndeclared))
			}
		})
	}
}

func TestReporterImpl_ReportJSON_OutputStructure(t *testing.T) {
	data := types.ReportData{
		Unused: []string{"OLD_API_KEY"},
		Missing: map[string]types.Location{
			"DATABASE_URL": {FilePath: "src/db.go", LineNumber: 15},
		},
		AllLocations: map[string][]types.Location{
			"DATABASE_URL": {
				{FilePath: "src/db.go", LineNumber: 15},
				{FilePath: "src/db.go", LineNumber: 20},
			},
		},
		FilesScanned:      42,
		VariablesDeclared: 15,
		VariablesUsed:     16,
		CodeUsesNotInDocker: map[string][]types.Location{
			"MISSING_VAR": {
				{FilePath: "src/app.js", LineNumber: 10},
			},
		},
		DockerDeclaresUnused: []string{"BUILD_VERSION"},
		DockerUsesUndeclared: map[string]types.Location{
			"UNDEFINED_VAR": {FilePath: "Dockerfile", LineNumber: 15},
		},
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	r := NewReporter(&out, &errOut)

	err := r.ReportJSON(data)
	if err != nil {
		t.Fatalf("ReportJSON() error = %v", err)
	}

	var result types.JSONOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify structure matches expected format
	if result.Summary.MismatchesFound != 5 {
		t.Errorf("MismatchesFound = %d, want 5", result.Summary.MismatchesFound)
	}

	// Verify missing variable has all locations
	if len(result.Missing) != 1 {
		t.Fatalf("Missing length = %d, want 1", len(result.Missing))
	}
	if result.Missing[0].Variable != "DATABASE_URL" {
		t.Errorf("Missing[0].Variable = %s, want DATABASE_URL", result.Missing[0].Variable)
	}
	if len(result.Missing[0].Locations) != 2 {
		t.Errorf("Missing[0].Locations length = %d, want 2", len(result.Missing[0].Locations))
	}
}

func TestReporterImpl_ReportJSON_NilSliceHandling(t *testing.T) {
	// Test that nil slices are normalized to empty arrays in JSON output
	data := types.ReportData{
		Unused:               nil, // nil slice
		Missing:              map[string]types.Location{},
		AllLocations:         map[string][]types.Location{},
		FilesScanned:         10,
		VariablesDeclared:    5,
		VariablesUsed:        5,
		CodeUsesNotInDocker:  map[string][]types.Location{},
		DockerDeclaresUnused: nil, // nil slice
		DockerUsesUndeclared: map[string]types.Location{},
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	r := NewReporter(&out, &errOut)

	err := r.ReportJSON(data)
	if err != nil {
		t.Fatalf("ReportJSON() error = %v", err)
	}

	// Parse JSON output
	var result types.JSONOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v\nOutput: %s", err, out.String())
	}

	// Verify nil slices are marshaled as empty arrays, not null
	jsonStr := out.String()
	if bytes.Contains(out.Bytes(), []byte(`"unused": null`)) {
		t.Error("Expected 'unused' to be an empty array [], got null")
	}
	if bytes.Contains(out.Bytes(), []byte(`"docker_declares_unused": null`)) {
		t.Error("Expected 'docker_declares_unused' to be an empty array [], got null")
	}

	// Verify the parsed result has empty slices
	if result.Unused == nil {
		t.Error("Expected Unused to be empty slice, got nil")
	}
	if len(result.Unused) != 0 {
		t.Errorf("Expected Unused length 0, got %d", len(result.Unused))
	}
	if result.DockerfileIssues.DockerDeclaresUnused == nil {
		t.Error("Expected DockerDeclaresUnused to be empty slice, got nil")
	}
	if len(result.DockerfileIssues.DockerDeclaresUnused) != 0 {
		t.Errorf("Expected DockerDeclaresUnused length 0, got %d", len(result.DockerfileIssues.DockerDeclaresUnused))
	}

	t.Logf("JSON output:\n%s", jsonStr)
}

func TestReporterImpl_ReportJSON_NilLocations(t *testing.T) {
	// Test that nil locations in AllLocations are handled gracefully
	data := types.ReportData{
		Unused: []string{},
		Missing: map[string]types.Location{
			"MISSING_VAR": {FilePath: "app.js", LineNumber: 10},
		},
		AllLocations: map[string][]types.Location{
			// MISSING_VAR is not in AllLocations - should fallback to Missing location
		},
		FilesScanned:         10,
		VariablesDeclared:    5,
		VariablesUsed:        6,
		CodeUsesNotInDocker:  map[string][]types.Location{},
		DockerDeclaresUnused: []string{},
		DockerUsesUndeclared: map[string]types.Location{},
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	r := NewReporter(&out, &errOut)

	err := r.ReportJSON(data)
	if err != nil {
		t.Fatalf("ReportJSON() error = %v", err)
	}

	// Parse JSON output
	var result types.JSONOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v\nOutput: %s", err, out.String())
	}

	// Verify locations is not null
	if len(result.Missing) != 1 {
		t.Fatalf("Expected 1 missing variable, got %d", len(result.Missing))
	}

	missingVar := result.Missing[0]
	if missingVar.Variable != "MISSING_VAR" {
		t.Errorf("Expected variable MISSING_VAR, got %s", missingVar.Variable)
	}

	// Should have fallback location from Missing map
	if missingVar.Locations == nil {
		t.Error("Expected Locations to be non-nil")
	}
	if len(missingVar.Locations) != 1 {
		t.Errorf("Expected 1 location (fallback), got %d", len(missingVar.Locations))
	}
	if len(missingVar.Locations) > 0 {
		loc := missingVar.Locations[0]
		if loc.FilePath != "app.js" || loc.LineNumber != 10 {
			t.Errorf("Expected location app.js:10, got %s:%d", loc.FilePath, loc.LineNumber)
		}
	}

	// Verify JSON doesn't contain null for locations
	jsonStr := out.String()
	if bytes.Contains(out.Bytes(), []byte(`"locations": null`)) {
		t.Error("Expected 'locations' to be an array, got null")
	}

	t.Logf("JSON output:\n%s", jsonStr)
}
