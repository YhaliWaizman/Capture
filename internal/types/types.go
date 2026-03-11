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
	// Additional data for JSON output
	AllLocations         map[string][]Location // All locations for each variable
	FilesScanned         int
	VariablesDeclared    int
	VariablesUsed        int
	CodeUsesNotInDocker  map[string][]Location
	DockerDeclaresUnused []string
	DockerUsesUndeclared map[string]Location
}

// JSONOutput represents the complete JSON output structure
type JSONOutput struct {
	Summary          Summary           `json:"summary"`
	Unused           []string          `json:"unused"`
	Missing          []MissingVariable `json:"missing"`
	DockerfileIssues DockerfileIssues  `json:"dockerfile_issues"`
}

// Summary contains scan statistics
type Summary struct {
	FilesScanned      int `json:"files_scanned"`
	VariablesDeclared int `json:"variables_declared"`
	VariablesUsed     int `json:"variables_used"`
	MismatchesFound   int `json:"mismatches_found"`
}

// MissingVariable represents a variable used but not declared
type MissingVariable struct {
	Variable  string     `json:"variable"`
	Locations []Location `json:"locations"`
}

// DockerfileIssues contains Dockerfile-specific mismatches
type DockerfileIssues struct {
	CodeUsesNotInDocker  []MissingVariable     `json:"code_uses_not_in_docker"`
	DockerDeclaresUnused []string              `json:"docker_declares_unused"`
	DockerUsesUndeclared []DockerUndeclaredVar `json:"docker_uses_undeclared"`
}

// DockerUndeclaredVar represents a variable used in Dockerfile but not declared
type DockerUndeclaredVar struct {
	Variable string   `json:"variable"`
	Location Location `json:"location"`
}
