package reporter

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/yhaliwaizman/capture/internal/types"
)

// helper to run ReportSARIF and return parsed document
func runSARIF(t *testing.T, data types.ReportData) (types.SARIFDocument, string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	r := NewReporter(&out, &errBuf)
	if err := r.ReportSARIF(data); err != nil {
		t.Fatalf("ReportSARIF returned error: %v", err)
	}
	raw := out.String()
	var doc types.SARIFDocument
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("Failed to unmarshal SARIF output: %v\nOutput:\n%s", err, raw)
	}
	return doc, raw
}

// Task 2.1: ReportSARIF writes to r.out and returns error on marshal failure
func TestReportSARIF_WritesToOut(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := NewReporter(&out, &errBuf)
	err := r.ReportSARIF(types.ReportData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("expected output, got empty")
	}
}

// Task 2.1: Verify SARIF envelope structure
func TestReportSARIF_EnvelopeStructure(t *testing.T) {
	doc, _ := runSARIF(t, types.ReportData{})

	if doc.Version != "2.1.0" {
		t.Errorf("version = %q, want %q", doc.Version, "2.1.0")
	}
	if doc.Schema != "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json" {
		t.Errorf("schema = %q, want SARIF 2.1.0 schema URL", doc.Schema)
	}
	if len(doc.Runs) != 1 {
		t.Fatalf("runs count = %d, want 1", len(doc.Runs))
	}
	if doc.Runs[0].Tool.Driver.Name != "capture" {
		t.Errorf("driver name = %q, want %q", doc.Runs[0].Tool.Driver.Name, "capture")
	}
	if doc.Runs[0].Tool.Driver.InformationURI == "" {
		t.Error("informationUri is empty")
	}
}

// Task 2.2: Rule generation - only rules with results are included
func TestReportSARIF_RuleFiltering(t *testing.T) {
	data := types.ReportData{
		Unused: []string{"FOO"},
		Missing: map[string]types.Location{
			"BAR": {FilePath: "app.go", LineNumber: 5},
		},
	}
	doc, _ := runSARIF(t, data)
	rules := doc.Runs[0].Tool.Driver.Rules

	if len(rules) != 2 {
		t.Fatalf("rules count = %d, want 2", len(rules))
	}
	if rules[0].ID != "ENV001" {
		t.Errorf("rules[0].ID = %q, want ENV001", rules[0].ID)
	}
	if rules[1].ID != "ENV002" {
		t.Errorf("rules[1].ID = %q, want ENV002", rules[1].ID)
	}
}

// Task 2.2: All 5 rules present when all categories populated
func TestReportSARIF_AllFiveRules(t *testing.T) {
	data := types.ReportData{
		Unused: []string{"A"},
		Missing: map[string]types.Location{
			"B": {FilePath: "f.go", LineNumber: 1},
		},
		CodeUsesNotInDocker: map[string][]types.Location{
			"C": {{FilePath: "g.go", LineNumber: 2}},
		},
		DockerDeclaresUnused: []string{"D"},
		DockerUsesUndeclared: map[string]types.Location{
			"E": {FilePath: "Dockerfile", LineNumber: 3},
		},
	}
	doc, _ := runSARIF(t, data)
	rules := doc.Runs[0].Tool.Driver.Rules

	if len(rules) != 5 {
		t.Fatalf("rules count = %d, want 5", len(rules))
	}
	expectedIDs := []string{"ENV001", "ENV002", "ENV003", "ENV004", "ENV005"}
	for i, id := range expectedIDs {
		if rules[i].ID != id {
			t.Errorf("rules[%d].ID = %q, want %q", i, rules[i].ID, id)
		}
		if rules[i].Name == "" {
			t.Errorf("rules[%d].Name is empty", i)
		}
		if rules[i].ShortDescription.Text == "" {
			t.Errorf("rules[%d].ShortDescription.Text is empty", i)
		}
		if rules[i].HelpURI == "" {
			t.Errorf("rules[%d].HelpURI is empty", i)
		}
	}
}

// Task 2.2: Rules sorted by id
func TestReportSARIF_RulesSortedByID(t *testing.T) {
	data := types.ReportData{
		DockerUsesUndeclared: map[string]types.Location{
			"Z": {FilePath: "Dockerfile", LineNumber: 1},
		},
		Unused: []string{"A"},
	}
	doc, _ := runSARIF(t, data)
	rules := doc.Runs[0].Tool.Driver.Rules

	if len(rules) != 2 {
		t.Fatalf("rules count = %d, want 2", len(rules))
	}
	if rules[0].ID != "ENV001" || rules[1].ID != "ENV005" {
		t.Errorf("rules not sorted: got %s, %s", rules[0].ID, rules[1].ID)
	}
}

