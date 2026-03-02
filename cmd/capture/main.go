package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yhaliwaizman/capture/internal/detector"
	"github.com/yhaliwaizman/capture/internal/diff"
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

	// Step 3: Detect environment variable usage in source files
	used := make(map[string]bool)
	allLocations := make(map[string][]types.Location)

	for _, filePath := range files {
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

	// Determine exit code (Requirements 2.1, 2.2)
	if len(diffResult.Unused) > 0 || len(diffResult.Missing) > 0 {
		return 1 // Mismatches found
	}
	return 0 // No mismatches
}
