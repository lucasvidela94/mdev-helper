package cmd

import (
	"fmt"
	"os"

	"github.com/sombi/mobile-dev-helper/internal/detector"
	"github.com/sombi/mobile-dev-helper/internal/service"
	"github.com/spf13/cobra"
)

var (
	cacheType     string
	reinstallFlag bool
)

// cleanCmd represents the clean command.
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean caches and temporary files",
	Long: `Clean caches and temporary files from mobile development environments.

This command will remove:
- Node modules cache
- Gradle cache
- Metro bundler cache
- iOS derived data
- Android build cache

Use --dry-run to see what would be deleted without actually deleting.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		target := args[0]
		projectPath, _ := os.Getwd()

		cacheService := service.NewCacheService(projectPath)

		if cfg.DryRun {
			fmt.Printf("Dry run: would clean %s cache\n", target)
			return
		}

		results, err := cacheService.Clean(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("=== Clean Results ===")
		fmt.Println()

		successCount := 0

		for _, result := range results {
			status := "✓"
			if result.Success {
				successCount++
			} else {
				status = "✗"
			}

			fmt.Printf("%s %s\n", status, result.Path)
			if result.SizeFreed != "" {
				fmt.Printf("  Freed: %s\n", result.SizeFreed)
			}
			if result.Error != "" {
				fmt.Printf("  Error: %s\n", result.Error)
			}
			fmt.Println()
		}

		fmt.Printf("Cleaned %d/%d locations\n", successCount, len(results))
	},
}

// Subcommands for clean
var cleanGradleCmd = &cobra.Command{
	Use:   "gradle",
	Short: "Clean Gradle cache",
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		cacheService := service.NewCacheService(projectPath)

		if cfg.DryRun {
			fmt.Println("Dry run: would clean gradle cache")
			return
		}

		results, err := cacheService.Clean("gradle")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printCleanResults(results)
	},
}

var cleanNPMCmd = &cobra.Command{
	Use:   "npm",
	Short: "Clean npm cache and node_modules",
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		cacheService := service.NewCacheService(projectPath)

		if cfg.DryRun {
			fmt.Println("Dry run: would clean npm cache and node_modules")
			return
		}

		results, err := cacheService.Clean("npm")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printCleanResults(results)
	},
}

var cleanMetroCmd = &cobra.Command{
	Use:   "metro",
	Short: "Clean Metro bundler cache",
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		cacheService := service.NewCacheService(projectPath)

		if cfg.DryRun {
			fmt.Println("Dry run: would clean Metro bundler cache")
			return
		}

		results, err := cacheService.Clean("metro")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printCleanResults(results)
	},
}

var cleanPodsCmd = &cobra.Command{
	Use:   "pods",
	Short: "Clean iOS Pods and DerivedData",
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		cacheService := service.NewCacheService(projectPath)

		if cfg.DryRun {
			fmt.Println("Dry run: would clean iOS Pods and DerivedData")
			return
		}

		results, err := cacheService.Clean("pods")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printCleanResults(results)
	},
}

var cleanAndroidCmd = &cobra.Command{
	Use:   "android",
	Short: "Clean Android build cache",
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		cacheService := service.NewCacheService(projectPath)

		if cfg.DryRun {
			fmt.Println("Dry run: would clean Android build cache")
			return
		}

		results, err := cacheService.Clean("android")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printCleanResults(results)
	},
}

var cleanAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Clean all caches",
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		cacheService := service.NewCacheService(projectPath)

		if cfg.DryRun {
			fmt.Println("Dry run: would clean all caches")
			return
		}

		results, err := cacheService.Clean("all")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printCleanResults(results)
	},
}

// cleanModulesCmd cleans node_modules with optional reinstall.
var cleanModulesCmd = &cobra.Command{
	Use:   "modules",
	Short: "Clean node_modules with auto-detected package manager",
	Long: `Clean node_modules directory and optionally reinstall dependencies.

This command will:
1. Auto-detect the package manager from lock files (bun, pnpm, yarn, npm)
2. Remove the node_modules directory
3. Optionally reinstall dependencies (with --reinstall flag)

The package manager is detected in priority order: bun > pnpm > yarn > npm`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		modulesService := service.NewModulesService(projectPath)

		if cfg.DryRun {
			managerInfo := modulesService.GetManagerInfo()
			if managerInfo.IsDetected {
				fmt.Printf("Dry run: would clean node_modules using %s\n", managerInfo.Manager)
				if reinstallFlag {
					fmt.Printf("Dry run: would then run '%s'\n", managerInfo.InstallCmd)
				}
			} else {
				fmt.Printf("Dry run: %s\n", managerInfo.Error)
			}
			return
		}

		var result *service.ModulesResult
		var err error

		if reinstallFlag {
			result, err = modulesService.CleanAndReinstall()
		} else {
			result, err = modulesService.Clean()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printModulesResult(result)
	},
}

// cleanAndroidFolderCmd cleans the android/ folder with safety checks.
var cleanAndroidFolderCmd = &cobra.Command{
	Use:   "android-folder",
	Short: "Clean the android/ folder with safety checks",
	Long: `Clean the android/ directory with workflow-aware warnings and safety checks.

