// Package cmd contains the CLI commands.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/sombi/mobile-dev-helper/internal/config"
	"github.com/sombi/mobile-dev-helper/internal/detector/flutter"
	"github.com/sombi/mobile-dev-helper/internal/detector/jdk"
	"github.com/sombi/mobile-dev-helper/internal/detector/sdk"
	"github.com/spf13/cobra"
)

// configCmd represents the config command group
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage mdev configuration",
	Long: `Manage mdev configuration settings.

This command group allows you to initialize, view, and modify your mdev configuration.
Configuration is stored in ~/.mdev.yaml by default.`,
}

// configInitCmd initializes the configuration via interactive wizard
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration with interactive wizard",
	Long: `Initialize mdev configuration via an interactive wizard.

The wizard will:
1. Auto-detect installed SDKs (Android SDK, JDK, Flutter)
2. Present detected paths for confirmation
3. Allow manual entry for missing SDKs
4. Save configuration to ~/.mdev.yaml`,
	Example: `  # Run the initialization wizard
  mdev config init`,
	RunE: runConfigInit,
}

// configGetCmd retrieves a config value
var configGetCmd = &cobra.Command{
	Use:   "get \u003ckey\u003e",
	Short: "Get a configuration value",
	Long: `Retrieve a configuration value by key.

Available keys:
  - android_home: Path to Android SDK
  - java_home: Path to JDK
  - flutter_home: Path to Flutter SDK
  - verbose: Verbose mode setting
  - dry_run: Dry run mode setting
  - cache_dir: Cache directory path
  - log_level: Log level setting
  - project_path: Default project path

Environment variables take precedence over config file values.`,
	Example: `  # Get Android SDK path
  mdev config get android_home

  # Get JDK path
  mdev config get java_home`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// configSetCmd sets a config value
var configSetCmd = &cobra.Command{
	Use:   "set \u003ckey\u003e \u003cvalue\u003e",
	Short: "Set a configuration value",
	Long: `Set a configuration value by key.

Available keys:
  - android_home: Path to Android SDK
  - java_home: Path to JDK
  - flutter_home: Path to Flutter SDK
  - verbose: Verbose mode (true/false)
  - dry_run: Dry run mode (true/false)
  - cache_dir: Cache directory path
  - log_level: Log level (debug/info/warn/error)
  - project_path: Default project path

Paths are validated by default. Use --force to skip validation.`,
	Example: `  # Set Android SDK path
  mdev config set android_home /opt/android-sdk

  # Set JDK path with force (skip validation)
  mdev config set java_home /custom/jdk --force`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configListCmd lists all config values
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long: `Display all configuration values in a table format.

Shows both the config file value and the effective value
(environment variables take precedence).`,
	Example: `  # List all configuration values
  mdev config list`,
	RunE: runConfigList,
}

var configSetForce bool

func init() {
	configSetCmd.Flags().BoolVarP(&configSetForce, "force", "f", false, "Skip path validation")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║         mdev Configuration Initialization Wizard         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Detect Android SDK
	fmt.Println("📱 Detecting Android SDK...")
	sdkInfo := sdk.Detect()
	androidHome := ""
	if sdkInfo.IsValid {
		fmt.Printf("   Found: %s\n", sdkInfo.Path)
		fmt.Printf("   Version: %s\n", sdkInfo.Version)
		if confirm(reader, "   Use this path?") {
			androidHome = sdkInfo.Path
		}
	} else {
		fmt.Printf("   %s\n", sdkInfo.Error)
	}

	if androidHome == "" {
		androidHome = promptForPath(reader, "   Enter Android SDK path (or leave empty to skip):")
	}

	fmt.Println()

	// Detect JDK
	fmt.Println("☕ Detecting JDK...")
	jdkInfo := jdk.Detect()
	javaHome := ""
	if jdkInfo.IsValid {
		fmt.Printf("   Found: %s\n", jdkInfo.Path)
		fmt.Printf("   Version: %s\n", jdkInfo.Version)
		if confirm(reader, "   Use this path?") {
			javaHome = jdkInfo.Path
		}
	} else {
		fmt.Printf("   %s\n", jdkInfo.Error)
	}

	if javaHome == "" {
		javaHome = promptForPath(reader, "   Enter JDK path (or leave empty to skip):")
	}

	fmt.Println()

	// Detect Flutter
	fmt.Println("🐦 Detecting Flutter SDK...")
	flutterInfo := flutter.Detect()
	flutterHome := ""
	if flutterInfo.IsValid {
		fmt.Printf("   Found: %s\n", flutterInfo.Path)
		fmt.Printf("   Version: %s\n", flutterInfo.Version)
		if confirm(reader, "   Use this path?") {
			flutterHome = flutterInfo.Path
		}
	} else {
		fmt.Printf("   %s\n", flutterInfo.Error)
	}

	if flutterHome == "" {
		flutterHome = promptForPath(reader, "   Enter Flutter SDK path (or leave empty to skip):")
	}

	fmt.Println()

	// Show summary
	fmt.Println("📋 Configuration Summary:")
	fmt.Println("─────────────────────────")
	if androidHome != "" {
		fmt.Printf("   Android SDK: %s\n", androidHome)
	} else {
		fmt.Println("   Android SDK: (not set)")
	}
	if javaHome != "" {
		fmt.Printf("   JDK:         %s\n", javaHome)
	} else {
		fmt.Println("   JDK:         (not set)")
	}
	if flutterHome != "" {
		fmt.Printf("   Flutter:     %s\n", flutterHome)
	} else {
		fmt.Println("   Flutter:     (not set)")
	}
	fmt.Println()

	if !confirm(reader, "Save this configuration?") {
		fmt.Println("Configuration cancelled.")
		return nil
	}

	// Update config
	cfg.AndroidHome = androidHome
	cfg.JavaHome = javaHome
	cfg.FlutterHome = flutterHome

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Configuration saved successfully!")
	fmt.Printf("   Location: %s\n", cfg.ConfigPath)

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	value, err := cfg.GetEffectiveValue(key)
	if err != nil {
		return err
	}

	fmt.Println(value)
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Validate key
	validKeys := []string{"android_home", "java_home", "flutter_home", "verbose", "dry_run", "cache_dir", "log_level", "project_path"}
	isValidKey := false
	for _, k := range validKeys {
		if k == key {
			isValidKey = true
			break
		}
	}
	if !isValidKey {
		return fmt.Errorf("invalid config key: %s\nValid keys are: %s", key, strings.Join(validKeys, ", "))
	}

	// Validate path if it's a path key and not forcing
	pathKeys := []string{"android_home", "java_home", "flutter_home", "cache_dir", "project_path"}
	isPathKey := false
	for _, k := range pathKeys {
		if k == key {
			isPathKey = true
			break
		}
	}

	if isPathKey && value != "" && !configSetForce {
		if err := config.ValidatePath(value); err != nil {
			return fmt.Errorf("path validation failed: %w\nUse --force to skip validation", err)
		}
	}

	// Set the value
	switch key {
	case "android_home":
		cfg.AndroidHome = value
	case "java_home":
		cfg.JavaHome = value
	case "flutter_home":
		cfg.FlutterHome = value
	case "verbose":
		cfg.Verbose = strings.ToLower(value) == "true"
	case "dry_run":
		cfg.DryRun = strings.ToLower(value) == "true"
	case "cache_dir":
		cfg.CacheDir = value
	case "log_level":
		cfg.LogLevel = value
	case "project_path":
		cfg.ProjectPath = value
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Set %s = %s\n", key, value)
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "KEY\tCONFIG FILE\tEFFECTIVE VALUE\tSOURCE")
	fmt.Fprintln(w, "────\t───────────\t───────────────\t──────")

	// List all config keys
	keys := []string{"android_home", "java_home", "flutter_home", "verbose", "dry_run", "cache_dir", "log_level", "project_path"}

	for _, key := range keys {
		effectiveValue, _ := cfg.GetEffectiveValue(key)
		configValue := getConfigFileValue(key)

		source := "config file"
		if effectiveValue != configValue {
			source = "env var"
		}
		if configValue == "" {
			configValue = "(not set)"
		}
		if effectiveValue == "" {
			effectiveValue = "(not set)"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", key, configValue, effectiveValue, source)
	}

	w.Flush()
	return nil
}

func getConfigFileValue(key string) string {
	switch key {
	case "android_home":
		return cfg.AndroidHome
	case "java_home":
		return cfg.JavaHome
	case "flutter_home":
		return cfg.FlutterHome
	case "verbose":
		return fmt.Sprintf("%t", cfg.Verbose)
	case "dry_run":
		return fmt.Sprintf("%t", cfg.DryRun)
	case "cache_dir":
		return cfg.CacheDir
	case "log_level":
		return cfg.LogLevel
	case "project_path":
		return cfg.ProjectPath
	default:
		return ""
	}
}

func confirm(reader *bufio.Reader, prompt string) bool {
	fmt.Printf("%s [Y/n]: ", prompt)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

func promptForPath(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	path, _ := reader.ReadString('\n')
	return strings.TrimSpace(path)
}
