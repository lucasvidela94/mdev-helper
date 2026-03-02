package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sombi/mobile-dev-helper/internal/logs"
	logservice "github.com/sombi/mobile-dev-helper/internal/service/logs"
	"github.com/spf13/cobra"
)

var (
	logsFollow  bool
	logsFilter  string
	logsLevel   string
	logsTag     string
	logsPackage string
	logsLines   int
	logsOutput  string
	logsFormat  string
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream and manage logs from mobile development tools",
	Long: `Stream and manage logs from various mobile development sources.

This command provides unified access to logs from:
- Android devices (via adb logcat)
- Metro bundler (React Native)
- Flutter (via flutter logs)

Examples:
  # Stream Android logs
  mdev logs android

  # Stream Android logs with filtering
  mdev logs android --package com.example.app --priority W

  # Follow logs (continuous streaming)
  mdev logs android --follow

  # Stream Flutter logs
  mdev logs flutter

  # Export logs to file
  mdev logs android --output logs.txt --format raw`,
}

// logsAndroidCmd represents the logs android subcommand
var logsAndroidCmd = &cobra.Command{
	Use:   "android",
	Short: "Stream Android device logs via adb logcat",
	Long: `Stream Android device logs using adb logcat.

This command streams logs from connected Android devices and provides
filtering options for package, tag, and priority level.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		streamer := logservice.NewAndroidStreamer()

		// Check prerequisites
		if err := streamer.CheckPrerequisites(); err != nil {
			return fmt.Errorf("adb not found: %w\n\nMake sure Android SDK is installed and adb is in your PATH", err)
		}

		// Build stream options
		opts := buildStreamOptions()

		// Create context with cancellation
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Handle interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		// Start streaming
		entries, err := streamer.Stream(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to start log streaming: %w", err)
		}

		// Setup formatter
		formatter := logs.NewDefaultFormatter()
		if !isTerminal() {
			formatter.Colorize = false
		}

		// Setup exporter if output file specified
		var exporter *logs.StreamingExporter
		if logsOutput != "" {
			format, err := logs.ParseExportFormat(logsFormat)
			if err != nil {
				return err
			}
			exporter, err = logs.NewStreamingExporter(format, logsOutput)
			if err != nil {
				return fmt.Errorf("failed to create exporter: %w", err)
			}
			defer exporter.Close()
		}

		// Process entries
		for entry := range entries {
			// Format and print
			formatted := formatter.Format(entry)
			fmt.Println(formatted)

			// Export if enabled
			if exporter != nil {
				if err := exporter.Write(entry); err != nil {
					logger.Error("Failed to write to export file", "error", err)
				}
			}
		}

		return nil
	},
}

// logsMetroCmd represents the logs metro subcommand
var logsMetroCmd = &cobra.Command{
	Use:   "metro",
	Short: "Stream Metro bundler logs",
	Long: `Stream logs from Metro bundler for React Native projects.

Note: This command requires Metro to be running. Start Metro with
'npx react-native start' in another terminal before running this command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		streamer := logservice.NewMetroStreamer("")

		// Check if this is a React Native project
		if !streamer.IsReactNativeProject() {
			return fmt.Errorf("not a React Native project\n\nMake sure you're in a React Native project directory with package.json containing react-native dependency")
		}

		// Check prerequisites
		if err := streamer.CheckPrerequisites(); err != nil {
			return err
		}

		// Build stream options
		opts := buildStreamOptions()

		// Create context with cancellation
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Handle interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		// Start streaming
		entries, err := streamer.Stream(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to start log streaming: %w", err)
		}

		// Setup formatter
		formatter := logs.NewDefaultFormatter()
		if !isTerminal() {
			formatter.Colorize = false
		}

		// Process entries
		for entry := range entries {
			formatted := formatter.Format(entry)
			fmt.Println(formatted)
		}

		return nil
	},
}

