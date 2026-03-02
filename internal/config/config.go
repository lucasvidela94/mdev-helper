// Package config provides configuration loading for the CLI.
package config

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/sombi/mobile-dev-helper/internal/constants"
	"github.com/spf13/viper"
)

// Config represents the CLI configuration.
type Config struct {
	Verbose     bool   `mapstructure:"verbose"`
	DryRun      bool   `mapstructure:"dry_run"`
	ConfigPath  string `mapstructure:"-"`
	CacheDir    string `mapstructure:"cache_dir"`
	LogLevel    string `mapstructure:"log_level"`
	ProjectPath string `mapstructure:"project_path"`
	AndroidHome string `mapstructure:"android_home"`
	JavaHome    string `mapstructure:"java_home"`
	FlutterHome string `mapstructure:"flutter_home"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}

	return &Config{
		Verbose:     false,
		DryRun:      false,
		CacheDir:    path.Join(homeDir, ".cache", constants.CacheDirName),
		LogLevel:    "info",
		ProjectPath: "",
		AndroidHome: "",
		JavaHome:    "",
		FlutterHome: "",
	}
}

// Load loads configuration from file and environment variables.
// Priority: flags > env vars > config file > defaults
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	defaults := DefaultConfig()
	v.SetDefault("verbose", defaults.Verbose)
	v.SetDefault("dry_run", defaults.DryRun)
	v.SetDefault("cache_dir", defaults.CacheDir)
	v.SetDefault("log_level", defaults.LogLevel)
	v.SetDefault("project_path", defaults.ProjectPath)
	v.SetDefault("android_home", defaults.AndroidHome)
	v.SetDefault("java_home", defaults.JavaHome)
	v.SetDefault("flutter_home", defaults.FlutterHome)

	// Set config file location
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Default config location: ~/.mdev.yaml
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		// Check for auto-migration: if new config doesn't exist but old one does, migrate it
		newConfigPath := filepath.Join(homeDir, constants.ConfigFileName+".yaml")
		oldConfigPath := filepath.Join(homeDir, constants.LegacyConfigFileName+".yaml")

		if _, err := os.Stat(newConfigPath); os.IsNotExist(err) {
			if _, err := os.Stat(oldConfigPath); err == nil {
				// Old config exists, new doesn't - migrate it
				if migrateErr := migrateConfig(oldConfigPath, newConfigPath); migrateErr != nil {
					// Log migration error but continue with defaults
					fmt.Fprintf(os.Stderr, "Warning: failed to migrate config: %v\n", migrateErr)
				}
			}
		}

		v.SetConfigName(constants.ConfigFileName)
		v.AddConfigPath(homeDir)
		v.AddConfigPath(".")
	}

	// Environment variables - support both new and legacy prefixes
	// Check for new prefix first
	v.SetEnvPrefix(constants.EnvPrefix)
	v.AutomaticEnv()

	// Also check legacy env vars for backward compatibility
	checkLegacyEnvVars(v)

	// Read config file (if exists)
	_ = v.ReadInConfig()

	// Unmarshal to Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.ConfigPath = v.ConfigFileUsed()

	return &cfg, nil
}

// migrateConfig copies the old config file to the new location.
func migrateConfig(oldPath, newPath string) error {
	src, err := os.Open(oldPath)
	if err != nil {
		return fmt.Errorf("failed to open old config: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(newPath)
	if err != nil {
		return fmt.Errorf("failed to create new config: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Config migrated from %s to %s\n", oldPath, newPath)
	return nil
}

// checkLegacyEnvVars checks for legacy environment variables and applies them
// if the new equivalent is not set. Prints deprecation warnings.
func checkLegacyEnvVars(v *viper.Viper) {
	// Map of legacy env var name -> (new env var name, config key)
	legacyMappings := []struct {
		legacyVar string
		newVar    string
		configKey string
	}{
		{constants.LegacyEnvPrefix + "_VERBOSE", constants.EnvPrefix + "_VERBOSE", "verbose"},
		{constants.LegacyEnvPrefix + "_DRY_RUN", constants.EnvPrefix + "_DRY_RUN", "dry_run"},
		{constants.LegacyEnvPrefix + "_CACHE_DIR", constants.EnvPrefix + "_CACHE_DIR", "cache_dir"},
		{constants.LegacyEnvPrefix + "_LOG_LEVEL", constants.EnvPrefix + "_LOG_LEVEL", "log_level"},
		{constants.LegacyEnvPrefix + "_PROJECT_PATH", constants.EnvPrefix + "_PROJECT_PATH", "project_path"},
		{constants.LegacyEnvPrefix + "_ANDROID_HOME", constants.EnvPrefix + "_ANDROID_HOME", "android_home"},
		{constants.LegacyEnvPrefix + "_JAVA_HOME", constants.EnvPrefix + "_JAVA_HOME", "java_home"},
		{constants.LegacyEnvPrefix + "_FLUTTER_HOME", constants.EnvPrefix + "_FLUTTER_HOME", "flutter_home"},
	}

	for _, mapping := range legacyMappings {
		// Only use legacy if new is not set
		if os.Getenv(mapping.newVar) == "" {
			if legacyValue := os.Getenv(mapping.legacyVar); legacyValue != "" {
				v.Set(mapping.configKey, legacyValue)
				fmt.Fprintf(os.Stderr,
					"Warning: %s is deprecated, use %s instead\n",
					mapping.legacyVar, mapping.newVar)
			}
		}
	}
}

// Save writes the configuration to the config file.
func (c *Config) Save() error {
	if c.ConfigPath == "" {
		// Use default location
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		c.ConfigPath = filepath.Join(homeDir, constants.ConfigFileName+".yaml")
	}

	// Create viper instance and set values
	v := viper.New()
	v.Set("verbose", c.Verbose)
	v.Set("dry_run", c.DryRun)
	v.Set("cache_dir", c.CacheDir)
	v.Set("log_level", c.LogLevel)
	v.Set("project_path", c.ProjectPath)
	v.Set("android_home", c.AndroidHome)
	v.Set("java_home", c.JavaHome)
	v.Set("flutter_home", c.FlutterHome)

	// Ensure directory exists
	configDir := filepath.Dir(c.ConfigPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := v.WriteConfigAs(c.ConfigPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidatePath checks if a given path exists and is a valid directory.
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("cannot access path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// GetEffectiveValue returns the effective value for a config key,
// considering environment variable overrides.
// Priority: env vars > config file
func (c *Config) GetEffectiveValue(key string) (string, error) {
	switch key {
	case "android_home":
		// Check ANDROID_HOME env var first (standard Android SDK env var)
		if envVal := os.Getenv("ANDROID_HOME"); envVal != "" {
			return envVal, nil
		}
		// Then check MDEV_ANDROID_HOME
		if envVal := os.Getenv(constants.EnvPrefix + "_ANDROID_HOME"); envVal != "" {
			return envVal, nil
		}
		// Then check legacy env var
		if envVal := os.Getenv(constants.LegacyEnvPrefix + "_ANDROID_HOME"); envVal != "" {
			return envVal, nil
		}
		return c.AndroidHome, nil
	case "java_home":
		// Check JAVA_HOME env var first (standard JDK env var)
		if envVal := os.Getenv("JAVA_HOME"); envVal != "" {
			return envVal, nil
		}
		// Then check MDEV_JAVA_HOME
		if envVal := os.Getenv(constants.EnvPrefix + "_JAVA_HOME"); envVal != "" {
			return envVal, nil
		}
		// Then check legacy env var
		if envVal := os.Getenv(constants.LegacyEnvPrefix + "_JAVA_HOME"); envVal != "" {
			return envVal, nil
		}
		return c.JavaHome, nil
	case "flutter_home":
		// Check FLUTTER_HOME env var first (standard Flutter env var)
		if envVal := os.Getenv("FLUTTER_HOME"); envVal != "" {
			return envVal, nil
		}
		// Then check MDEV_FLUTTER_HOME
		if envVal := os.Getenv(constants.EnvPrefix + "_FLUTTER_HOME"); envVal != "" {
			return envVal, nil
		}
		// Then check legacy env var
		if envVal := os.Getenv(constants.LegacyEnvPrefix + "_FLUTTER_HOME"); envVal != "" {
			return envVal, nil
		}
		return c.FlutterHome, nil
	case "verbose":
		return fmt.Sprintf("%t", c.Verbose), nil
	case "dry_run":
		return fmt.Sprintf("%t", c.DryRun), nil
	case "cache_dir":
		return c.CacheDir, nil
	case "log_level":
		return c.LogLevel, nil
	case "project_path":
		return c.ProjectPath, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}
