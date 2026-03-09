package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yhaliwaizman/capture/internal/detector"
	"github.com/yhaliwaizman/capture/internal/diff"
	"github.com/yhaliwaizman/capture/internal/dockerfile"
	"github.com/yhaliwaizman/capture/internal/parser"
	"github.com/yhaliwaizman/capture/internal/reporter"
	"github.com/yhaliwaizman/capture/internal/types"
	"github.com/yhaliwaizman/capture/internal/walker"
)

// ScanConfig holds the configuration for the scan command
type ScanConfig struct {
	Dir     string
	EnvFile string
	Ignore  []string
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
  - Detect variable usage in source code (JS, TS, Go, Python)
  - Report mismatches and inconsistencies`,
	Example: `  capture scan --dir ./project --env-file .env
  capture scan --dir . --env-file .env --ignore vendor,tmp`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runScan,
}

func init() {
	// Define flags
	scanCmd.Flags().StringVar(&scanConfig.Dir, "dir", ".", "Directory to scan (required)")
	scanCmd.Flags().StringVar(&scanConfig.EnvFile, "env-file", ".env", "Path to .env file (required)")
	scanCmd.Flags().StringSliceVar(&scanConfig.Ignore, "ignore", []string{}, "Comma-separated list of directories to ignore")

	// Mark required flags
	scanCmd.MarkFlagRequired("dir")
	scanCmd.MarkFlagRequired("env-file")
}

func runScan(cmd *cobra.Command, args []string) error {
	// Validate .env file exists
	if _, err := os.Stat(scanConfig.EnvFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: .env file does not exist: %s\n", scanConfig.EnvFile)
		return NewExitError(err, 2)
	}

	// Validate directory exists
	if info, err := os.Stat(scanConfig.Dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory does not exist: %s\n", scanConfig.Dir)
		return NewExitError(err, 2)
	} else if err != nil {
		// Handle permission errors
		fmt.Fprintf(os.Stderr, "Error: cannot access directory: %v\n", err)
		return NewExitError(err, 2)
	} else if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: path is not a directory: %s\n", scanConfig.Dir)
		return NewExitError(fmt.Errorf("not a directory"), 2)
	}

	// Trim whitespace from ignore directories
	for i := range scanConfig.Ignore {
		scanConfig.Ignore[i] = strings.TrimSpace(scanConfig.Ignore[i])
	}

	// Execute the scan
	exitCode := executeScan(&scanConfig)

	if exitCode != 0 {
		return NewExitError(nil, exitCode)
	}

	return nil
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

	// Step 1: Parse .env file
	declared, err := envParser.Parse(config.EnvFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse .env file: %v\n", err)
		return 2
	}

	// Step 2: Walk directory tree to find source files
	files, err := fileWalker.Walk(config.Dir, config.Ignore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to walk directory: %v\n", err)
		return 2
	}

	// Step 2.5: Separate Dockerfiles from source files
	var dockerfiles []string
	var sourceFiles []string

	for _, filePath := range files {
		baseName := filepath.Base(filePath)
		isDockerfile := baseName == "Dockerfile" ||
			filepath.Ext(baseName) == ".dockerfile" ||
			strings.HasPrefix(baseName, "Dockerfile")

		if isDockerfile {
			dockerfiles = append(dockerfiles, filePath)
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

	// Step 5: Prepare report data with first location for each missing variable
	reportData := types.ReportData{
		Unused:  diffResult.Unused,
		Missing: make(map[string]types.Location),
	}

	for _, varName := range diffResult.Missing {
		if locs, ok := allLocations[varName]; ok && len(locs) > 0 {
			reportData.Missing[varName] = locs[0]
		}
	}

	// Step 6: Generate report
	rep.Report(reportData)

	// Step 6.5: Report Docker-specific mismatches
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

	// Determine exit code
	if len(diffResult.Unused) > 0 || len(diffResult.Missing) > 0 || dockerMismatches {
		return 1 // Mismatches found
	}
	return 0 // No mismatches
}
