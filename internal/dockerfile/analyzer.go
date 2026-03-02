package dockerfile

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/yhaliwaizman/capture/internal/types"
)

// DockerfileAnalyzer analyzes Dockerfile for environment variable declarations and usage
type DockerfileAnalyzer struct {
	// Regex patterns compiled once for performance
	envPattern       *regexp.Regexp // Matches ENV instructions
	argPattern       *regexp.Regexp // Matches ARG instructions
	varUsagePattern  *regexp.Regexp // Matches $VAR and ${VAR}
	validNamePattern *regexp.Regexp // Validates variable names
	fromPattern      *regexp.Regexp // Validates FROM instruction exists
}

// NewDockerfileAnalyzer creates a new analyzer with compiled regex patterns
func NewDockerfileAnalyzer() *DockerfileAnalyzer {
	return &DockerfileAnalyzer{
		envPattern:       regexp.MustCompile(envInstructionPattern),
		argPattern:       regexp.MustCompile(argInstructionPattern),
		varUsagePattern:  regexp.MustCompile(varUsagePattern),
		validNamePattern: regexp.MustCompile(validNamePattern),
		fromPattern:      regexp.MustCompile(fromInstructionPattern),
	}
}

// AnalysisResult holds both declared and used variables from Dockerfile
type AnalysisResult struct {
	Declared map[string]bool             // Variables declared via ENV/ARG
	Used     map[string][]types.Location // Variables used in instructions
}

// Analyze processes a Dockerfile and extracts declarations and usage
func (a *DockerfileAnalyzer) Analyze(filePath string) (*AnalysisResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Step 1: Preprocess line continuations
	lines := preprocessLines(scanner)
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Step 2: Validate FROM instruction exists
	if !a.validateDockerfile(lines) {
		// Not a valid Dockerfile, skip silently
		return &AnalysisResult{
			Declared: make(map[string]bool),
			Used:     make(map[string][]types.Location),
		}, nil
	}

	result := &AnalysisResult{
		Declared: make(map[string]bool),
		Used:     make(map[string][]types.Location),
	}

	// Step 3 & 4: Extract declarations and usage
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract ENV declarations
		if a.envPattern.MatchString(line) {
			varNames := a.extractENVDeclarations(line)
			for _, varName := range varNames {
				result.Declared[varName] = true
			}
		}

		// Extract ARG declarations
		if a.argPattern.MatchString(line) {
			varNames := a.extractARGDeclarations(line)
			for _, varName := range varNames {
				result.Declared[varName] = true
			}
		}

		// Extract variable usage from all instructions
		usageMap := a.extractVariableUsage(line, lineNum+1, filePath)
		for varName, location := range usageMap {
			result.Used[varName] = append(result.Used[varName], location)
		}
	}

	return result, nil
}

// preprocessLines handles line continuations with backslash
func preprocessLines(scanner *bufio.Scanner) []string {
	var lines []string
	var currentLine strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Check if line ends with backslash (continuation)
		if strings.HasSuffix(strings.TrimRight(line, " \t"), "\\") {
			// Remove trailing backslash and whitespace
			line = strings.TrimRight(line, " \t")
			line = strings.TrimSuffix(line, "\\")
			currentLine.WriteString(line)
			currentLine.WriteString(" ") // Add space between continued lines
		} else {
			// Line doesn't continue
			currentLine.WriteString(line)
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}
	}

	// Handle case where file ends with continuation
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// extractENVDeclarations parses ENV instruction and extracts variable names
func (a *DockerfileAnalyzer) extractENVDeclarations(line string) []string {
	var varNames []string

	matches := a.envPattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return varNames
	}

	// Get the part after "ENV "
	rest := strings.TrimSpace(matches[1])

	// Handle two formats:
	// 1. ENV KEY=value KEY2=value2 (space-separated key=value pairs)
	// 2. ENV KEY value (single key-value with space)

	// Check if it contains '=' - if so, it's format 1
	if strings.Contains(rest, "=") {
		// Split by spaces and extract keys from key=value pairs
		parts := strings.Fields(rest)
		for _, part := range parts {
			if idx := strings.Index(part, "="); idx > 0 {
				key := part[:idx]
				key = strings.TrimSpace(key)
				if a.validNamePattern.MatchString(key) {
					varNames = append(varNames, key)
				}
			}
		}
	} else {
		// Format 2: ENV KEY value
		parts := strings.Fields(rest)
		if len(parts) > 0 {
			key := parts[0]
			if a.validNamePattern.MatchString(key) {
				varNames = append(varNames, key)
			}
		}
	}

	return varNames
}

// extractARGDeclarations parses ARG instruction and extracts variable names
func (a *DockerfileAnalyzer) extractARGDeclarations(line string) []string {
	var varNames []string

	matches := a.argPattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return varNames
	}

	// Get the part after "ARG "
	rest := strings.TrimSpace(matches[1])

	// Handle two formats:
	// 1. ARG KEY
	// 2. ARG KEY=default

	// Extract key (everything before '=' or the whole string)
	key := rest
	if idx := strings.Index(rest, "="); idx > 0 {
		key = rest[:idx]
	}

	key = strings.TrimSpace(key)
	if a.validNamePattern.MatchString(key) {
		varNames = append(varNames, key)
	}

	return varNames
}

// extractVariableUsage finds $VAR and ${VAR} in instruction lines
func (a *DockerfileAnalyzer) extractVariableUsage(line string, lineNumber int, filePath string) map[string]types.Location {
	usageMap := make(map[string]types.Location)

	// Find all variable references
	matches := a.varUsagePattern.FindAllStringSubmatch(line, -1)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]

			// Skip if we've already recorded this variable on this line
			if _, exists := usageMap[varName]; exists {
				continue
			}

			// Only record valid uppercase variable names
			if a.validNamePattern.MatchString(varName) {
				usageMap[varName] = types.Location{
					FilePath:   filePath,
					LineNumber: lineNumber,
				}
			}
		}
	}

	return usageMap
}

// validateDockerfile checks if file has at least one FROM instruction
func (a *DockerfileAnalyzer) validateDockerfile(lines []string) bool {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Check if line starts with FROM (case-insensitive)
		if a.fromPattern.MatchString(line) {
			return true
		}
	}
	return false
}