// Task 2.3: Result mapping - correct ruleId and level
func TestReportSARIF_ResultMapping(t *testing.T) {
	data := types.ReportData{
		Unused: []string{"UNUSED_VAR"},
		Missing: map[string]types.Location{
			"MISSING_VAR": {FilePath: "app.go", LineNumber: 10},
		},
		CodeUsesNotInDocker: map[string][]types.Location{
			"CODE_VAR": {{FilePath: "src.go", LineNumber: 20}},
		},
		DockerDeclaresUnused: []string{"DOCKER_UNUSED"},
		DockerUsesUndeclared: map[string]types.Location{
			"DOCKER_UNDECL": {FilePath: "Dockerfile", LineNumber: 30},
		},
	}
	doc, _ := runSARIF(t, data)
	results := doc.Runs[0].Results

	if len(results) != 5 {
		t.Fatalf("results count = %d, want 5", len(results))
	}

	// Results should be sorted by ruleId
	expected := []struct {
		ruleID string
		level  string
	}{
		{"ENV001", "warning"},
		{"ENV002", "error"},
		{"ENV003", "warning"},
		{"ENV004", "warning"},
		{"ENV005", "error"},
	}
	for i, exp := range expected {
		if results[i].RuleID != exp.ruleID {
			t.Errorf("results[%d].RuleID = %q, want %q", i, results[i].RuleID, exp.ruleID)
		}
		if results[i].Level != exp.level {
			t.Errorf("results[%d].Level = %q, want %q", i, results[i].Level, exp.level)
		}
		if results[i].Message.Text == "" {
			t.Errorf("results[%d].Message.Text is empty", i)
		}
	}
}

// Task 2.3: ruleIndex references correct rule in filtered array
func TestReportSARIF_RuleIndex(t *testing.T) {
	data := types.ReportData{
		Missing: map[string]types.Location{
			"X": {FilePath: "a.go", LineNumber: 1},
		},
		DockerDeclaresUnused: []string{"Y"},
	}
	doc, _ := runSARIF(t, data)
	rules := doc.Runs[0].Tool.Driver.Rules
	results := doc.Runs[0].Results

	for _, r := range results {
		if r.RuleIndex < 0 || r.RuleIndex >= len(rules) {
			t.Fatalf("ruleIndex %d out of range [0, %d)", r.RuleIndex, len(rules))
		}
		if rules[r.RuleIndex].ID != r.RuleID {
			t.Errorf("ruleIndex %d points to rule %q, but result ruleId is %q",
				r.RuleIndex, rules[r.RuleIndex].ID, r.RuleID)
		}
	}
}

// Task 2.3: Results sorted by ruleId then variable name
func TestReportSARIF_ResultsSorted(t *testing.T) {
	data := types.ReportData{
		Unused: []string{"ZEBRA", "ALPHA", "MIDDLE"},
	}
	doc, _ := runSARIF(t, data)
	results := doc.Runs[0].Results

	if len(results) != 3 {
		t.Fatalf("results count = %d, want 3", len(results))
	}
	// All ENV001, sorted by variable name in message
	for i := 1; i < len(results); i++ {
		if results[i].Message.Text < results[i-1].Message.Text {
			t.Errorf("results not sorted: %q before %q", results[i-1].Message.Text, results[i].Message.Text)
		}
	}
}

// Task 2.4: Physical location mapping for ENV002
func TestReportSARIF_LocationENV002(t *testing.T) {
	data := types.ReportData{
		Missing: map[string]types.Location{
			"DB_URL": {FilePath: "src/db.go", LineNumber: 42},
		},
	}
	doc, _ := runSARIF(t, data)
	results := doc.Runs[0].Results

	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	r := results[0]
	if len(r.Locations) != 1 {
		t.Fatalf("locations count = %d, want 1", len(r.Locations))
	}
	loc := r.Locations[0].PhysicalLocation
	if loc.ArtifactLocation.URI != "src/db.go" {
		t.Errorf("uri = %q, want %q", loc.ArtifactLocation.URI, "src/db.go")
	}
	if loc.Region.StartLine != 42 {
		t.Errorf("startLine = %d, want 42", loc.Region.StartLine)
	}
}

// Task 2.4: Physical location mapping for ENV003 (first location)
func TestReportSARIF_LocationENV003(t *testing.T) {
	data := types.ReportData{
		CodeUsesNotInDocker: map[string][]types.Location{
			"API_KEY": {
				{FilePath: "first.go", LineNumber: 10},
				{FilePath: "second.go", LineNumber: 20},
			},
		},
	}
	doc, _ := runSARIF(t, data)
	r := doc.Runs[0].Results[0]
	if len(r.Locations) != 1 {
		t.Fatalf("locations count = %d, want 1", len(r.Locations))
	}
	if r.Locations[0].PhysicalLocation.ArtifactLocation.URI != "first.go" {
		t.Errorf("uri = %q, want first.go", r.Locations[0].PhysicalLocation.ArtifactLocation.URI)
	}
}