// logsFlutterCmd represents the logs flutter subcommand
var logsFlutterCmd = &cobra.Command{
	Use:   "flutter",
	Short: "Stream Flutter device logs",
	Long: `Stream logs from Flutter devices using flutter logs.

This command streams logs from connected Flutter devices and emulators.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		streamer := logservice.NewFlutterStreamer()

		// Check prerequisites
		if err := streamer.CheckPrerequisites(); err != nil {
			return fmt.Errorf("flutter not found: %w\n\nMake sure Flutter SDK is installed and flutter is in your PATH", err)
		}

		// Build stream options
		opts := buildStreamOptions()

		// Create context with cancellation
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Handle interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()

		// Start streaming
		entries, err := streamer.Stream(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to start log streaming: %w", err)
		}

		// Setup formatter
		formatter := logs.NewDefaultFormatter()
		if !isTerminal() {
			formatter.Colorize = false
		}

		// Setup exporter if output file specified
		var exporter *logs.StreamingExporter
		if logsOutput != "" {
			format, err := logs.ParseExportFormat(logsFormat)
			if err != nil {
				return err
			}
			exporter, err = logs.NewStreamingExporter(format, logsOutput)
			if err != nil {
				return fmt.Errorf("failed to create exporter: %w", err)
			}
			defer exporter.Close()
		}

		// Process entries
		for entry := range entries {
			// Format and print
			formatted := formatter.Format(entry)
			fmt.Println(formatted)

			// Export if enabled
			if exporter != nil {
				if err := exporter.Write(entry); err != nil {
					logger.Error("Failed to write to export file", "error", err)
				}
			}
		}

		return nil
	},
}

// logsClearCmd represents the logs clear subcommand
var logsClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear device logs",
	Long:  `Clear logs from Android devices or Flutter devices.`,
}

// logsClearAndroidCmd clears Android logs
var logsClearAndroidCmd = &cobra.Command{
	Use:   "android",
	Short: "Clear Android device logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		streamer := logservice.NewAndroidStreamer()
		if err := streamer.ClearLogs(); err != nil {
			return err
		}
		fmt.Println("✓ Android logs cleared")
		return nil
	},
}

// logsClearFlutterCmd clears Flutter logs
var logsClearFlutterCmd = &cobra.Command{
	Use:   "flutter",
	Short: "Clear Flutter device logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		streamer := logservice.NewFlutterStreamer()
		if err := streamer.ClearLogs(); err != nil {
			return err
		}
		fmt.Println("✓ Flutter logs cleared")
		return nil
	},
}

// buildStreamOptions builds StreamOptions from command flags.
func buildStreamOptions() logs.StreamOptions {
	opts := logs.StreamOptions{
		Follow:  logsFollow,
		Filter:  logsFilter,
		Tag:     logsTag,
		Package: logsPackage,
		Lines:   logsLines,
	}

	if logsLevel != "" {
		opts.Level = logs.ParseLogLevel(logsLevel)
	}

	return opts
}

// isTerminal checks if stdout is a terminal.
func isTerminal() bool {
	// Simple check - in production you'd use term.IsTerminal
	return true
}

func init() {
	rootCmd.AddCommand(logsCmd)

	// Add subcommands
	logsCmd.AddCommand(logsAndroidCmd)
	logsCmd.AddCommand(logsMetroCmd)
	logsCmd.AddCommand(logsFlutterCmd)
	logsCmd.AddCommand(logsClearCmd)

	// Add clear subcommands
	logsClearCmd.AddCommand(logsClearAndroidCmd)
	logsClearCmd.AddCommand(logsClearFlutterCmd)

	// Add flags to android command
	logsAndroidCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Continuously stream new log entries")
	logsAndroidCmd.Flags().StringVarP(&logsFilter, "filter", "F", "", "Filter logs by text pattern")
	logsAndroidCmd.Flags().StringVarP(&logsLevel, "priority", "p", "", "Filter by priority level (V/D/I/W/E/F)")
	logsAndroidCmd.Flags().StringVarP(&logsTag, "tag", "t", "", "Filter by tag")
	logsAndroidCmd.Flags().StringVarP(&logsPackage, "package", "P", "", "Filter by package name")
	logsAndroidCmd.Flags().IntVarP(&logsLines, "lines", "l", 0, "Limit number of lines (0 = unlimited)")
	logsAndroidCmd.Flags().StringVarP(&logsOutput, "output", "o", "", "Export logs to file")
	logsAndroidCmd.Flags().StringVar(&logsFormat, "format", "raw", "Export format (raw, json)")

	// Add flags to flutter command
	logsFlutterCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Continuously stream new log entries")
	logsFlutterCmd.Flags().StringVarP(&logsFilter, "filter", "F", "", "Filter logs by text pattern")
	logsFlutterCmd.Flags().StringVarP(&logsLevel, "priority", "p", "", "Filter by priority level (V/D/I/W/E/F)")
	logsFlutterCmd.Flags().StringVarP(&logsTag, "tag", "t", "", "Filter by tag")
	logsFlutterCmd.Flags().IntVarP(&logsLines, "lines", "l", 0, "Limit number of lines (0 = unlimited)")
	logsFlutterCmd.Flags().StringVarP(&logsOutput, "output", "o", "", "Export logs to file")
	logsFlutterCmd.Flags().StringVar(&logsFormat, "format", "raw", "Export format (raw, json)")

	// Add flags to metro command
	logsMetroCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Continuously stream new log entries")
	logsMetroCmd.Flags().StringVarP(&logsFilter, "filter", "F", "", "Filter logs by text pattern")
	logsMetroCmd.Flags().StringVarP(&logsLevel, "priority", "p", "", "Filter by priority level")
	logsMetroCmd.Flags().IntVarP(&logsLines, "lines", "l", 0, "Limit number of lines (0 = unlimited)")
}
