package cmd

import (
	"fmt"
	"os"

	"github.com/sombi/mobile-dev-helper/internal/detector/pathtools"
	"github.com/sombi/mobile-dev-helper/internal/detector/shell"
	"github.com/sombi/mobile-dev-helper/internal/service"
	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command.
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose environment issues",
	Long: `Diagnose common issues with mobile development environments.

This command will check:
- Node.js version and path
- Java/JDK installation
- Android SDK configuration
- Gradle wrapper status
- Environment variables
- Project-specific configurations`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		envService := service.NewEnvironmentService(projectPath)
		report := envService.Detect()

		fmt.Println("=== Environment Diagnosis ===")
		fmt.Println()

		// JDK Status
		fmt.Print("JDK: ")
		if report.JDK != nil && report.JDK.IsValid {
			fmt.Printf("✓ %s (%s) - %s\n", report.JDK.Version, report.JDK.MajorVersion, report.JDK.Vendor)
		} else {
			fmt.Println("✗ Not found")
		}

		// Android SDK Status
		fmt.Print("Android SDK: ")
		if report.SDK != nil && report.SDK.IsValid {
			fmt.Printf("✓ %s\n", report.SDK.Path)
			if len(report.SDK.Platforms) > 0 {
				fmt.Printf("  Platforms: %v\n", report.SDK.Platforms)
			}
			if len(report.SDK.BuildTools) > 0 {
				fmt.Printf("  Build Tools: %v\n", report.SDK.BuildTools)
			}
		} else {
			fmt.Println("✗ Not found")
		}

		// Node.js Status
		fmt.Print("Node.js: ")
		if report.Node != nil && report.Node.IsValid {
			fmt.Printf("✓ %s\n", report.Node.Version)
		} else {
			fmt.Println("✗ Not found")
		}

		// Gradle Status
		fmt.Print("Gradle: ")
		if report.Gradle != nil && report.Gradle.IsValid {
			fmt.Printf("✓ %s\n", report.Gradle.GradleVersion)
		} else if report.Gradle != nil && !report.Gradle.HasWrapper {
			fmt.Println("○ No gradlew found")
		} else {
			fmt.Println("✗ Broken")
		}

		// Expo Status
		if report.Expo != nil && report.Expo.HasExpo {
			fmt.Print("Expo: ")
			if report.Expo.IsManaged {
				fmt.Println("✓ Managed Workflow")
			} else if report.Expo.IsBare {
				fmt.Println("✓ Bare Workflow")
			}
		}

		// Flutter Status
		fmt.Print("Flutter: ")
		if report.Flutter != nil && report.Flutter.IsValid {
			fmt.Printf("✓ %s (%s)\n", report.Flutter.Version, report.Flutter.Channel)
			if report.Flutter.Doctor != nil && len(report.Flutter.Doctor.Issues) > 0 {
				fmt.Printf("  Issues: %d found\n", len(report.Flutter.Doctor.Issues))
			}
		} else {
			fmt.Println("✗ Not found")
		}

		// Dart Status
		fmt.Print("Dart: ")
		if report.Dart != nil && report.Dart.IsValid {
			if report.Dart.IsBundled {
				fmt.Printf("✓ %s (bundled with Flutter)\n", report.Dart.Version)
			} else {
				fmt.Printf("✓ %s\n", report.Dart.Version)
			}
		} else {
			fmt.Println("✗ Not found")
		}

		// Shell Info
		shellInfo := shell.Detect()
		fmt.Printf("Shell: ✓ %s (%s)\n", shellInfo.Name, shellInfo.Type)
		if shellInfo.ConfigFile != "" {
			fmt.Printf("  Config: %s\n", shellInfo.ConfigFilePath)
		}

		// PATH Tools
		tools := pathtools.Detect()
		fmt.Println("\n=== PATH Tools ===")
		fmt.Printf("adb: %s\n", formatToolStatus(tools.ADB))
		fmt.Printf("emulator: %s\n", formatToolStatus(tools.Emulator))
		fmt.Printf("sdkmanager: %s\n", formatToolStatus(tools.SDKManager))
		fmt.Printf("node: %s\n", formatToolStatus(tools.Node))
		fmt.Printf("java: %s\n", formatToolStatus(tools.Java))

		fmt.Println()

		// Warnings
		if len(report.Warnings) > 0 {
			fmt.Println("=== Warnings ===")
			for _, warning := range report.Warnings {
				fmt.Printf("⚠ %s\n", warning)
			}
			fmt.Println()
		}

		// Errors
		if len(report.Errors) > 0 {
			fmt.Println("=== Errors ===")
			for _, err := range report.Errors {
				fmt.Printf("✗ %s\n", err)
			}
			fmt.Println()
		}

		// Summary
		if len(report.Errors) == 0 && len(report.Warnings) == 0 {
			fmt.Println("✓ Environment looks healthy!")
		} else if len(report.Errors) == 0 {
			fmt.Println("✓ No critical errors, but there are warnings.")
		} else {
			fmt.Println("✗ Environment has issues that need attention.")
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

// formatToolStatus formats the status of a tool for display.
func formatToolStatus(tool *pathtools.ToolInfo) string {
	if tool.InPath {
		if tool.Version != "" {
			return fmt.Sprintf("✓ %s (%s)", tool.Path, tool.Version)
		}
		return fmt.Sprintf("✓ %s", tool.Path)
	}
	if tool.Path != "" {
		return fmt.Sprintf("○ Found at %s (not in PATH)", tool.Path)
	}
	return "✗ Not found"
}
