package types

// EnvParser reads .env files and extracts declared variable names
type EnvParser interface {
	Parse(filePath string) (map[string]bool, error)
}

// FileWalker recursively traverses directories to find source files
type FileWalker interface {
	Walk(rootDir string, ignoreDirs []string) ([]string, error)
}

// LanguageDetector detects environment variable usage in source files
type LanguageDetector interface {
	Detect(filePath string) (map[string][]Location, error)
}

// DiffEngine compares declared and used variable sets
type DiffEngine interface {
	Compare(declared, used map[string]bool) DiffResult
}

// Reporter formats and outputs analysis results
type Reporter interface {
	Report(data ReportData)
}
