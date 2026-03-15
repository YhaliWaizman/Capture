package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSARIFDocumentJSONKeys(t *testing.T) {
	doc := SARIFDocument{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "capture",
						InformationURI: "https://github.com/syncd-one/syncd",
						Rules: []SARIFReportingDescriptor{
							{
								ID:               "ENV001",
								Name:             "unused-variable",
								ShortDescription: SARIFMessage{Text: "Variable declared but not used"},
								HelpURI:          "https://github.com/syncd-one/syncd#unused-variable",
							},
						},
					},
				},
				Results: []SARIFResult{
					{
						RuleID:    "ENV002",
						RuleIndex: 0,
						Level:     "error",
						Message:   SARIFMessage{Text: "Missing variable: DB_HOST"},
						Locations: []SARIFLocation{
							{
								PhysicalLocation: SARIFPhysicalLocation{
									ArtifactLocation: SARIFArtifactLocation{URI: "src/main.go"},
									Region:           SARIFRegion{StartLine: 10},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal SARIFDocument: %v", err)
	}

	output := string(data)

	// Verify camelCase JSON keys from SARIF 2.1.0 spec
	expectedKeys := []string{
		`"$schema"`,
		`"informationUri"`,
		`"shortDescription"`,
		`"helpUri"`,
		`"ruleId"`,
		`"ruleIndex"`,
		`"physicalLocation"`,
		`"artifactLocation"`,
		`"startLine"`,
	}

	for _, key := range expectedKeys {
		if !strings.Contains(output, key) {
			t.Errorf("Expected JSON to contain key %s, but it was not found", key)
		}
	}
}

func TestSARIFResultOmitsLocationsWhenNil(t *testing.T) {
	result := SARIFResult{
		RuleID:    "ENV001",
		RuleIndex: 0,
		Level:     "warning",
		Message:   SARIFMessage{Text: "Unused variable: UNUSED_VAR"},
		Locations: nil,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal SARIFResult: %v", err)
	}

	output := string(data)

	if strings.Contains(output, `"locations"`) {
		t.Error("Expected 'locations' key to be omitted when Locations is nil")
	}
}

func TestSARIFResultIncludesLocationsWhenPopulated(t *testing.T) {
	result := SARIFResult{
		RuleID:    "ENV002",
		RuleIndex: 0,
		Level:     "error",
		Message:   SARIFMessage{Text: "Missing variable: API_KEY"},
		Locations: []SARIFLocation{
			{
				PhysicalLocation: SARIFPhysicalLocation{
					ArtifactLocation: SARIFArtifactLocation{URI: "src/config.go"},
					Region:           SARIFRegion{StartLine: 25},
				},
			},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal SARIFResult: %v", err)
	}

	output := string(data)

	if !strings.Contains(output, `"locations"`) {
		t.Error("Expected 'locations' key to be present when Locations is populated")
	}

	if !strings.Contains(output, `"src/config.go"`) {
		t.Error("Expected location URI 'src/config.go' in output")
	}

	if !strings.Contains(output, `"startLine":25`) {
		t.Error("Expected startLine 25 in output")
	}
}
