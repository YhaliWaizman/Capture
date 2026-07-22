package detector

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/yhaliwaizman/capture/internal/types"
)

var (
	upperIdentifierPattern = regexp.MustCompile(`^[A-Z0-9_]+$`)
	nonAlnumPattern        = regexp.MustCompile(`[^A-Za-z0-9]+`)
	sensitiveVarNameRegex  = regexp.MustCompile(`(?i)(password|passwd|secret|token|api[_-]?key|private[_-]?key|client[_-]?secret|access[_-]?key)`)
)

type secretPattern struct {
	name    string
	pattern *regexp.Regexp
}

// HardcodedSecretDetector detects potential hardcoded secrets in source files.
type HardcodedSecretDetector struct {
	patterns          []secretPattern
	assignmentPattern *regexp.Regexp
}

// NewHardcodedSecretDetector creates a new HardcodedSecretDetector.
func NewHardcodedSecretDetector() *HardcodedSecretDetector {
	return &HardcodedSecretDetector{
		patterns: []secretPattern{
			{name: "AWS Access Key", pattern: regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
			{name: "GitHub Token", pattern: regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{36,}`)},
			{name: "Stripe Key", pattern: regexp.MustCompile(`(?:sk|pk)_(?:live|test)_[A-Za-z0-9]{16,}`)},
			{name: "Google API Key", pattern: regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`)},
			{name: "JWT Token", pattern: regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`)},
			{name: "Private Key", pattern: regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH |)?PRIVATE KEY-----`)},
		},
		assignmentPattern: regexp.MustCompile(`(?i)(?:\b(?:const|let|var|final|private|public|protected)\s+)?["']?([a-z_][a-z0-9_-]*)["']?\s*(?::=|=|:)\s*["']([^"']{8,})["']`),
	}
}

// Detect scans a file and returns possible hardcoded secret findings.
func (d *HardcodedSecretDetector) Detect(filePath string) ([]types.HardcodedSecret, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var findings []types.HardcodedSecret
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		for _, p := range d.patterns {
			if p.pattern.FindStringIndex(line) == nil {
				continue
			}
			key := fmt.Sprintf("%d:%s", lineNumber, p.name)
			if seen[key] {
				continue
			}
			seen[key] = true
			findings = append(findings, types.HardcodedSecret{
				Type: p.name,
				Location: types.Location{
					FilePath:   filePath,
					LineNumber: lineNumber,
				},
				Suggestion: "Move the value to an environment variable and load it at runtime.",
			})
		}

		matches := d.assignmentPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}
			varName := match[1]
			if !sensitiveVarNameRegex.MatchString(varName) {
				continue
			}
			value := match[2]
			if isPlaceholderSecretValue(value) {
				continue
			}
			envVar := toEnvVarName(varName)
			key := fmt.Sprintf("%d:assignment:%s", lineNumber, envVar)
			if seen[key] {
				continue
			}
			seen[key] = true

			findings = append(findings, types.HardcodedSecret{
				Type: "Hardcoded credential",
				Location: types.Location{
					FilePath:   filePath,
					LineNumber: lineNumber,
				},
				Suggestion: fmt.Sprintf("Replace with an environment variable (for example: %s).", envVar),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Location.FilePath != findings[j].Location.FilePath {
			return findings[i].Location.FilePath < findings[j].Location.FilePath
		}
		if findings[i].Location.LineNumber != findings[j].Location.LineNumber {
			return findings[i].Location.LineNumber < findings[j].Location.LineNumber
		}
		return findings[i].Type < findings[j].Type
	})

	return findings, nil
}

func isPlaceholderSecretValue(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "password", "secret", "changeme", "example", "test", "dummy", "your_password", "your_secret", "your_api_key":
		return true
	}
	if strings.Contains(v, "${") || strings.Contains(v, "process.env") || strings.Contains(v, "os.getenv") {
		return true
	}
	if upperIdentifierPattern.MatchString(value) {
		return true
	}
	return false
}

func toEnvVarName(varName string) string {
	normalized := nonAlnumPattern.ReplaceAllString(varName, "_")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		return "SECRET_VALUE"
	}
	return strings.ToUpper(normalized)
}
