package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/yhaliwaizman/capture/internal/types"
)

// Reporter defines the interface for formatting and outputting analysis results
type Reporter interface {
	Report(data types.ReportData)
	ReportJSON(data types.ReportData) error
}

// ReporterImpl implements the Reporter interface
type ReporterImpl struct {
	out io.Writer
	err io.Writer
}

// NewReporter creates a new Reporter instance
// out is used for analysis results (stdout)
// err is used for warnings and errors (stderr)
func NewReporter(out, err io.Writer) *ReporterImpl {
	return &ReporterImpl{
		out: out,
		err: err,
	}
}

// Report formats and outputs the analysis results
// Implements requirements 10.1-10.9
func (r *ReporterImpl) Report(data types.ReportData) {
	// Requirement 10.6: Output "No environment mismatches found." when no mismatches
	if len(data.Unused) == 0 && len(data.Missing) == 0 {
		fmt.Fprintln(r.out, "No environment mismatches found.")
		return
	}

	// Requirement 10.1, 10.2: Format "Declared but unused:" section
	if len(data.Unused) > 0 {
		fmt.Fprintln(r.out, "Declared but unused:")
		for _, varName := range data.Unused {
			fmt.Fprintf(r.out, "- %s\n", varName)
		}
	}

	// Requirement 10.5: Add blank line separator between sections when both exist
	if len(data.Unused) > 0 && len(data.Missing) > 0 {
		fmt.Fprintln(r.out, "")
	}

	// Requirement 10.3, 10.4: Format "Used but not declared:" section
	if len(data.Missing) > 0 {
		fmt.Fprintln(r.out, "Used but not declared:")

		// Requirement 10.7: Sort missing variables alphabetically for determinism
		vars := make([]string, 0, len(data.Missing))
		for v := range data.Missing {
			vars = append(vars, v)
		}
		sort.Strings(vars)

		// Requirement 10.8: Use first location for each missing variable
		for _, varName := range vars {
			location := data.Missing[varName]
			fmt.Fprintf(r.out, "- %s (%s:%d)\n", varName, location.FilePath, location.LineNumber)
		}
	}
}

// ReportJSON formats and outputs the analysis results as JSON
func (r *ReporterImpl) ReportJSON(data types.ReportData) error {
	// Normalize nil slices to empty slices for consistent JSON output
	unused := data.Unused
	if unused == nil {
		unused = []string{}
	}

	dockerDeclaresUnused := data.DockerDeclaresUnused
	if dockerDeclaresUnused == nil {
		dockerDeclaresUnused = []string{}
	}

	// Build missing variables with all locations
	missing := make([]types.MissingVariable, 0, len(data.Missing))
	missingVars := make([]string, 0, len(data.Missing))
	for v := range data.Missing {
		missingVars = append(missingVars, v)
	}
	sort.Strings(missingVars)

	for _, varName := range missingVars {
		locations := data.AllLocations[varName]
		// Ensure locations is never nil - use empty slice if not found
		if locations == nil {
			// Fallback: use the single location from Missing map if AllLocations is empty
			if loc, ok := data.Missing[varName]; ok {
				locations = []types.Location{loc}
			} else {
				locations = []types.Location{}
			}
		}
		missing = append(missing, types.MissingVariable{
			Variable:  varName,
			Locations: locations,
		})
	}

	// Build code uses not in docker
	codeUsesNotInDocker := make([]types.MissingVariable, 0, len(data.CodeUsesNotInDocker))
	codeVars := make([]string, 0, len(data.CodeUsesNotInDocker))
	for v := range data.CodeUsesNotInDocker {
		codeVars = append(codeVars, v)
	}
	sort.Strings(codeVars)

	for _, varName := range codeVars {
		locations := data.CodeUsesNotInDocker[varName]
		// Ensure locations is never nil
		if locations == nil {
			locations = []types.Location{}
		}
		codeUsesNotInDocker = append(codeUsesNotInDocker, types.MissingVariable{
			Variable:  varName,
			Locations: locations,
		})
	}

	// Build docker uses undeclared
	dockerUsesUndeclared := make([]types.DockerUndeclaredVar, 0, len(data.DockerUsesUndeclared))
	dockerVars := make([]string, 0, len(data.DockerUsesUndeclared))
	for v := range data.DockerUsesUndeclared {
		dockerVars = append(dockerVars, v)
	}
	sort.Strings(dockerVars)

	for _, varName := range dockerVars {
		location := data.DockerUsesUndeclared[varName]
		dockerUsesUndeclared = append(dockerUsesUndeclared, types.DockerUndeclaredVar{
			Variable: varName,
			Location: location,
		})
	}

	// Calculate total mismatches
	mismatchesFound := len(unused) + len(missing) +
		len(codeUsesNotInDocker) + len(dockerDeclaresUnused) +
		len(dockerUsesUndeclared)

	// Build JSON output
	output := types.JSONOutput{
		Summary: types.Summary{
			FilesScanned:      data.FilesScanned,
			VariablesDeclared: data.VariablesDeclared,
			VariablesUsed:     data.VariablesUsed,
			MismatchesFound:   mismatchesFound,
		},
		Unused:  unused,
		Missing: missing,
		DockerfileIssues: types.DockerfileIssues{
			CodeUsesNotInDocker:  codeUsesNotInDocker,
			DockerDeclaresUnused: dockerDeclaresUnused,
			DockerUsesUndeclared: dockerUsesUndeclared,
		},
	}

	// Marshal to JSON with indentation
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to output
	fmt.Fprintln(r.out, string(jsonBytes))
	return nil
}
