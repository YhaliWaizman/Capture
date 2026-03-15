package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/yhaliwaizman/capture/internal/types"
)

// randomVarName generates a random environment variable name.
func randomVarName(r *rand.Rand) string {
	length := r.Intn(8) + 1
	b := make([]byte, length)
	for i := range b {
		b[i] = byte('A' + r.Intn(26))
	}
	return string(b)
}

// randomLocation generates a random Location.
func randomLocation(r *rand.Rand) types.Location {
	pathLen := r.Intn(5) + 1
	pathBytes := make([]byte, pathLen)
	for i := range pathBytes {
		pathBytes[i] = byte('a' + r.Intn(26))
	}
	return types.Location{
		FilePath:   string(pathBytes) + ".go",
		LineNumber: r.Intn(500) + 1,
	}
}

// generateReportData creates a random ReportData for property testing.
func generateReportData(r *rand.Rand) types.ReportData {
	data := types.ReportData{}

	// Unused (0-3 entries)
	n := r.Intn(4)
	if n > 0 {
		data.Unused = make([]string, n)
		for i := range data.Unused {
			data.Unused[i] = randomVarName(r)
		}
	}

	// Missing (0-3 entries)
	n = r.Intn(4)
	if n > 0 {
		data.Missing = make(map[string]types.Location, n)
		for i := 0; i < n; i++ {
			data.Missing[randomVarName(r)] = randomLocation(r)
		}
	}

	// CodeUsesNotInDocker (0-3 entries)
	n = r.Intn(4)
	if n > 0 {
		data.CodeUsesNotInDocker = make(map[string][]types.Location, n)
		for i := 0; i < n; i++ {
			locCount := r.Intn(3) + 1
			locs := make([]types.Location, locCount)
			for j := range locs {
				locs[j] = randomLocation(r)
			}
			data.CodeUsesNotInDocker[randomVarName(r)] = locs
		}
	}

	// DockerDeclaresUnused (0-3 entries)
	n = r.Intn(4)
	if n > 0 {
		data.DockerDeclaresUnused = make([]string, n)
		for i := range data.DockerDeclaresUnused {
			data.DockerDeclaresUnused[i] = randomVarName(r)
		}
	}

	// DockerUsesUndeclared (0-3 entries)
	n = r.Intn(4)
	if n > 0 {
		data.DockerUsesUndeclared = make(map[string]types.Location, n)
		for i := 0; i < n; i++ {
			data.DockerUsesUndeclared[randomVarName(r)] = randomLocation(r)
		}
	}

	return data
}

// reportDataGenValue wraps ReportData for testing/quick.
type reportDataGenValue struct {
	Data types.ReportData
}

// Generate implements quick.Generator for reportDataGenValue.
func (reportDataGenValue) Generate(r *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(reportDataGenValue{Data: generateReportData(r)})
}

// sarifFromData is a helper that runs ReportSARIF and returns the raw output.
func sarifFromData(data types.ReportData) ([]byte, error) {
	var out, errBuf bytes.Buffer
	r := NewReporter(&out, &errBuf)
	if err := r.ReportSARIF(data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// Feature: sarif-output-format, Property 7: SARIF Serialization Round-Trip
// Validates: Requirements 10.7, 6.3
func TestProperty_SARIFRoundTripValidity(t *testing.T) {
	f := func(input reportDataGenValue) bool {
		raw, err := sarifFromData(input.Data)
		if err != nil {
			t.Logf("ReportSARIF error: %v", err)
			return false
		}

		// Deserialize back into SARIFDocument
		var doc types.SARIFDocument
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Logf("Unmarshal error: %v", err)
			return false
		}

		// Verify structural validity
		if doc.Version != "2.1.0" {
			t.Logf("version = %q, want 2.1.0", doc.Version)
			return false
		}
		if doc.Schema == "" {
			t.Log("schema is empty")
			return false
		}
		if len(doc.Runs) != 1 {
			t.Logf("runs count = %d, want 1", len(doc.Runs))
			return false
		}
		run := doc.Runs[0]
		if run.Tool.Driver.Name != "capture" {
			t.Logf("driver name = %q, want capture", run.Tool.Driver.Name)
			return false
		}
		if run.Tool.Driver.InformationURI == "" {
			t.Log("informationUri is empty")
			return false
		}
		if run.Results == nil {
			t.Log("results is nil")
			return false
		}
		if run.Tool.Driver.Rules == nil && len(run.Results) > 0 {
			t.Log("rules is nil but results exist")
			return false
		}
		return true
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Feature: sarif-output-format, Property 6: Output Idempotence
// Validates: Requirements 7.3, 10.6
func TestProperty_SARIFDeterministicOutput(t *testing.T) {
	f := func(input reportDataGenValue) bool {
		out1, err1 := sarifFromData(input.Data)
		if err1 != nil {
			t.Logf("First call error: %v", err1)
			return false
		}

		out2, err2 := sarifFromData(input.Data)
		if err2 != nil {
			t.Logf("Second call error: %v", err2)
			return false
		}

		if !bytes.Equal(out1, out2) {
			t.Log("Outputs differ between two calls with identical input")
			return false
		}
		return true
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Feature: sarif-output-format, Property 2: Rule-Result Correspondence
// Validates: Requirements 3.7
func TestProperty_SARIFRuleResultCorrespondence(t *testing.T) {
	f := func(input reportDataGenValue) bool {
		raw, err := sarifFromData(input.Data)
		if err != nil {
			t.Logf("ReportSARIF error: %v", err)
			return false
		}

		var doc types.SARIFDocument
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Logf("Unmarshal error: %v", err)
			return false
		}

		run := doc.Runs[0]
		rules := run.Tool.Driver.Rules
		results := run.Results

		// Build set of ruleIds that appear in results
		resultRuleIDs := make(map[string]bool)
		for _, r := range results {
			resultRuleIDs[r.RuleID] = true
		}

		// Every rule in rules array must have at least one result referencing it
		for _, rule := range rules {
			if !resultRuleIDs[rule.ID] {
				t.Logf("Rule %q in rules array but no result references it", rule.ID)
				return false
			}
		}

		// Every ruleId in results must have a corresponding rule in the rules array
		ruleIDs := make(map[string]bool)
		for _, rule := range rules {
			ruleIDs[rule.ID] = true
		}
		for _, r := range results {
			if !ruleIDs[r.RuleID] {
				t.Logf("Result references ruleId %q but no rule exists for it", r.RuleID)
				return false
			}
		}

		// Verify ruleIndex is valid for each result
		for i, r := range results {
			if r.RuleIndex < 0 || r.RuleIndex >= len(rules) {
				t.Logf("Result[%d] ruleIndex %d out of range [0, %d)", i, r.RuleIndex, len(rules))
				return false
			}
			if rules[r.RuleIndex].ID != r.RuleID {
				t.Logf("Result[%d] ruleIndex %d points to %q, expected %q",
					i, r.RuleIndex, rules[r.RuleIndex].ID, r.RuleID)
				return false
			}
		}

		return true
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Ensure reportDataGenValue is used (avoid unused import for fmt if needed)
var _ = fmt.Sprintf
