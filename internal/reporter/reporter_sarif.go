package reporter

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/yhaliwaizman/capture/internal/types"
)

const (
	sarifVersion = "2.1.0"
	sarifSchema  = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json"
	toolName     = "capture"
	toolInfoURI  = "https://github.com/syncd-one/syncd"
)

// sarifRuleDef holds the static definition for each rule.
type sarifRuleDef struct {
	id          string
	name        string
	description string
	level       string
}

var sarifRules = []sarifRuleDef{
	{id: "ENV001", name: "unused-variable", description: "Variable is declared in .env but not used in code", level: "warning"},
	{id: "ENV002", name: "missing-variable", description: "Variable is used in code but not declared in .env", level: "error"},
	{id: "ENV003", name: "code-uses-not-in-docker", description: "Variable is used in code but not declared in Dockerfile or .env", level: "warning"},
	{id: "ENV004", name: "docker-declares-unused", description: "Variable is declared in Dockerfile but not used in code", level: "warning"},
	{id: "ENV005", name: "docker-uses-undeclared", description: "Variable is used in Dockerfile but not declared", level: "error"},
}

// ReportSARIF formats and outputs the analysis results as SARIF 2.1.0 JSON.
func (r *ReporterImpl) ReportSARIF(data types.ReportData) error {
	results, activeRuleIDs := buildSARIFResults(data)
	rules := buildSARIFRules(activeRuleIDs)

	// Assign ruleIndex based on the filtered, sorted rules array
	ruleIndexMap := make(map[string]int, len(rules))
	for i, rule := range rules {
		ruleIndexMap[rule.ID] = i
	}
	for i := range results {
		results[i].RuleIndex = ruleIndexMap[results[i].RuleID]
	}

	doc := types.SARIFDocument{
		Version: sarifVersion,
		Schema:  sarifSchema,
		Runs: []types.SARIFRun{
			{
				Tool: types.SARIFTool{
					Driver: types.SARIFDriver{
						Name:           toolName,
						InformationURI: toolInfoURI,
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SARIF: %w", err)
	}

	fmt.Fprintln(r.out, string(jsonBytes))
	return nil
}

// buildSARIFRules returns the filtered, sorted rules array for the given active rule IDs.
func buildSARIFRules(activeRuleIDs map[string]bool) []types.SARIFReportingDescriptor {
	var rules []types.SARIFReportingDescriptor
	for _, def := range sarifRules {
		if !activeRuleIDs[def.id] {
			continue
		}
		rules = append(rules, types.SARIFReportingDescriptor{
			ID:   def.id,
			Name: def.name,
			ShortDescription: types.SARIFMessage{
				Text: def.description,
			},
			HelpURI: toolInfoURI + "#" + def.name,
		})
	}
	// sarifRules is already ordered by id, so rules inherits that order.
	// Explicit sort for safety.
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ID < rules[j].ID
	})
	return rules
}

// buildSARIFResults builds the sorted results slice and returns the set of active rule IDs.
func buildSARIFResults(data types.ReportData) ([]types.SARIFResult, map[string]bool) {
	activeRuleIDs := make(map[string]bool)
	var results []types.SARIFResult

	// ENV001 - unused variables (no locations)
	unused := data.Unused
	if unused == nil {
		unused = []string{}
	}
	sortedUnused := make([]string, len(unused))
	copy(sortedUnused, unused)
	sort.Strings(sortedUnused)
	for _, varName := range sortedUnused {
		activeRuleIDs["ENV001"] = true
		results = append(results, types.SARIFResult{
			RuleID:  "ENV001",
			Level:   "warning",
			Message: types.SARIFMessage{Text: fmt.Sprintf("%s: Variable is declared in .env but not used in code", varName)},
		})
	}

	// ENV002 - missing variables (with locations from data.Missing)
	missingVars := sortedKeys(data.Missing)
	for _, varName := range missingVars {
		activeRuleIDs["ENV002"] = true
		loc := data.Missing[varName]
		result := types.SARIFResult{
			RuleID:  "ENV002",
			Level:   "error",
			Message: types.SARIFMessage{Text: fmt.Sprintf("%s: Variable is used in code but not declared in .env", varName)},
		}
		if loc.FilePath != "" {
			result.Locations = []types.SARIFLocation{
				makeLocation(loc.FilePath, loc.LineNumber),
			}
		}
		results = append(results, result)
	}

	// ENV003 - code uses not in docker (with first location)
	codeVars := sortedKeysLocSlice(data.CodeUsesNotInDocker)
	for _, varName := range codeVars {
		activeRuleIDs["ENV003"] = true
		locs := data.CodeUsesNotInDocker[varName]
		result := types.SARIFResult{
			RuleID:  "ENV003",
			Level:   "warning",
			Message: types.SARIFMessage{Text: fmt.Sprintf("%s: Variable is used in code but not declared in Dockerfile or .env", varName)},
		}
		if len(locs) > 0 && locs[0].FilePath != "" {
			result.Locations = []types.SARIFLocation{
				makeLocation(locs[0].FilePath, locs[0].LineNumber),
			}
		}
		results = append(results, result)
	}

	// ENV004 - docker declares unused (no locations)
	dockerDeclaresUnused := data.DockerDeclaresUnused
	if dockerDeclaresUnused == nil {
		dockerDeclaresUnused = []string{}
	}
	sortedDockerDeclares := make([]string, len(dockerDeclaresUnused))
	copy(sortedDockerDeclares, dockerDeclaresUnused)
	sort.Strings(sortedDockerDeclares)
	for _, varName := range sortedDockerDeclares {
		activeRuleIDs["ENV004"] = true
		results = append(results, types.SARIFResult{
			RuleID:  "ENV004",
			Level:   "warning",
			Message: types.SARIFMessage{Text: fmt.Sprintf("%s: Variable is declared in Dockerfile but not used in code", varName)},
		})
	}

	// ENV005 - docker uses undeclared (with location)
	dockerVars := sortedKeysLoc(data.DockerUsesUndeclared)
	for _, varName := range dockerVars {
		activeRuleIDs["ENV005"] = true
		loc := data.DockerUsesUndeclared[varName]
		result := types.SARIFResult{
			RuleID:  "ENV005",
			Level:   "error",
			Message: types.SARIFMessage{Text: fmt.Sprintf("%s: Variable is used in Dockerfile but not declared", varName)},
		}
		if loc.FilePath != "" {
			result.Locations = []types.SARIFLocation{
				makeLocation(loc.FilePath, loc.LineNumber),
			}
		}
		results = append(results, result)
	}

	// Sort results by ruleId, then by variable name (extracted from message prefix)
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].RuleID != results[j].RuleID {
			return results[i].RuleID < results[j].RuleID
		}
		return results[i].Message.Text < results[j].Message.Text
	})

	// Ensure non-nil slices for valid JSON
	if results == nil {
		results = []types.SARIFResult{}
	}

	return results, activeRuleIDs
}

// makeLocation creates a SARIFLocation with forward-slash path separators.
func makeLocation(filePath string, lineNumber int) types.SARIFLocation {
	return types.SARIFLocation{
		PhysicalLocation: types.SARIFPhysicalLocation{
			ArtifactLocation: types.SARIFArtifactLocation{
				URI: filepath.ToSlash(filePath),
			},
			Region: types.SARIFRegion{
				StartLine: lineNumber,
			},
		},
	}
}

// sortedKeys returns sorted keys from a map[string]Location.
func sortedKeys(m map[string]types.Location) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedKeysLocSlice returns sorted keys from a map[string][]Location.
func sortedKeysLocSlice(m map[string][]types.Location) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedKeysLoc returns sorted keys from a map[string]Location (same as sortedKeys, aliased for clarity).
func sortedKeysLoc(m map[string]types.Location) []string {
	return sortedKeys(m)
}
