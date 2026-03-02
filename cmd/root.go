// Package cmd contains the CLI commands.
package cmd

import (
	"fmt"
	"log/slog"

	"github.com/sombi/mobile-dev-helper/internal/config"
	"github.com/sombi/mobile-dev-helper/internal/constants"
	"github.com/sombi/mobile-dev-helper/internal/logging"
	"github.com/sombi/mobile-dev-helper/internal/version"
	"github.com/spf13/cobra"
)

var (
	cfg        *config.Config
	logger     *slog.Logger
	verbose    bool
	dryRun     bool
	configPath string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   constants.AppName,
	Short: "A CLI tool for managing mobile development environments",
	Long: `mdev helps maintain clean and functional mobile development environments.

It provides commands to clean caches, diagnose issues, and get information about your setup.

Supported frameworks: React Native, Flutter, Ionic, Kotlin`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		var err error
		cfg, err = config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override verbose from flag if provided
		if cmd.Flags().Changed("verbose") {
			cfg.Verbose = verbose
		}

		// Override dry-run from flag if provided
		if cmd.Flags().Changed("dry-run") {
			cfg.DryRun = dryRun
		}

		// Setup logging
		logger = logging.Setup(cfg.Verbose)

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be done without actually doing it")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file (default: ~/.mdev.yaml)")

	// Version flag
	rootCmd.Version = version.Version
	rootCmd.PersistentFlags().BoolP("version", "V", false, "Print version information")

	// Add subcommands
	rootCmd.AddCommand(configCmd)
}
