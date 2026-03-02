package cmd

import (
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/config"
	"github.com/sombi/mobile-dev-helper/internal/logging"
	"github.com/spf13/cobra"
)

func TestRootCommandExecute(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		description string
	}{
		{
			name:        "no args shows help",
			args:        []string{},
			wantErr:     false,
			description: "Running without args should show help",
		},
		{
			name:        "help flag",
			args:        []string{"--help"},
			wantErr:     false,
			description: "Help flag should work",
		},
		{
			name:        "version flag",
			args:        []string{"--version"},
			wantErr:     false,
			description: "Version flag should work",
		},
		{
			name:        "verbose flag",
			args:        []string{"--verbose"},
			wantErr:     false,
			description: "Verbose flag should not error",
		},
		{
			name:        "invalid flag",
			args:        []string{"--invalid-flag"},
			wantErr:     true,
			description: "Invalid flag should error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the root command for each test
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRootCommandFlags(t *testing.T) {
	tests := []struct {
		name      string
		flagName  string
		flagValue string
		wantErr   bool
	}{
		{
			name:      "verbose long flag",
			flagName:  "verbose",
			flagValue: "true",
			wantErr:   false,
		},
		{
			name:      "config flag",
			flagName:  "config",
			flagValue: "/tmp/test-config.yaml",
			wantErr:   false,
		},
		{
			name:      "version pseudo flag",
			flagName:  "version",
			flagValue: "true",
			wantErr:   true, // Version is a special flag in Cobra
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command instance to test flags
			cmd := &cobra.Command{
				Use:   "test",
				Short: "test",
				Run:   func(cmd *cobra.Command, args []string) {},
			}

			var verbose bool
			var configPath string
			cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
			cmd.Flags().StringVar(&configPath, "config", "", "Path to config file")

			err := cmd.Flags().Set(tt.flagName, tt.flagValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Flags.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRootCommandPersistentPreRunE(t *testing.T) {
	// Test the PersistentPreRunE function with different configurations
	tests := []struct {
		name       string
		configPath string
		verbose    bool
		wantErr    bool
	}{
		{
			name:       "default config",
			configPath: "",
			verbose:    false,
			wantErr:    false,
		},
		{
			name:       "non-existent config path",
			configPath: "/tmp/non-existent-config-12345.yaml",
			verbose:    false,
			wantErr:    false, // Viper will use defaults if file doesn't exist
		},
		{
			name:       "verbose mode",
			configPath: "",
			verbose:    true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
				PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
					// This mimics the actual PersistentPreRunE behavior
					cfg, err := config.Load(tt.configPath)
					if err != nil {
						return err
					}

					// Override verbose from flag if provided
					if cmd.Flags().Changed("verbose") {
						cfg.Verbose = tt.verbose
					}

					// Setup logging
					logger = logging.Setup(cfg.Verbose)

					return nil
				},
				Run: func(cmd *cobra.Command, args []string) {},
			}

			// Add flags to command
			var verbose bool
			var configPath string
			cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
			cmd.Flags().StringVar(&configPath, "config", "", "Path to config file")

			// Set flag values if testing with specific values
			if tt.verbose {
				cmd.Flags().Set("verbose", "true")
			}
			if tt.configPath != "" {
				cmd.Flags().Set("config", tt.configPath)
			}

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("PersistentPreRunE.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRootCommandVersion(t *testing.T) {
	// Verify version is set correctly
	if rootCmd.Version == "" {
		t.Error("Version should be set on root command")
	}

	// Test version command shows version
	rootCmd.SetArgs([]string{"--version"})

	oldVersion := rootCmd.Version
	rootCmd.Version = "test-version-1.0.0"

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Version command failed: %v", err)
	}

	rootCmd.Version = oldVersion
}

func TestRootCommandSubcommands(t *testing.T) {
	// Verify subcommands are registered
	subcommands := []string{"info", "doctor", "clean"}

	for _, sub := range subcommands {
		t.Run(sub, func(t *testing.T) {
			found := false
			for _, c := range rootCmd.Commands() {
				if c.Name() == sub {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Subcommand %q not found in root command", sub)
			}
		})
	}
}

func TestExecuteReturnsError(t *testing.T) {
	// Test that Execute returns an error when command fails
	// We can simulate this by setting invalid args after init
	// Note: This test verifies error handling works

	rootCmd.SetArgs([]string{"--invalid-arg-that-does-not-exist"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid argument")
	}
}

func TestDryRunPersistentFlag(t *testing.T) {
	// Test that dry-run flag is available on root command and inherited by subcommands
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "dry-run flag on root",
			args:    []string{"--dry-run"},
			wantErr: false,
		},
		{
			name:    "dry-run flag with clean subcommand",
			args:    []string{"clean", "--dry-run"},
			wantErr: false,
		},
		{
			name:    "dry-run flag with clean gradle subcommand",
			args:    []string{"clean", "gradle", "--dry-run"},
			wantErr: false,
		},
		{
			name:    "dry-run short flag -n on root",
			args:    []string{"-n"},
			wantErr: false,
		},
		{
			name:    "dry-run short flag -n with clean subcommand",
			args:    []string{"clean", "-n"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the root command for each test
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
