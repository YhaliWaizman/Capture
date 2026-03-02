package reporter

import (
	"fmt"
	"io"
	"sort"

	"github.com/yhaliwaizman/capture/internal/types"
)

// Reporter defines the interface for formatting and outputting analysis results
type Reporter interface {
	Report(data types.ReportData)
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
