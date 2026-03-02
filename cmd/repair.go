package cmd

import (
	"fmt"
	"os"

	"github.com/sombi/mobile-dev-helper/internal/detector"
	"github.com/sombi/mobile-dev-helper/internal/service"
	"github.com/spf13/cobra"
)

// repairCmd represents the repair command.
var repairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair broken development tools",
	Long: `Repair broken development tools like gradlew wrapper.

This command will attempt to fix:
- Gradle wrapper (chmod, download missing files)`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		target := args[0]
		projectPath, _ := os.Getwd()

		repairService := service.NewRepairService(projectPath)

		switch target {
		case "gradle":
			result := repairService.RepairGradlew()
			printRepairResult(result)
		case "gradlew":
			result := repairService.RepairGradlew()
			printRepairResult(result)
		default:
			fmt.Fprintf(os.Stderr, "Unknown repair target: %s\n", target)
			os.Exit(1)
		}
	},
}

// repairGradlewCmd specifically repairs gradlew
var repairGradlewCmd = &cobra.Command{
	Use:   "gradlew",
	Short: "Repair Gradle wrapper",
	Long: `Repairs the Gradle wrapper by:
- Making gradlew executable
- Downloading missing wrapper files
- Validating the wrapper works`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		repairService := service.NewRepairService(projectPath)

		result := repairService.RepairGradlew()
		printRepairResult(result)
	},
}

// validateCmd validates if tools are working
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate development tools",
	Long: `Validate that development tools are working properly.

Currently validates:
- Gradle wrapper`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		repairService := service.NewRepairService(projectPath)

		result, err := repairService.ValidateGradlew()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printRepairResult(result)
	},
}

func printRepairResult(result *detector.RepairResult) {
	fmt.Printf("Tool: %s\n", result.Tool)
	fmt.Printf("Action: %s\n", result.Action)

	if result.Success {
		fmt.Printf("Status: ✓ Success\n")
		fmt.Printf("Message: %s\n", result.Message)
	} else {
		fmt.Printf("Status: ✗ Failed\n")
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
		os.Exit(1)
	}
}

func init() {
	repairCmd.AddCommand(repairGradlewCmd)
	repairCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(repairCmd)
}