This command will:
1. Detect the project type (Expo Managed, Expo Bare, or React Native Bare)
2. Check for uncommitted git changes in the android/ folder
3. Show appropriate warnings based on the workflow type
4. Remove the android/ folder
5. Optionally regenerate it for Expo projects (with --reinstall flag)

WARNING: For Bare workflow projects, the android/ folder contains native code
modifications. Make sure you have committed any custom changes before deleting.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		androidService := service.NewAndroidFolderService(projectPath)

		if cfg.DryRun {
			info, err := androidService.GetInfo()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if info.Success {
				fmt.Printf("Dry run: would remove android/ folder (%s)\n", info.SizeFreed)
				for _, warning := range info.Warnings {
					fmt.Printf("  ⚠ %s\n", warning)
				}
			} else {
				fmt.Printf("Dry run: %s\n", info.Error)
			}
			return
		}

		var result *service.AndroidFolderResult
		var err error

		if reinstallFlag {
			result, err = androidService.CleanWithReinstall(cfg.DryRun)
		} else {
			result, err = androidService.Clean(cfg.DryRun)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printAndroidFolderResult(result)
	},
}

var cleanFlutterCmd = &cobra.Command{
	Use:   "flutter",
	Short: "Clean Flutter build cache and pub cache",
	Long: `Clean Flutter build cache, pub cache, and project build directories.

This command will remove:
- ~/.pub-cache (global pub cache)
- Project build/ directory
- Project .dart_tool/ directory
- Flutter tool cache`,
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()
		cacheService := service.NewCacheService(projectPath)

		if cfg.DryRun {
			fmt.Println("Dry run: would clean Flutter cache")
			return
		}

		results, err := cacheService.Clean("flutter")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		printCleanResults(results)
	},
}

func printCleanResults(results []detector.CleanResult) {
	fmt.Println("=== Clean Results ===")
	fmt.Println()

	successCount := 0

	for _, result := range results {
		status := "✓"
		if result.Success {
			successCount++
		} else {
			status = "✗"
		}

		fmt.Printf("%s %s\n", status, result.Path)
		if result.SizeFreed != "" {
			fmt.Printf("  Freed: %s\n", result.SizeFreed)
		}
		if result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
		}
		fmt.Println()
	}

	fmt.Printf("Cleaned %d/%d locations\n", successCount, len(results))
}

func printModulesResult(result *service.ModulesResult) {
	fmt.Println("=== Modules Clean Result ===")
	fmt.Println()

	if result.Success {
		fmt.Println("✓ Successfully cleaned node_modules")
		fmt.Printf("  Package Manager: %s\n", result.Manager)
		fmt.Printf("  Action: %s\n", result.Action)
		if result.SizeFreed != "" {
			fmt.Printf("  Space Freed: %s\n", result.SizeFreed)
		}
		if result.ReinstallCmd != "" {
			fmt.Printf("  Reinstall Command: %s\n", result.ReinstallCmd)
		}
	} else {
		fmt.Println("✗ Failed to clean node_modules")
		if result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
		}
	}
	fmt.Println()
}

func printAndroidFolderResult(result *service.AndroidFolderResult) {
	fmt.Println("=== Android Folder Clean Result ===")
	fmt.Println()

	if result.Success {
		fmt.Println("✓ Successfully cleaned android/ folder")
		fmt.Printf("  Action: %s\n", result.Action)
		if result.SizeFreed != "" {
			fmt.Printf("  Space Freed: %s\n", result.SizeFreed)
		}
		if result.HasGitChanges {
			fmt.Println("  ⚠ Warning: There were uncommitted git changes in android/")
		}
		if result.IsExpoManaged {
			fmt.Println("  ℹ Project Type: Expo Managed workflow")
		}
	} else {
		fmt.Println("✗ Failed to clean android/ folder")
		if result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Println("Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
	}
	fmt.Println()
}

func init() {
	// dry-run flag is now defined as PersistentFlag in root.go

	// Add flags for modules and android-folder commands
	cleanModulesCmd.Flags().BoolVarP(&reinstallFlag, "reinstall", "r", false, "Reinstall dependencies after cleaning")
	cleanAndroidFolderCmd.Flags().BoolVarP(&reinstallFlag, "reinstall", "r", false, "Regenerate android folder after cleaning (Expo projects only)")

	// Add subcommands
	cleanCmd.AddCommand(cleanGradleCmd)
	cleanCmd.AddCommand(cleanNPMCmd)
	cleanCmd.AddCommand(cleanMetroCmd)
	cleanCmd.AddCommand(cleanPodsCmd)
	cleanCmd.AddCommand(cleanAndroidCmd)
	cleanCmd.AddCommand(cleanAllCmd)
	cleanCmd.AddCommand(cleanModulesCmd)
	cleanCmd.AddCommand(cleanAndroidFolderCmd)
	cleanCmd.AddCommand(cleanFlutterCmd)

	rootCmd.AddCommand(cleanCmd)
}
