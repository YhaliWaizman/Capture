package compose

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yhaliwaizman/capture/internal/types"
	"gopkg.in/yaml.v3"
)

// ComposeAnalyzer extracts environment declarations and variable usage from Docker Compose files.
type ComposeAnalyzer struct {
	validNamePattern *regexp.Regexp
	varUsagePattern  *regexp.Regexp
}

// AnalysisResult holds Compose declarations, variable usage, and env_file references.
type AnalysisResult struct {
	Declared        map[string]types.Location
	Used            map[string][]types.Location
	EnvFiles        map[string]types.Location
	MissingEnvFiles map[string]types.Location
}

func NewComposeAnalyzer() *ComposeAnalyzer {
	return &ComposeAnalyzer{
		validNamePattern: regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`),
		varUsagePattern:  regexp.MustCompile(`\$\{([A-Z][A-Z0-9_]*)(?::[-+?][^}]*)?\}`),
	}
}

func (a *ComposeAnalyzer) Analyze(filePath string) (*AnalysisResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, err
	}

	result := &AnalysisResult{
		Declared:        make(map[string]types.Location),
		Used:            make(map[string][]types.Location),
		EnvFiles:        make(map[string]types.Location),
		MissingEnvFiles: make(map[string]types.Location),
	}

	if len(root.Content) == 0 {
		return result, nil
	}

	doc := root.Content[0]
	a.extractSubstitutions(doc, filePath, result.Used)
	a.extractServiceConfig(doc, filePath, result)

	return result, nil
}

func (a *ComposeAnalyzer) extractSubstitutions(node *yaml.Node, filePath string, used map[string][]types.Location) {
	if node == nil {
		return
	}

	if node.Kind == yaml.ScalarNode {
		matches := a.varUsagePattern.FindAllStringSubmatch(node.Value, -1)
		seen := make(map[string]bool)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			varName := match[1]
			if seen[varName] {
				continue
			}
			seen[varName] = true
			used[varName] = append(used[varName], types.Location{
				FilePath:   filePath,
				LineNumber: node.Line,
			})
		}
	}

	for _, child := range node.Content {
		a.extractSubstitutions(child, filePath, used)
	}
}

func (a *ComposeAnalyzer) extractServiceConfig(doc *yaml.Node, composePath string, result *AnalysisResult) {
	services := mapValue(doc, "services")
	if services == nil || services.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(services.Content); i += 2 {
		serviceConfig := services.Content[i+1]

		environment := mapValue(serviceConfig, "environment")
		if environment != nil {
			a.extractEnvironment(environment, composePath, result.Declared)
		}

		envFile := mapValue(serviceConfig, "env_file")
		if envFile != nil {
			a.extractEnvFiles(envFile, composePath, result.EnvFiles, result.MissingEnvFiles)
		}
	}
}

func (a *ComposeAnalyzer) extractEnvironment(node *yaml.Node, composePath string, declared map[string]types.Location) {
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			key := strings.TrimSpace(node.Content[i].Value)
			if !a.validNamePattern.MatchString(key) {
				continue
			}
			if _, exists := declared[key]; !exists {
				declared[key] = types.Location{
					FilePath:   composePath,
					LineNumber: node.Content[i].Line,
				}
			}
		}
	case yaml.SequenceNode:
		for _, item := range node.Content {
			entry := strings.TrimSpace(item.Value)
			if entry == "" {
				continue
			}

			key := entry
			if idx := strings.Index(entry, "="); idx > 0 {
				key = strings.TrimSpace(entry[:idx])
			}

			if !a.validNamePattern.MatchString(key) {
				continue
			}
			if _, exists := declared[key]; !exists {
				declared[key] = types.Location{
					FilePath:   composePath,
					LineNumber: item.Line,
				}
			}
		}
	}
}

func (a *ComposeAnalyzer) extractEnvFiles(
	node *yaml.Node,
	composePath string,
	envFiles map[string]types.Location,
	missing map[string]types.Location,
) {
	registerPath := func(rawPath string, line int) {
		value := strings.TrimSpace(rawPath)
		if value == "" {
			return
		}

		resolved := value
		if !filepath.IsAbs(value) {
			resolved = filepath.Join(filepath.Dir(composePath), value)
		}
		resolved = filepath.Clean(resolved)

		location := types.Location{FilePath: composePath, LineNumber: line}
		if _, exists := envFiles[resolved]; !exists {
			envFiles[resolved] = location
		}

		if _, err := os.Stat(resolved); err != nil {
			if _, exists := missing[resolved]; !exists {
				missing[resolved] = location
			}
		}
	}

	switch node.Kind {
	case yaml.ScalarNode:
		registerPath(node.Value, node.Line)
	case yaml.SequenceNode:
		for _, item := range node.Content {
			if item.Kind != yaml.ScalarNode {
				continue
			}
			registerPath(item.Value, item.Line)
		}
	}
}

func mapValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func SortedKeysLocation(m map[string]types.Location) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
