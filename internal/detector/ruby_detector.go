package detector

import (
	"bufio"
	"os"
	"regexp"

	"github.com/yhaliwaizman/capture/internal/types"
)

// RubyDetector detects environment variable usage in Ruby files.
type RubyDetector struct {
	patterns []*regexp.Regexp
}

// NewRubyDetector creates a new RubyDetector with compiled regex patterns.
func NewRubyDetector() *RubyDetector {
	return &RubyDetector{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`ENV\["([A-Z][A-Z0-9_]*)"\]`),
			regexp.MustCompile(`ENV\['([A-Z][A-Z0-9_]*)'\]`),
			regexp.MustCompile(`ENV\.fetch\("([A-Z][A-Z0-9_]*)"`),
			regexp.MustCompile(`ENV\.fetch\('([A-Z][A-Z0-9_]*)'`),
		},
	}
}

// Detect scans a Ruby file for environment variable usage.
func (d *RubyDetector) Detect(filePath string) (map[string][]types.Location, error) {
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

	for varName := range result {
		sortLocations(result[varName])
	}

	return result, nil
}
