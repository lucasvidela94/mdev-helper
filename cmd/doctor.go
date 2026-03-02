package cmd

import (
	"fmt"
	"os"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
	"github.com/sombi/mobile-dev-helper/internal/doctor/checker"
	"github.com/sombi/mobile-dev-helper/internal/doctor/formatter"
	"github.com/spf13/cobra"
)

var (
	doctorFormat   string
	doctorVerbose  bool
	doctorNoColor  bool
	doctorExitCode bool
)

// doctorCmd represents the doctor command.
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose environment issues",
	Long: `Diagnose common issues with mobile development environments.

This command performs comprehensive checks on your mobile development environment:
- Disk space availability
- Port availability (Metro, ADB)
- Tool versions (Node.js, Java, Android SDK, Flutter)
- Duplicate SDK installations
- Performance optimizations

Exit codes:
  0 - All checks passed
  1 - Warnings present (no errors)
  2 - Errors present

Use --format=json for machine-readable output suitable for CI/CD pipelines.`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)

	doctorCmd.Flags().StringVarP(&doctorFormat, "format", "f", "human", "Output format (human, json)")
	doctorCmd.Flags().BoolVarP(&doctorVerbose, "verbose", "v", false, "Show detailed information")
	doctorCmd.Flags().BoolVar(&doctorNoColor, "no-color", false, "Disable colored output")
	doctorCmd.Flags().BoolVar(&doctorExitCode, "exit-code", true, "Exit with non-zero code on errors/warnings")
}

func runDoctor(cmd *cobra.Command, args []string) error {
	projectPath, _ := os.Getwd()

	// Create runner and register all checkers
	runner := doctor.NewRunner()
	registerCheckers(runner, projectPath)

	// Run all checks
	report := runner.Run()

	// Get formatter
	f, ok := formatter.DefaultRegistry.Get(doctorFormat)
	if !ok {
		return fmt.Errorf("unknown format: %s (available: %v)", doctorFormat, formatter.DefaultRegistry.List())
	}

	// Configure formatter options for human format
	if _, ok := f.(*formatter.HumanFormatter); ok {
		opts := formatter.FormatOptions{
			UseColors:  !doctorNoColor,
			Verbose:    doctorVerbose,
			ShowPassed: true,
		}
		// Create new formatter with options (HumanFormatter doesn't have SetOptions, so we get a new one)
		f = formatter.NewHumanFormatter(opts)
	}

	// Format and output report
	if err := f.Format(report, os.Stdout); err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	// Exit with appropriate code if requested
	if doctorExitCode {
		os.Exit(report.Summary.ExitCode)
	}

	return nil
}

// registerCheckers registers all available checkers with the runner.
func registerCheckers(runner *doctor.Runner, projectPath string) {
	// Environment checks
	runner.Register(checker.NewDiskChecker("."))
	runner.Register(checker.NewPortChecker())
	runner.Register(checker.NewDuplicateSDKChecker())

	// Tool checks
	runner.Register(checker.NewToolVersionChecker())

	// Performance checks
	runner.Register(checker.NewPerformanceChecker(projectPath))
}
