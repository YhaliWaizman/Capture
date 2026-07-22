package detector

import (
	"bufio"
	"os"
	"regexp"

	"github.com/yhaliwaizman/capture/internal/types"
)

// PHPDetector detects environment variable usage in PHP files.
type PHPDetector struct {
	patterns []*regexp.Regexp
}

// NewPHPDetector creates a new PHPDetector with compiled regex patterns.
func NewPHPDetector() *PHPDetector {
	return &PHPDetector{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\$_ENV\[['"]([A-Z][A-Z0-9_]*)['"]\]`),
			regexp.MustCompile(`\$_SERVER\[['"]([A-Z][A-Z0-9_]*)['"]\]`),
			regexp.MustCompile(`getenv\(['"]([A-Z][A-Z0-9_]*)['"]\)`),
		},
	}
}

// Detect scans a PHP file for environment variable usage.
func (d *PHPDetector) Detect(filePath string) (map[string][]types.Location, error) {
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
