package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yhaliwaizman/capture/internal/detector"
	"github.com/yhaliwaizman/capture/internal/diff"
	"github.com/yhaliwaizman/capture/internal/dockerfile"
	"github.com/yhaliwaizman/capture/internal/parser"
	"github.com/yhaliwaizman/capture/internal/reporter"
	"github.com/yhaliwaizman/capture/internal/types"
	"github.com/yhaliwaizman/capture/internal/walker"
)

// CLIFlags holds the parsed command-line flags
type CLIFlags struct {
	Root    string
	EnvFile string
	Ignore  []string
}

func main() {
	flags, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	exitCode := run(flags)
	os.Exit(exitCode)
}

// parseFlags parses and validates command-line flags
// Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
func parseFlags() (*CLIFlags, error) {
	// Check if "scan" command is provided
	if len(os.Args) < 2 || os.Args[1] != "scan" {
		return nil, fmt.Errorf("usage: capture scan --root <dir> --env-file <file> [--ignore <dirs>]")
	}

	// Create a new flag set for the scan command
	scanCmd := flag.NewFlagSet("scan", flag.ContinueOnError)
	scanCmd.SetOutput(os.Stderr)

	root := scanCmd.String("root", "", "Root directory to scan (required)")
	envFile := scanCmd.String("env-file", "", "Path to .env file (required)")
	ignore := scanCmd.String("ignore", "", "Comma-separated list of directories to ignore (optional)")

	// Parse flags starting from os.Args[2] (after "scan")
	if err := scanCmd.Parse(os.Args[2:]); err != nil {
		return nil, err
	}

	// Validate required flags
	if *root == "" {
		return nil, fmt.Errorf("--root flag is required")
	}
	if *envFile == "" {
		return nil, fmt.Errorf("--env-file flag is required")
	}

	// Parse ignore directories
	var ignoreDirs []string
	if *ignore != "" {
		ignoreDirs = strings.Split(*ignore, ",")
		// Trim whitespace from each directory name
		for i := range ignoreDirs {
			ignoreDirs[i] = strings.TrimSpace(ignoreDirs[i])
		}
	}

	return &CLIFlags{
		Root:    *root,
		EnvFile: *envFile,
		Ignore:  ignoreDirs,
	}, nil
}

// run executes the main analysis pipeline
// Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 11.5, 12.1
func run(flags *CLIFlags) int {
	// Validate .env file exists (Requirement 2.3)
	if _, err := os.Stat(flags.EnvFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: .env file does not exist: %s\n", flags.EnvFile)
		return 2
	}

	// Validate root directory exists (Requirement 2.4)
	if info, err := os.Stat(flags.Root); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: root directory does not exist: %s\n", flags.Root)
		return 2
	} else if err != nil {
		// Handle permission errors (Requirement 2.5)
		fmt.Fprintf(os.Stderr, "Error: cannot access root directory: %v\n", err)
		return 2
	} else if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: root path is not a directory: %s\n", flags.Root)
		return 2
	}

	// Initialize components
	envParser := parser.NewEnvParser()
	fileWalker := walker.NewFileWalker()
	detectorFactory := detector.NewDetectorFactory()
	diffEngine := diff.NewDiffEngine()
	rep := reporter.NewReporter(os.Stdout, os.Stderr)

	// Step 1: Parse .env file
	declared, err := envParser.Parse(flags.EnvFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse .env file: %v\n", err)
		return 2
	}

	// Step 2: Walk directory tree to find source files
	files, err := fileWalker.Walk(flags.Root, flags.Ignore)
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

	// Determine exit code (Requirements 2.1, 2.2)
	if len(diffResult.Unused) > 0 || len(diffResult.Missing) > 0 || dockerMismatches {
		return 1 // Mismatches found
	}
	return 0 // No mismatches
}
