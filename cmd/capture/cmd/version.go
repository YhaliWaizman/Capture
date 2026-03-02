package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information (can be set via ldflags during build)
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version, commit, and build date of capture.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("capture version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
	},
}