// Task 2.4: Physical location mapping for ENV005
func TestReportSARIF_LocationENV005(t *testing.T) {
	data := types.ReportData{
		DockerUsesUndeclared: map[string]types.Location{
			"SECRET": {FilePath: "Dockerfile", LineNumber: 7},
		},
	}
	doc, _ := runSARIF(t, data)
	r := doc.Runs[0].Results[0]
	if len(r.Locations) != 1 {
		t.Fatalf("locations count = %d, want 1", len(r.Locations))
	}
	loc := r.Locations[0].PhysicalLocation
	if loc.ArtifactLocation.URI != "Dockerfile" {
		t.Errorf("uri = %q, want Dockerfile", loc.ArtifactLocation.URI)
	}
	if loc.Region.StartLine != 7 {
		t.Errorf("startLine = %d, want 7", loc.Region.StartLine)
	}
}

// Task 2.4: ENV001 and ENV004 omit locations
func TestReportSARIF_NoLocationsForENV001AndENV004(t *testing.T) {
	data := types.ReportData{
		Unused:               []string{"VAR1"},
		DockerDeclaresUnused: []string{"VAR2"},
	}
	doc, _ := runSARIF(t, data)
	for _, r := range doc.Runs[0].Results {
		if len(r.Locations) != 0 {
			t.Errorf("result %s should have no locations, got %d", r.RuleID, len(r.Locations))
		}
	}
}

// Task 2.4: Forward-slash path separators
func TestReportSARIF_ForwardSlashPaths(t *testing.T) {
	data := types.ReportData{
		Missing: map[string]types.Location{
			"VAR": {FilePath: "src/config/db.go", LineNumber: 1},
		},
	}
	doc, _ := runSARIF(t, data)
	uri := doc.Runs[0].Results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI
	if uri != "src/config/db.go" {
		t.Errorf("uri = %q, want forward-slash path", uri)
	}
}

// Task 2.5: Empty results handling
func TestReportSARIF_EmptyResults(t *testing.T) {
	doc, raw := runSARIF(t, types.ReportData{})

	if len(doc.Runs[0].Results) != 0 {
		t.Errorf("results count = %d, want 0", len(doc.Runs[0].Results))
	}
	if len(doc.Runs[0].Tool.Driver.Rules) != 0 {
		t.Errorf("rules count = %d, want 0", len(doc.Runs[0].Tool.Driver.Rules))
	}
	// Verify it's valid JSON
	if !json.Valid([]byte(raw[:len(raw)-1])) { // trim trailing newline
		t.Error("output is not valid JSON")
	}
}

// Task 2.5: Nil slices in ReportData produce valid SARIF
func TestReportSARIF_NilSlicesProduceValidSARIF(t *testing.T) {
	data := types.ReportData{
		Unused:               nil,
		Missing:              nil,
		CodeUsesNotInDocker:  nil,
		DockerDeclaresUnused: nil,
		DockerUsesUndeclared: nil,
	}
	doc, _ := runSARIF(t, data)
	if doc.Runs[0].Results == nil {
		t.Error("results should be non-nil empty slice")
	}
}

// Task 2.6: Deterministic output
func TestReportSARIF_Deterministic(t *testing.T) {
	data := types.ReportData{
		Unused: []string{"Z_VAR", "A_VAR", "M_VAR"},
		Missing: map[string]types.Location{
			"X": {FilePath: "a.go", LineNumber: 1},
			"B": {FilePath: "b.go", LineNumber: 2},
		},
		CodeUsesNotInDocker: map[string][]types.Location{
			"Q": {{FilePath: "q.go", LineNumber: 3}},
		},
		DockerDeclaresUnused: []string{"W", "D"},
		DockerUsesUndeclared: map[string]types.Location{
			"R": {FilePath: "Dockerfile", LineNumber: 4},
		},
	}

	var out1, err1 bytes.Buffer
	r1 := NewReporter(&out1, &err1)
	if err := r1.ReportSARIF(data); err != nil {
		t.Fatal(err)
	}

	var out2, err2 bytes.Buffer
	r2 := NewReporter(&out2, &err2)
	if err := r2.ReportSARIF(data); err != nil {
		t.Fatal(err)
	}

	if out1.String() != out2.String() {
		t.Error("output is not deterministic")
	}
}

// Task 2.6: 2-space indentation and trailing newline
func TestReportSARIF_IndentationAndTrailingNewline(t *testing.T) {
	var out, errBuf bytes.Buffer
	r := NewReporter(&out, &errBuf)
	if err := r.ReportSARIF(types.ReportData{}); err != nil {
		t.Fatal(err)
	}
	raw := out.String()
	// Trailing newline
	if raw[len(raw)-1] != '\n' {
		t.Error("output does not end with newline")
	}
	// 2-space indentation (check for "  " prefix on indented lines)
	if !bytes.Contains(out.Bytes(), []byte("\n  \"")) {
		t.Error("output does not use 2-space indentation")
	}
}
