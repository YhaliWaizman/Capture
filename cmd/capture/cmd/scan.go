package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	composeanalyzer "github.com/yhaliwaizman/capture/internal/compose"
	"github.com/yhaliwaizman/capture/internal/detector"
	"github.com/yhaliwaizman/capture/internal/diff"
	"github.com/yhaliwaizman/capture/internal/dockerfile"
	"github.com/yhaliwaizman/capture/internal/parser"
	"github.com/yhaliwaizman/capture/internal/reporter"
	"github.com/yhaliwaizman/capture/internal/types"
	"github.com/yhaliwaizman/capture/internal/walker"
	"gopkg.in/yaml.v3"
)

// ScanConfig holds the configuration for the scan command
type ScanConfig struct {
	Dir      string
	EnvFiles []string
	Ignore   []string
	Format   string
	Config   string
}

var scanConfig ScanConfig

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan project for environment variable mismatches",
	Long: `Scan analyzes your project to identify mismatches between environment variables 
declared in .env files, Dockerfiles, and those referenced in source code.

The tool will:
  - Parse .env file for declared variables
  - Analyze Dockerfiles for ENV/ARG declarations
  - Detect variable usage in source code (JS, TS, Go, Python, Ruby, PHP, Java, Kotlin)
  - Report mismatches and inconsistencies`,
	Example: `  capture scan --dir ./project --env-file .env
  capture scan --dir . --env-file .env --env-file .env.local --ignore vendor,tmp
  capture scan --dir . --env-file .env --format json`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runScan,
}

func init() {
	// Define flags
	scanCmd.Flags().StringVar(&scanConfig.Dir, "dir", ".", "Directory to scan (required)")
	scanCmd.Flags().StringSliceVar(&scanConfig.EnvFiles, "env-file", []string{".env"}, "Path to .env file (repeatable). Later files override earlier ones")
	scanCmd.Flags().StringSliceVar(&scanConfig.Ignore, "ignore", []string{}, "Comma-separated list of directories to ignore")
	scanCmd.Flags().StringVar(&scanConfig.Format, "format", "text", "Output format: text, json, or sarif")
	scanCmd.Flags().StringVar(&scanConfig.Config, "config", "", "Path to config file (.capture.yaml/.yml/.json)")
}

func runScan(cmd *cobra.Command, args []string) error {
	config, err := resolveScanConfig(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return NewExitError(err, 2)
	}

	// Validate format flag
	if config.Format != "text" && config.Format != "json" && config.Format != "sarif" {
		fmt.Fprintf(os.Stderr, "Error: invalid format '%s'. Must be 'text', 'json', or 'sarif'\n", config.Format)
		return NewExitError(fmt.Errorf("invalid format"), 2)
	}

	// Validate directory exists
	if info, err := os.Stat(config.Dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory does not exist: %s\n", config.Dir)
		return NewExitError(err, 2)
	} else if err != nil {
		// Handle permission errors
		fmt.Fprintf(os.Stderr, "Error: cannot access directory: %v\n", err)
		return NewExitError(err, 2)
	} else if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: path is not a directory: %s\n", config.Dir)
		return NewExitError(fmt.Errorf("not a directory"), 2)
	}

	// Trim whitespace from ignore directories
	for i := range config.Ignore {
		config.Ignore[i] = strings.TrimSpace(config.Ignore[i])
	}

	// Execute the scan
	exitCode := executeScan(&config)

	if exitCode != 0 {
		return NewExitError(nil, exitCode)
	}

	return nil
}

type scanFileConfig struct {
	Root     string   `yaml:"root" json:"root"`
	EnvFiles []string `yaml:"env_files" json:"env_files"`
	Ignore   []string `yaml:"ignore" json:"ignore"`
	Format   string   `yaml:"format" json:"format"`
}

