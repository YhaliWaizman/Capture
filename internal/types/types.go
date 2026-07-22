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
	AllLocations             map[string][]Location // All locations for each variable
	DeclaredSources          map[string]string     // Variable -> declaring env file (last file wins)
	FilesScanned             int
	VariablesDeclared        int
	VariablesUsed            int
	HardcodedSecrets         []HardcodedSecret
	CodeUsesNotInDocker      map[string][]Location
	DockerDeclaresUnused     []string
	DockerUsesUndeclared     map[string]Location
	ComposeDeclaresNotInEnv  map[string]Location
	ComposeUsesUndefined     map[string]Location
	EnvDeclaresUnusedCompose []string
	ComposeMissingEnvFiles   map[string]Location
}

// JSONOutput represents the complete JSON output structure
type JSONOutput struct {
	Summary          Summary           `json:"summary"`
	Unused           []string          `json:"unused"`
	Missing          []MissingVariable `json:"missing"`
	DeclaredSources  map[string]string `json:"declared_sources"`
	HardcodedSecrets []HardcodedSecret `json:"hardcoded_secrets"`
	DockerfileIssues DockerfileIssues  `json:"dockerfile_issues"`
	ComposeIssues    ComposeIssues     `json:"compose_issues"`
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

// HardcodedSecret represents a possible hardcoded secret in source code.
type HardcodedSecret struct {
	Type       string   `json:"type"`
	Location   Location `json:"location"`
	Suggestion string   `json:"suggestion"`
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

// ComposeIssues contains Docker Compose-specific mismatches
type ComposeIssues struct {
	ComposeDeclaresNotInEnv  []ComposeVariableIssue `json:"compose_declares_not_in_env"`
	ComposeUsesUndefined     []ComposeVariableIssue `json:"compose_uses_undefined"`
	EnvDeclaresUnusedCompose []string               `json:"env_declares_unused_in_compose"`
	ComposeMissingEnvFiles   []ComposeEnvFileIssue  `json:"compose_missing_env_files"`
}

// ComposeVariableIssue represents a variable issue found in a compose file.
type ComposeVariableIssue struct {
	Variable string   `json:"variable"`
	Location Location `json:"location"`
}

// ComposeEnvFileIssue represents a missing env_file reference in a compose file.
type ComposeEnvFileIssue struct {
	Path     string   `json:"path"`
	Location Location `json:"location"`
}
