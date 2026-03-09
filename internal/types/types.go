package types

// Location represents a file path and line number where a variable is used
type Location struct {
	FilePath   string // Relative to scan directory
	LineNumber int    // 1-indexed line number
}

// DiffResult contains the results of comparing declared and used variable sets
type DiffResult struct {
	Unused  []string // Variables declared but not used
	Missing []string // Variables used but not declared
}

// ReportData contains the data needed to generate a report
type ReportData struct {
	Unused  []string            // Variables declared but not used
	Missing map[string]Location // Variables used but not declared, mapped to first location
}
