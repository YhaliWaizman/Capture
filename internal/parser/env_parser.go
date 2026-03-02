package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// EnvParser defines the interface for parsing .env files
type EnvParser interface {
	Parse(filePath string) (map[string]bool, error)
}

// EnvParserImpl implements the EnvParser interface
type EnvParserImpl struct {
	keyPattern   *regexp.Regexp
	validPattern *regexp.Regexp
}

// NewEnvParser creates a new EnvParser instance
func NewEnvParser() *EnvParserImpl {
	return &EnvParserImpl{
		keyPattern:   regexp.MustCompile(`^([A-Z][A-Z0-9_]*)\s*=`),
		validPattern: regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`),
	}
}

// Parse reads a .env file and extracts declared variable names
func (p *EnvParserImpl) Parse(filePath string) (map[string]bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	declared := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Extract KEY from KEY=VALUE pattern (use trimmed line)
		matches := p.keyPattern.FindStringSubmatch(trimmed)
		if len(matches) > 1 {
			key := strings.TrimSpace(matches[1])

			// Filter keys matching ^[A-Z][A-Z0-9_]*$
			if p.validPattern.MatchString(key) {
				declared[key] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return declared, nil
}