func resolveScanConfig(cmd *cobra.Command) (ScanConfig, error) {
	config := ScanConfig{
		Dir:      ".",
		EnvFiles: []string{".env"},
		Ignore:   []string{},
		Format:   "text",
	}

	configPath, err := findConfigPath(cmd)
	if err != nil {
		return config, err
	}
	if configPath != "" {
		fileConfig, err := parseScanFileConfig(configPath)
		if err != nil {
			return config, err
		}
		config.Config = configPath
		if fileConfig.Root != "" {
			config.Dir = fileConfig.Root
		}
		if len(fileConfig.EnvFiles) > 0 {
			config.EnvFiles = fileConfig.EnvFiles
		}
		if len(fileConfig.Ignore) > 0 {
			config.Ignore = fileConfig.Ignore
		}
		if fileConfig.Format != "" {
			config.Format = fileConfig.Format
		}
	}

	if cmd.Flags().Changed("dir") {
		config.Dir = scanConfig.Dir
	}
	if cmd.Flags().Changed("env-file") {
		config.EnvFiles = scanConfig.EnvFiles
	}
	if cmd.Flags().Changed("ignore") {
		config.Ignore = scanConfig.Ignore
	}
	if cmd.Flags().Changed("format") {
		config.Format = scanConfig.Format
	}
	if cmd.Flags().Changed("config") {
		config.Config = scanConfig.Config
	}

	return config, nil
}

func findConfigPath(cmd *cobra.Command) (string, error) {
	if cmd.Flags().Changed("config") {
		if scanConfig.Config == "" {
			return "", fmt.Errorf("--config cannot be empty")
		}
		if _, err := os.Stat(scanConfig.Config); err != nil {
			return "", fmt.Errorf("failed to read config file %s: %w", scanConfig.Config, err)
		}
		return scanConfig.Config, nil
	}

	for _, candidate := range []string{".capture.yaml", ".capture.yml", ".capture.json"} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", nil
}

func parseScanFileConfig(configPath string) (scanFileConfig, error) {
	var fileConfig scanFileConfig

	file, err := os.Open(configPath)
	if err != nil {
		return fileConfig, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".yaml", ".yml":
		decoder := yaml.NewDecoder(file)
		decoder.KnownFields(true)
		if err := decoder.Decode(&fileConfig); err != nil {
			return fileConfig, fmt.Errorf("invalid YAML config %s: %w", configPath, err)
		}
	case ".json":
		decoder := json.NewDecoder(file)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&fileConfig); err != nil {
			return fileConfig, fmt.Errorf("invalid JSON config %s: %w", configPath, err)
		}
	default:
		return fileConfig, fmt.Errorf("unsupported config format for %s (use .yaml, .yml, or .json)", configPath)
	}

	return fileConfig, nil
}

