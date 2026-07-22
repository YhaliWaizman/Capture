package cmd

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

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
	Dir         string
	EnvFiles    []string
	Ignore      []string
	Format      string
	Config      string
	Workers     int
	Incremental bool
	NoCache     bool
	Watch       bool
	Fix         bool
	DryRun      bool
	Yes         bool
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
  capture scan --dir . --env-file .env --format json
  capture scan --dir . --env-file .env --fix --yes`,
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
	scanCmd.Flags().IntVar(&scanConfig.Workers, "workers", runtime.NumCPU(), "Number of parallel workers for source file scanning")
	scanCmd.Flags().BoolVar(&scanConfig.Incremental, "incremental", false, "Only scan files changed since the last scan (uses .capture/cache.json)")
	scanCmd.Flags().BoolVar(&scanConfig.NoCache, "no-cache", false, "Disable scan cache and force a full scan")
	scanCmd.Flags().BoolVar(&scanConfig.Watch, "watch", false, "Watch files and re-run scans on changes")
	scanCmd.Flags().BoolVar(&scanConfig.Fix, "fix", false, "Add missing variables to the first --env-file with empty values")
	scanCmd.Flags().BoolVar(&scanConfig.DryRun, "dry-run", false, "Preview changes from --fix without writing files")
	scanCmd.Flags().BoolVar(&scanConfig.Yes, "yes", false, "Skip confirmation prompt for --fix")
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
	if config.Workers < 1 {
		fmt.Fprintf(os.Stderr, "Error: invalid workers value '%d'. Must be >= 1\n", config.Workers)
		return NewExitError(fmt.Errorf("invalid workers"), 2)
	}
	if config.DryRun && !config.Fix {
		fmt.Fprintln(os.Stderr, "Error: --dry-run requires --fix")
		return NewExitError(fmt.Errorf("invalid dry-run usage"), 2)
	}
	if config.Yes && !config.Fix {
		fmt.Fprintln(os.Stderr, "Error: --yes requires --fix")
		return NewExitError(fmt.Errorf("invalid yes usage"), 2)
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

	if config.Watch {
		exitCode := runWatchScan(&config)
		if exitCode != 0 {
			return NewExitError(nil, exitCode)
		}
		return nil
	}

	exitCode := executeScan(&config)
	if exitCode != 0 {
		return NewExitError(nil, exitCode)
	}

	return nil
}

type scanFileConfig struct {
	Root        string   `yaml:"root" json:"root"`
	EnvFiles    []string `yaml:"env_files" json:"env_files"`
	Ignore      []string `yaml:"ignore" json:"ignore"`
	Format      string   `yaml:"format" json:"format"`
	Workers     int      `yaml:"workers" json:"workers"`
	Incremental bool     `yaml:"incremental" json:"incremental"`
	NoCache     bool     `yaml:"no_cache" json:"no_cache"`
	Watch       bool     `yaml:"watch" json:"watch"`
	Fix         bool     `yaml:"fix" json:"fix"`
	DryRun      bool     `yaml:"dry_run" json:"dry_run"`
	Yes         bool     `yaml:"yes" json:"yes"`
}

func resolveScanConfig(cmd *cobra.Command) (ScanConfig, error) {
	config := ScanConfig{
		Dir:         ".",
		EnvFiles:    []string{".env"},
		Ignore:      []string{},
		Format:      "text",
		Workers:     runtime.NumCPU(),
		Incremental: false,
		NoCache:     false,
		Watch:       false,
		Fix:         false,
		DryRun:      false,
		Yes:         false,
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
		if fileConfig.Workers > 0 {
			config.Workers = fileConfig.Workers
		}
		if fileConfig.Incremental {
			config.Incremental = true
		}
		if fileConfig.NoCache {
			config.NoCache = true
		}
		if fileConfig.Watch {
			config.Watch = true
		}
		if fileConfig.Fix {
			config.Fix = true
		}
		if fileConfig.DryRun {
			config.DryRun = true
		}
		if fileConfig.Yes {
			config.Yes = true
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
	if cmd.Flags().Changed("workers") {
		config.Workers = scanConfig.Workers
	}
	if cmd.Flags().Changed("incremental") {
		config.Incremental = scanConfig.Incremental
	}
	if cmd.Flags().Changed("no-cache") {
		config.NoCache = scanConfig.NoCache
	}
	if cmd.Flags().Changed("watch") {
		config.Watch = scanConfig.Watch
	}
	if cmd.Flags().Changed("fix") {
		config.Fix = scanConfig.Fix
	}
	if cmd.Flags().Changed("dry-run") {
		config.DryRun = scanConfig.DryRun
	}
	if cmd.Flags().Changed("yes") {
		config.Yes = scanConfig.Yes
	}

	return config, nil
}

const (
	watchDebounceDelay = 500 * time.Millisecond
	watchPollInterval  = 250 * time.Millisecond
)

func runWatchScan(config *ScanConfig) int {
	printWatchStatus("Running initial scan...")
	if initialExitCode := executeScan(config); initialExitCode == 2 {
		return 2
	}
	printWatchStatus("Watching for changes... (Press Ctrl+C to stop)")

	lastSnapshot, err := watchSnapshot(config.Dir, config.Ignore, config.EnvFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize watcher snapshot: %v\n", err)
		return 2
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer signal.Stop(sigChan)

	ticker := time.NewTicker(watchPollInterval)
	defer ticker.Stop()

	pendingChange := false
	lastChangeAt := time.Time{}
	lastChangedFile := ""

	for {
		select {
		case <-sigChan:
			fmt.Fprintln(os.Stderr, "")
			printWatchStatus("Watch mode stopped.")
			return 0
		case <-ticker.C:
			nextSnapshot, changedFile, changed, snapErr := checkWatchChanges(config.Dir, config.Ignore, config.EnvFiles, lastSnapshot)
			if snapErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to check changes: %v\n", snapErr)
				continue
			}
			if changed {
				lastSnapshot = nextSnapshot
				pendingChange = true
				lastChangeAt = time.Now()
				lastChangedFile = changedFile
			}
			if pendingChange && time.Since(lastChangeAt) >= watchDebounceDelay {
				fmt.Fprint(os.Stderr, "\033[2J\033[H")
				if lastChangedFile != "" {
					printWatchStatus(fmt.Sprintf("Change detected: %s", lastChangedFile))
				} else {
					printWatchStatus("Change detected")
				}
				printWatchStatus("Running scan...")
				_ = executeScan(config)
				printWatchStatus("Watching for changes... (Press Ctrl+C to stop)")
				pendingChange = false
				lastChangedFile = ""
			}
		}
	}
}

func printWatchStatus(message string) {
	fmt.Fprintf(os.Stderr, "[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

func checkWatchChanges(root string, ignoreDirs []string, envFiles []string, previous map[string]time.Time) (map[string]time.Time, string, bool, error) {
	current, err := watchSnapshot(root, ignoreDirs, envFiles)
	if err != nil {
		return nil, "", false, err
	}

	for path, modTime := range current {
		prevModTime, ok := previous[path]
		if !ok || !modTime.Equal(prevModTime) {
			return current, path, true, nil
		}
	}
	for path := range previous {
		if _, ok := current[path]; !ok {
			return current, path, true, nil
		}
	}

	return current, "", false, nil
}

func watchSnapshot(root string, ignoreDirs []string, envFiles []string) (map[string]time.Time, error) {
	snapshot := make(map[string]time.Time)

	ignoreMap := make(map[string]bool, len(ignoreDirs)+3)
	ignoreMap[".git"] = true
	ignoreMap["node_modules"] = true
	ignoreMap["vendor"] = true
	for _, dir := range ignoreDirs {
		dir = strings.TrimSpace(dir)
		if dir != "" {
			ignoreMap[dir] = true
		}
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			if ignoreMap[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		snapshot[filepath.Clean(path)] = info.ModTime()
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, envFile := range envFiles {
		envPath := filepath.Clean(envFile)
		info, statErr := os.Stat(envPath)
		if statErr == nil {
			snapshot[envPath] = info.ModTime()
			continue
		}
		if filepath.IsAbs(envPath) {
			continue
		}
		joinedPath := filepath.Join(root, envPath)
		joinedInfo, joinedErr := os.Stat(joinedPath)
		if joinedErr == nil {
			snapshot[filepath.Clean(joinedPath)] = joinedInfo.ModTime()
		}
	}

	return snapshot, nil
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
	firstReadableEnvFile := ""

	for _, envFile := range config.EnvFiles {
		parsedDeclared, err := envParser.Parse(envFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse .env file %s: %v\n", envFile, err)
			continue
		}

		if firstReadableEnvFile == "" {
			firstReadableEnvFile = envFile
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
	envFingerprint := computeEnvFingerprint(config.EnvFiles)

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

	used := make(map[string]bool)
	allLocations := make(map[string][]types.Location)
	cacheByFile := make(map[string]cachedFileResult)
	sourceFilesToScan := sourceFiles
	if config.Incremental && !config.NoCache {
		changedFiles, err := detectChangedFiles(config.Dir)
		if err == nil {
			cache, cacheErr := loadScanCache(config.Dir)
			if cacheErr == nil && cache.EnvFingerprint == envFingerprint {
				sourceFilesToScan = make([]string, 0, len(sourceFiles))
				for _, filePath := range sourceFiles {
					rel := relativeToRoot(config.Dir, filePath)
					cachedFile, ok := cache.Files[rel]
					if ok && !changedFiles[rel] {
						cacheByFile[rel] = cachedFile
						for varName, locs := range cachedFile.Variables {
							used[varName] = true
							allLocations[varName] = append(allLocations[varName], locs...)
						}
						continue
					}
					sourceFilesToScan = append(sourceFilesToScan, filePath)
				}
			}
		}
	}

	// Step 3: Detect environment variable usage in source files
	type detectResult struct {
		filePath  string
		locations map[string][]types.Location
		err       error
	}
	jobs := make(chan string)
	results := make(chan detectResult, len(sourceFilesToScan))
	workers := config.Workers
	if len(sourceFilesToScan) == 0 {
		workers = 0
	} else if workers > len(sourceFilesToScan) {
		workers = len(sourceFilesToScan)
	}
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				ext := filepath.Ext(filePath)
				d := detectorFactory.Create(ext)
				if d == nil {
					continue
				}
				locations, err := d.Detect(filePath)
				results <- detectResult{filePath: filePath, locations: locations, err: err}
			}
		}()
	}
	go func() {
		for _, filePath := range sourceFilesToScan {
			jobs <- filePath
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()
	for result := range results {
		if result.err != nil {
			// Soft error: log warning but continue processing
			fmt.Fprintf(os.Stderr, "Warning: failed to process file %s: %v\n", result.filePath, result.err)
			continue
		}
		rel := relativeToRoot(config.Dir, result.filePath)
		cacheByFile[rel] = cachedFileResult{
			Hash:      fileHash(result.filePath),
			Variables: result.locations,
		}
		for varName, locs := range result.locations {
			used[varName] = true
			allLocations[varName] = append(allLocations[varName], locs...)
		}
	}
	sortLocationMap(allLocations)
	if config.Incremental && !config.NoCache {
		if err := saveScanCache(config.Dir, scanCache{
			EnvFingerprint: envFingerprint,
			Files:          cacheByFile,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save cache: %v\n", err)
		}
	}

	// Step 4: Compare declared vs used variables
	diffResult := diffEngine.Compare(declared, used)

	if config.Fix && len(diffResult.Missing) > 0 {
		fixVars := append([]string(nil), diffResult.Missing...)
		targetEnvFile := firstReadableEnvFile

		if config.DryRun {
			fmt.Fprintf(os.Stderr, "Dry run: would add %d missing variable(s) to %s\n", len(fixVars), targetEnvFile)
			for _, varName := range fixVars {
				fmt.Fprintf(os.Stderr, "- %s=\n", varName)
			}
		} else {
			confirmed := config.Yes
			if !confirmed {
				var promptErr error
				confirmed, promptErr = promptForFixConfirmation(targetEnvFile, len(fixVars))
				if promptErr != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to read confirmation: %v\n", promptErr)
					return 2
				}
			}

			if !confirmed {
				fmt.Fprintln(os.Stderr, "Auto-fix canceled.")
			} else {
				if err := applyMissingVarsFix(targetEnvFile, fixVars); err != nil {
					fmt.Fprintf(os.Stderr, "Error: failed to apply --fix: %v\n", err)
					return 2
				}

				for _, varName := range fixVars {
					declared[varName] = true
					declaredSources[varName] = targetEnvFile
				}
				diffResult = diffEngine.Compare(declared, used)

				fmt.Fprintf(os.Stderr, "Auto-fix: added %d variable(s) to %s (backup: %s)\n", len(fixVars), targetEnvFile, targetEnvFile+".backup")
			}
		}
	}

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

func sortLocationMap(locations map[string][]types.Location) {
	for varName := range locations {
		sort.Slice(locations[varName], func(i, j int) bool {
			if locations[varName][i].FilePath == locations[varName][j].FilePath {
				return locations[varName][i].LineNumber < locations[varName][j].LineNumber
			}
			return locations[varName][i].FilePath < locations[varName][j].FilePath
		})
	}
}

type cachedFileResult struct {
	Hash      string                      `json:"hash"`
	Variables map[string][]types.Location `json:"variables"`
}

type scanCache struct {
	EnvFingerprint string                      `json:"env_fingerprint"`
	Files          map[string]cachedFileResult `json:"files"`
}

func cachePath(root string) string {
	return filepath.Join(root, ".capture", "cache.json")
}

func computeEnvFingerprint(envFiles []string) string {
	hasher := sha256.New()
	for _, envFile := range envFiles {
		cleanPath := filepath.Clean(envFile)
		_, _ = hasher.Write([]byte(cleanPath))
		_, _ = hasher.Write([]byte{0})
		content, err := os.ReadFile(envFile)
		if err != nil {
			_, _ = hasher.Write([]byte("ERR"))
			_, _ = hasher.Write([]byte(err.Error()))
			_, _ = hasher.Write([]byte{0})
			continue
		}
		_, _ = hasher.Write(content)
		_, _ = hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func fileHash(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func loadScanCache(root string) (scanCache, error) {
	data, err := os.ReadFile(cachePath(root))
	if err != nil {
		return scanCache{}, err
	}
	var cache scanCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return scanCache{}, err
	}
	if cache.Files == nil {
		cache.Files = make(map[string]cachedFileResult)
	}
	return cache, nil
}

func saveScanCache(root string, cache scanCache) error {
	if cache.Files == nil {
		cache.Files = make(map[string]cachedFileResult)
	}
	cacheDir := filepath.Join(root, ".capture")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(root), data, 0644)
}

func detectChangedFiles(root string) (map[string]bool, error) {
	cmd := exec.Command("git", "-C", root, "status", "--porcelain", "--untracked-files=all")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	changed := make(map[string]bool)
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		pathPart := strings.TrimSpace(line[3:])
		if pathPart == "" {
			continue
		}
		if arrow := strings.LastIndex(pathPart, " -> "); arrow != -1 {
			pathPart = pathPart[arrow+4:]
		}
		pathPart = strings.Trim(pathPart, `"`)
		if pathPart == "" {
			continue
		}
		changed[filepath.Clean(pathPart)] = true
	}
	return changed, nil
}

