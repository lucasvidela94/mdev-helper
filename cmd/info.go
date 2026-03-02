package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sombi/mobile-dev-helper/internal/service"
	"github.com/spf13/cobra"
)

// infoCmd represents the info command.
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show environment information",
	Long: `Display information about the mobile development environment.

This command will show:
- Detected frameworks (React Native, Flutter, etc.)
- Installed tool versions
- Environment paths
- Configuration settings`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		envService := service.NewEnvironmentService(projectPath)
		report := envService.Detect()

		// Check for JSON output flag
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if jsonOutput {
			jsonBytes, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonBytes))
			return
		}

		fmt.Println("=== Environment Information ===")
		fmt.Println()

		// JDK Info
		fmt.Println("--- JDK ---")
		if report.JDK != nil {
			fmt.Printf("Path: %s\n", report.JDK.Path)
			fmt.Printf("Version: %s\n", report.JDK.Version)
			fmt.Printf("Major Version: %s\n", report.JDK.MajorVersion)
			fmt.Printf("Vendor: %s\n", report.JDK.Vendor)
			fmt.Printf("JAVA_HOME: %s\n", report.JDK.JavaHome)
		} else {
			fmt.Println("Not detected")
		}
		fmt.Println()

		// Android SDK Info
		fmt.Println("--- Android SDK ---")
		if report.SDK != nil {
			fmt.Printf("Path: %s\n", report.SDK.Path)
			fmt.Printf("Version: %s\n", report.SDK.Version)
			fmt.Printf("Build Tools: %v\n", report.SDK.BuildTools)
			fmt.Printf("Platforms: %v\n", report.SDK.Platforms)
			if report.SDK.NDK != "" {
				fmt.Printf("NDK: %s\n", report.SDK.NDK)
			}
			if report.SDK.CommandLineTools != "" {
				fmt.Printf("Command Line Tools: %s\n", report.SDK.CommandLineTools)
			}
		} else {
			fmt.Println("Not detected")
		}
		fmt.Println()

		// Node.js Info
		fmt.Println("--- Node.js ---")
		if report.Node != nil {
			fmt.Printf("Path: %s\n", report.Node.Path)
			fmt.Printf("Version: %s\n", report.Node.Version)
		} else {
			fmt.Println("Not detected")
		}
		fmt.Println()

		// Gradle Info
		fmt.Println("--- Gradle Wrapper ---")
		if report.Gradle != nil {
			fmt.Printf("Path: %s\n", report.Gradle.Path)
			fmt.Printf("Version: %s\n", report.Gradle.GradleVersion)
			fmt.Printf("Valid: %v\n", report.Gradle.IsValid)
			fmt.Printf("JAVA_HOME: %s\n", report.Gradle.JavaHome)
		} else {
			fmt.Println("Not detected")
		}
		fmt.Println()

		// Expo Info
		if report.Expo != nil && report.Expo.HasExpo {
			fmt.Println("--- Expo ---")
			fmt.Printf("Project Path: %s\n", report.Expo.ProjectPath)
			fmt.Printf("Workflow: ")
			if report.Expo.IsManaged {
				fmt.Println("Managed")
			} else if report.Expo.IsBare {
				fmt.Println("Bare")
			}
			if report.Expo.ExpoVersion != "" {
				fmt.Printf("SDK Version: %s\n", report.Expo.ExpoVersion)
			}
			fmt.Println()
		}

		// Mise Tools
		if len(report.MiseTools) > 0 {
			fmt.Println("--- Mise Tools ---")
			for _, tool := range report.MiseTools {
				fmt.Printf("%s: %s\n", tool.Name, tool.Version)
			}
			fmt.Println()
		}

		// Metadata
		fmt.Println("--- Metadata ---")
		fmt.Printf("Report Time: %s\n", report.Timestamp)
	},
}

func init() {
	infoCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	rootCmd.AddCommand(infoCmd)
}
