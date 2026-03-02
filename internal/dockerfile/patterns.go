package dockerfile

// Regex patterns for Dockerfile parsing
//
// ENV Instruction Formats:
// - ENV KEY=value
// - ENV KEY value
// - ENV A=1 B=2 C=3
//
// Pattern: (?i)^ENV\s+(.+)
// - (?i) = case-insensitive
// - ^ENV = starts with ENV
// - \s+ = one or more whitespace
// - (.+) = capture rest of line
//
// ARG Instruction Formats:
// - ARG KEY
// - ARG KEY=default
//
// Pattern: (?i)^ARG\s+(.+)
//
// Variable Usage:
// - $VAR
// - ${VAR}
//
// Pattern: \$\{?([A-Z][A-Z0-9_]*)\}?
// - \$ = literal dollar sign
// - \{? = optional opening brace
// - ([A-Z][A-Z0-9_]*) = capture uppercase variable name
// - \}? = optional closing brace
//
// FROM Validation:
// Pattern: (?i)^FROM\s+
//
// Valid Variable Name:
// Pattern: ^[A-Z][A-Z0-9_]*$

const (
	// Instruction patterns (case-insensitive)
	envInstructionPattern  = `(?i)^ENV\s+(.+)`
	argInstructionPattern  = `(?i)^ARG\s+(.+)`
	fromInstructionPattern = `(?i)^FROM\s+`

	// Variable usage pattern (case-sensitive for variable names)
	varUsagePattern = `\$\{?([A-Z][A-Z0-9_]*)\}?`

	// Variable name validation (uppercase only)
	validNamePattern = `^[A-Z][A-Z0-9_]*$`
)