func relativeToRoot(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(rel)
}

func promptForFixConfirmation(envFile string, count int) (bool, error) {
	fmt.Fprintf(os.Stderr, "Add %d missing variable(s) to %s? [y/N]: ", count, envFile)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	response := strings.TrimSpace(strings.ToLower(line))
	return response == "y" || response == "yes", nil
}

func applyMissingVarsFix(envFile string, vars []string) error {
	if len(vars) == 0 {
		return nil
	}

	originalContent, err := os.ReadFile(envFile)
	if err != nil {
		return err
	}

	info, err := os.Stat(envFile)
	if err != nil {
		return err
	}

	backupPath := envFile + ".backup"
	if err := os.WriteFile(backupPath, originalContent, info.Mode().Perm()); err != nil {
		return err
	}

	var builder strings.Builder
	builder.Write(originalContent)
	if len(originalContent) > 0 && originalContent[len(originalContent)-1] != '\n' {
		builder.WriteByte('\n')
	}
	builder.WriteString("\n# Added by capture on ")
	builder.WriteString(time.Now().Format("2006-01-02"))
	builder.WriteByte('\n')
	for _, varName := range vars {
		builder.WriteString(varName)
		builder.WriteString("=\n")
	}

	return os.WriteFile(envFile, []byte(builder.String()), info.Mode().Perm())
}
