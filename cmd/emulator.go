package cmd

import (
	"fmt"
	"os"

	"github.com/sombi/mobile-dev-helper/internal/detector/emulator"
	"github.com/sombi/mobile-dev-helper/internal/detector/sdk"
	"github.com/spf13/cobra"
)

// emulatorCmd represents the emulator command.
var emulatorCmd = &cobra.Command{
	Use:   "emulator",
	Short: "Manage Android emulators",
	Long: `Manage Android emulators and AVDs.

This command provides subcommands to:
- List available AVDs (Android Virtual Devices)
- Check status of running emulators
- View emulator installation details`,
}

// emulatorListCmd represents the emulator list subcommand.
var emulatorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available Android Virtual Devices",
	Long:  `List all configured AVDs in the Android SDK.`,
	Run: func(cmd *cobra.Command, args []string) {
		// First detect SDK
		sdkInfo := sdk.Detect()
		if !sdkInfo.IsValid {
			fmt.Fprintf(os.Stderr, "Error: %s\n", sdkInfo.Error)
			os.Exit(1)
		}

		// Create emulator detector
		detector := emulator.NewDetector(sdkInfo.Path)
		avds, err := detector.ListAVDs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing AVDs: %v\n", err)
			os.Exit(1)
		}

		if len(avds) == 0 {
			fmt.Println("No AVDs configured.")
			fmt.Println("Create AVDs via Android Studio or using avdmanager.")
			return
		}

		fmt.Println("=== Available Android Virtual Devices ===")
		fmt.Println()

		for _, avd := range avds {
			fmt.Printf("Name: %s\n", avd.Name)
			if avd.Device != "" {
				fmt.Printf("  Device: %s\n", avd.Device)
			}
			if avd.TargetAPI != "" {
				fmt.Printf("  Target: %s\n", avd.TargetAPI)
			}
			if avd.Arch != "" {
				fmt.Printf("  Arch: %s\n", avd.Arch)
			}
			if avd.Error != "" {
				fmt.Printf("  Warning: %s\n", avd.Error)
			}
			fmt.Println()
		}
	},
}

// emulatorStatusCmd represents the emulator status subcommand.
var emulatorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show emulator status and running devices",
	Long:  `Display information about running emulators and emulator installation.`,
	Run: func(cmd *cobra.Command, args []string) {
		// First detect SDK
		sdkInfo := sdk.Detect()
		if !sdkInfo.IsValid {
			fmt.Fprintf(os.Stderr, "Error: %s\n", sdkInfo.Error)
			os.Exit(1)
		}

		// Create emulator detector
		detector := emulator.NewDetector(sdkInfo.Path)
		info := detector.Detect()

		fmt.Println("=== Emulator Status ===")
		fmt.Println()

		// Binary info
		if info.BinaryPath != "" {
			fmt.Printf("Emulator Binary: %s\n", info.BinaryPath)
			if info.Version != "" {
				fmt.Printf("Version: %s\n", info.Version)
			}
		} else {
			fmt.Println("Emulator Binary: Not found")
		}
		fmt.Println()

		// Running emulators
		if len(info.Running) > 0 {
			fmt.Println("Running Emulators:")
			for _, dev := range info.Running {
				fmt.Printf("  %s - %s", dev.DeviceID, dev.Status)
				if dev.Model != "" {
					fmt.Printf(" (%s)", dev.Model)
				}
				fmt.Println()
			}
		} else {
			fmt.Println("Running Emulators: None")
		}
		fmt.Println()

		// AVD count
		fmt.Printf("Available AVDs: %d\n", len(info.AVDs))
	},
}

func init() {
	emulatorCmd.AddCommand(emulatorListCmd)
	emulatorCmd.AddCommand(emulatorStatusCmd)
	rootCmd.AddCommand(emulatorCmd)
}
