package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "capture",
	Short: "Static analysis CLI tool for environment variable mismatches",
	Long: `capture is a static analysis CLI tool that identifies mismatches between 
environment variables declared in .env files, Dockerfiles, and source code.

It supports JavaScript, TypeScript, Go, Python, and Dockerfile analysis.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Handle different error types
		if exitErr, ok := err.(ExitError); ok {
			os.Exit(exitErr.Code)
		}
		// For Cobra flag validation errors, print to stderr and use exit code 2
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}

// ExitError wraps an error with an exit code
type ExitError struct {
	Err  error
	Code int
}

func (e ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// NewExitError creates a new ExitError
func NewExitError(err error, code int) ExitError {
	return ExitError{Err: err, Code: code}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(versionCmd)
}
