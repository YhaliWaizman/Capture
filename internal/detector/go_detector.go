package detector

import (
	"bufio"
	"os"
	"regexp"
	"sort"

	"github.com/yhaliwaizman/capture/internal/types"
)

// GoDetector detects environment variable usage in Go files
type GoDetector struct {
	patterns []*regexp.Regexp
}

// NewGoDetector creates a new GoDetector with compiled regex patterns
func NewGoDetector() *GoDetector {
	return &GoDetector{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`os\.Getenv\("([A-Z][A-Z0-9_]*)"\)`),
			regexp.MustCompile(`os\.LookupEnv\("([A-Z][A-Z0-9_]*)"\)`),
		},
	}
}

// Detect scans a Go file for environment variable usage
func (d *GoDetector) Detect(filePath string) (map[string][]types.Location, error) {
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
		sortGoLocations(result[varName])
	}

	return result, nil
}

// sortGoLocations sorts a slice of locations by file path then line number
func sortGoLocations(locations []types.Location) {
	sort.Slice(locations, func(i, j int) bool {
		if locations[i].FilePath != locations[j].FilePath {
			return locations[i].FilePath < locations[j].FilePath
		}
		return locations[i].LineNumber < locations[j].LineNumber
	})
}
