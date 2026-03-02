// Package cmd contains the CLI commands.
package cmd

import (
	"fmt"

	"github.com/sombi/mobile-dev-helper/internal/config"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long: `Display the current version, build time, and git commit of mdev.

This command shows detailed version information about the mdev binary,
including when it was built and from which git commit.`,
	Example: `  # Show version information
  mdev version`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

// printVersion displays version information.
func printVersion() {
	fmt.Printf("Version:    %s\n", config.Version)
	fmt.Printf("Build Time: %s\n", config.BuildTime)
	fmt.Printf("Git Commit: %s\n", config.GitCommit)
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