// executeScan performs the actual scanning logic
// This preserves the original run() function logic
func executeScan(config *ScanConfig) int {
	// Initialize components
	envParser := parser.NewEnvParser()
	fileWalker := walker.NewFileWalker()
	detectorFactory := detector.NewDetectorFactory()
	diffEngine := diff.NewDiffEngine()
	rep := reporter.NewReporter(os.Stdout, os.Stderr)

	// Step 1: Parse .env files in order. Later files override earlier ones.
	declared := make(map[string]bool)
	declaredSources := make(map[string]string)
	parsedEnvFiles := 0

	for _, envFile := range config.EnvFiles {
		parsedDeclared, err := envParser.Parse(envFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse .env file %s: %v\n", envFile, err)
			continue
		}

		parsedEnvFiles++
		for varName := range parsedDeclared {
			declared[varName] = true
			declaredSources[varName] = envFile
		}
	}

	if parsedEnvFiles == 0 {
		fmt.Fprintln(os.Stderr, "Error: no readable env files found from --env-file values")
		return 2
	}

	// Step 2: Walk directory tree to find source files
	files, err := fileWalker.Walk(config.Dir, config.Ignore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to walk directory: %v\n", err)
		return 2
	}

	// Step 2.5: Separate Dockerfiles and Compose files from source files
	var dockerfiles []string
	var composeFiles []string
	var sourceFiles []string

	for _, filePath := range files {
		baseName := filepath.Base(filePath)
		isDockerfile := baseName == "Dockerfile" ||
			filepath.Ext(baseName) == ".dockerfile" ||
			strings.HasPrefix(baseName, "Dockerfile")
		isComposeFile := baseName == "docker-compose.yml" ||
			baseName == "docker-compose.yaml" ||
			baseName == "compose.yml" ||
			baseName == "compose.yaml" ||
			(strings.HasPrefix(baseName, "docker-compose.") &&
				(strings.HasSuffix(baseName, ".yml") || strings.HasSuffix(baseName, ".yaml")))

		if isDockerfile {
			dockerfiles = append(dockerfiles, filePath)
		} else if isComposeFile {
			composeFiles = append(composeFiles, filePath)
		} else {
			sourceFiles = append(sourceFiles, filePath)
		}
	}

	// Step 2.6: Analyze Dockerfiles
	dockerAnalyzer := dockerfile.NewDockerfileAnalyzer()
	dockerDeclared := make(map[string]bool)
	dockerUsed := make(map[string][]types.Location)

	for _, dockerfilePath := range dockerfiles {
		result, err := dockerAnalyzer.Analyze(dockerfilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to analyze %s: %v\n", dockerfilePath, err)
			continue
		}

		// Merge declarations
		for varName := range result.Declared {
			dockerDeclared[varName] = true
		}

		// Merge usage locations
		for varName, locs := range result.Used {
			dockerUsed[varName] = append(dockerUsed[varName], locs...)
		}
	}

	// Step 2.7: Analyze Docker Compose files
	composeAnalyzer := composeanalyzer.NewComposeAnalyzer()
	composeDeclared := make(map[string]types.Location)
	composeDeclaredSet := make(map[string]bool)
	composeUsed := make(map[string][]types.Location)
	composeMissingEnvFiles := make(map[string]types.Location)

	for _, composeFilePath := range composeFiles {
		result, err := composeAnalyzer.Analyze(composeFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to analyze %s: %v\n", composeFilePath, err)
			continue
		}

		for varName, loc := range result.Declared {
			composeDeclaredSet[varName] = true
			if _, exists := composeDeclared[varName]; !exists {
				composeDeclared[varName] = loc
			}
		}

		for varName, locs := range result.Used {
			composeUsed[varName] = append(composeUsed[varName], locs...)
		}

		for envFilePath, loc := range result.MissingEnvFiles {
			if _, exists := composeMissingEnvFiles[envFilePath]; !exists {
				composeMissingEnvFiles[envFilePath] = loc
			}
		}

		for envFilePath, loc := range result.EnvFiles {
			parsedDeclared, err := envParser.Parse(envFilePath)
			if err != nil {
				continue
			}
			for varName := range parsedDeclared {
				composeDeclaredSet[varName] = true
				if _, exists := composeDeclared[varName]; !exists {
					composeDeclared[varName] = loc
				}
			}
		}
	}

	// Step 3: Detect environment variable usage in source files
	used := make(map[string]bool)
	allLocations := make(map[string][]types.Location)

	for _, filePath := range sourceFiles {
		ext := filepath.Ext(filePath)
		detector := detectorFactory.Create(ext)
		if detector == nil {
			continue
		}

		locations, err := detector.Detect(filePath)
		if err != nil {
			// Soft error: log warning but continue processing
			fmt.Fprintf(os.Stderr, "Warning: failed to process file %s: %v\n", filePath, err)
			continue
		}

		// Merge locations into allLocations and mark variables as used
		for varName, locs := range locations {
			used[varName] = true
			allLocations[varName] = append(allLocations[varName], locs...)
		}
	}

	// Step 4: Compare declared vs used variables
	diffResult := diffEngine.Compare(declared, used)

	// Step 4.5: Docker cross-comparison
	var dockerMismatches bool

	// Check 1: Code uses variables not declared in Dockerfile or .env
	var codeUsedNotInDocker []string
	for varName := range used {
		if !dockerDeclared[varName] && !declared[varName] {
			codeUsedNotInDocker = append(codeUsedNotInDocker, varName)
		}
	}
	sort.Strings(codeUsedNotInDocker)
	if len(codeUsedNotInDocker) > 0 {
		dockerMismatches = true
	}

	// Check 2: Dockerfile declares variables unused in code
	var dockerDeclaredNotUsed []string
	for varName := range dockerDeclared {
		if !used[varName] {
			dockerDeclaredNotUsed = append(dockerDeclaredNotUsed, varName)
		}
	}
	sort.Strings(dockerDeclaredNotUsed)
	if len(dockerDeclaredNotUsed) > 0 {
		dockerMismatches = true
	}

	// Check 3: Dockerfile uses undeclared variables
	dockerUsedUndeclared := make(map[string]types.Location)
	var dockerUsedUndeclaredKeys []string
	for varName, locs := range dockerUsed {
		if !dockerDeclared[varName] && len(locs) > 0 {
			dockerUsedUndeclared[varName] = locs[0]
			dockerUsedUndeclaredKeys = append(dockerUsedUndeclaredKeys, varName)
		}
	}
	sort.Strings(dockerUsedUndeclaredKeys)
	if len(dockerUsedUndeclared) > 0 {
		dockerMismatches = true
	}

	// Step 4.6: Docker Compose cross-comparison
	var composeMismatches bool

	composeDeclaresNotInEnv := make(map[string]types.Location)
	if len(composeFiles) > 0 {
		for varName, loc := range composeDeclared {
			if !declared[varName] {
				composeDeclaresNotInEnv[varName] = loc
			}
		}
	}
	if len(composeDeclaresNotInEnv) > 0 {
		composeMismatches = true
	}

	composeUsesUndefined := make(map[string]types.Location)
	if len(composeFiles) > 0 {
		for varName, locs := range composeUsed {
			if composeDeclaredSet[varName] || dockerDeclared[varName] || declared[varName] || len(locs) == 0 {
				continue
			}
			composeUsesUndefined[varName] = locs[0]
		}
	}
	if len(composeUsesUndefined) > 0 {
		composeMismatches = true
	}

	var envDeclaresUnusedCompose []string
	if len(composeFiles) > 0 {
		for varName := range declared {
			if !composeDeclaredSet[varName] && len(composeUsed[varName]) == 0 {
				envDeclaresUnusedCompose = append(envDeclaresUnusedCompose, varName)
			}
		}
		sort.Strings(envDeclaresUnusedCompose)
	}
	if len(envDeclaresUnusedCompose) > 0 {
		composeMismatches = true
	}

	if len(composeFiles) > 0 && len(composeMissingEnvFiles) > 0 {
		composeMismatches = true
	}

	// Step 5: Prepare report data with first location for each missing variable
	reportData := types.ReportData{
		Unused:                   diffResult.Unused,
		Missing:                  make(map[string]types.Location),
		AllLocations:             allLocations,
		DeclaredSources:          declaredSources,
		FilesScanned:             len(sourceFiles) + len(dockerfiles) + len(composeFiles),
		VariablesDeclared:        len(declared),
		VariablesUsed:            len(used),
		CodeUsesNotInDocker:      make(map[string][]types.Location),
		DockerDeclaresUnused:     dockerDeclaredNotUsed,
		DockerUsesUndeclared:     dockerUsedUndeclared,
		ComposeDeclaresNotInEnv:  composeDeclaresNotInEnv,
		ComposeUsesUndefined:     composeUsesUndefined,
		EnvDeclaresUnusedCompose: envDeclaresUnusedCompose,
		ComposeMissingEnvFiles:   composeMissingEnvFiles,
	}

	for _, varName := range diffResult.Missing {
		if locs, ok := allLocations[varName]; ok && len(locs) > 0 {
			reportData.Missing[varName] = locs[0]
		}
	}

	// Add code uses not in docker with locations
	for _, varName := range codeUsedNotInDocker {
		if locs, ok := allLocations[varName]; ok {
			reportData.CodeUsesNotInDocker[varName] = locs
		}
	}

	// Step 6: Generate report based on format
	switch config.Format {
	case "json":
		if err := rep.ReportJSON(reportData); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to generate JSON report: %v\n", err)
			return 2
		}
	case "sarif":
		if err := rep.ReportSARIF(reportData); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to generate SARIF report: %v\n", err)
			return 2
		}
	default:
		// Text format (default)
		rep.Report(reportData)

		// Step 6.5: Report Docker-specific mismatches for text format
		if len(codeUsedNotInDocker) > 0 {
			fmt.Fprintln(os.Stdout, "\nCode uses variables not in Dockerfile or .env:")
			for _, varName := range codeUsedNotInDocker {
				if locs, ok := allLocations[varName]; ok && len(locs) > 0 {
					fmt.Fprintf(os.Stdout, "- %s (%s:%d)\n", varName, locs[0].FilePath, locs[0].LineNumber)
				}
			}
		}

		if len(dockerDeclaredNotUsed) > 0 {
			fmt.Fprintln(os.Stdout, "\nDockerfile declares but code doesn't use:")
			for _, varName := range dockerDeclaredNotUsed {
				fmt.Fprintf(os.Stdout, "- %s\n", varName)
			}
		}

		if len(dockerUsedUndeclared) > 0 {
			fmt.Fprintln(os.Stdout, "\nDockerfile uses undeclared variables:")
			for _, varName := range dockerUsedUndeclaredKeys {
				location := dockerUsedUndeclared[varName]
				fmt.Fprintf(os.Stdout, "- %s (%s:%d)\n", varName, location.FilePath, location.LineNumber)
			}
		}

		if len(composeFiles) > 0 && (len(composeDeclaresNotInEnv) > 0 || len(composeUsesUndefined) > 0 || len(envDeclaresUnusedCompose) > 0 || len(composeMissingEnvFiles) > 0) {
			fmt.Fprintln(os.Stdout, "\nDocker Compose issues:")
		}

		if len(composeFiles) > 0 && len(composeDeclaresNotInEnv) > 0 {
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "compose files declare but not in .env:")
			for _, varName := range sortedKeys(composeDeclaresNotInEnv) {
				location := composeDeclaresNotInEnv[varName]
				fmt.Fprintf(os.Stdout, "- %s (%s:%d)\n", varName, location.FilePath, location.LineNumber)
			}
		}

		if len(composeFiles) > 0 && len(composeUsesUndefined) > 0 {
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "compose files use undefined variables:")
			for _, varName := range sortedKeys(composeUsesUndefined) {
				location := composeUsesUndefined[varName]
				fmt.Fprintf(os.Stdout, "- %s (%s:%d)\n", varName, location.FilePath, location.LineNumber)
			}
		}

		if len(composeFiles) > 0 && len(envDeclaresUnusedCompose) > 0 {
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, ".env declares but not used in compose:")
			for _, varName := range envDeclaresUnusedCompose {
				fmt.Fprintf(os.Stdout, "- %s\n", varName)
			}
		}

		if len(composeFiles) > 0 && len(composeMissingEnvFiles) > 0 {
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "compose files reference missing env_file entries:")
			for _, envFilePath := range sortedKeys(composeMissingEnvFiles) {
				location := composeMissingEnvFiles[envFilePath]
				fmt.Fprintf(os.Stdout, "- %s (%s:%d)\n", envFilePath, location.FilePath, location.LineNumber)
			}
		}
	}

	// Determine exit code
	if len(diffResult.Unused) > 0 || len(diffResult.Missing) > 0 || dockerMismatches || composeMismatches {
		return 1 // Mismatches found
	}
	return 0 // No mismatches
}

func sortedKeys(m map[string]types.Location) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
