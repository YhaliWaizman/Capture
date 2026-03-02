package detector

import (
	"bufio"
	"os"
	"regexp"
	"sort"

	"github.com/capture-env-analyzer/internal/types"
)

// JSDetector detects environment variable usage in JavaScript and TypeScript files
type JSDetector struct {
	patterns []*regexp.Regexp
}

// NewJSDetector creates a new JSDetector with compiled regex patterns
func NewJSDetector() *JSDetector {
	return &JSDetector{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`process\.env\.([A-Z][A-Z0-9_]*)`),
			regexp.MustCompile(`process\.env\["([A-Z][A-Z0-9_]*)"\]`),
			regexp.MustCompile(`process\.env\['([A-Z][A-Z0-9_]*)'\]`),
		},
	}
}

// Detect scans a JavaScript/TypeScript file for environment variable usage
func (d *JSDetector) Detect(filePath string) (map[string][]types.Location, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string][]types.Location)
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Apply each pattern to the line
		for _, pattern := range d.patterns {
			matches := pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					varName := match[1]
					location := types.Location{
						FilePath:   filePath,
						LineNumber: lineNumber,
					}
					result[varName] = append(result[varName], location)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Sort locations for deterministic output
	for varName := range result {
		sortLocations(result[varName])
	}

	return result, nil
}

// sortLocations sorts a slice of locations by file path then line number
func sortLocations(locations []types.Location) {
	sort.Slice(locations, func(i, j int) bool {
		if locations[i].FilePath != locations[j].FilePath {
			return locations[i].FilePath < locations[j].FilePath
		}
		return locations[i].LineNumber < locations[j].LineNumber
	})
}
