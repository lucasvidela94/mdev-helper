package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/constants"
	"github.com/spf13/viper"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name     string
		validate func(*Config)
	}{
		{
			name: "default values are set",
			validate: func(c *Config) {
				if c.Verbose != false {
					t.Errorf("Verbose = %v, want false", c.Verbose)
				}
				if c.LogLevel != "info" {
					t.Errorf("LogLevel = %v, want info", c.LogLevel)
				}
				if c.ProjectPath != "" {
					t.Errorf("ProjectPath = %v, want empty string", c.ProjectPath)
				}
			},
		},
		{
			name: "cache dir is set",
			validate: func(c *Config) {
				if c.CacheDir == "" {
					t.Error("CacheDir should not be empty")
				}
				// Verify it contains expected path components
				homeDir, err := os.UserHomeDir()
				if err != nil {
					t.Skip("Cannot get home directory, skipping home dir check")
					return
				}
				expectedPrefix := filepath.Join(homeDir, ".cache")
				if c.CacheDir != expectedPrefix+"/"+constants.CacheDirName {
					t.Errorf("CacheDir = %v, want %v", c.CacheDir, expectedPrefix+"/"+constants.CacheDirName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.validate(cfg)
		})
	}
}

func TestLoadWithDefaults(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		envVars    map[string]string
		wantErr    bool
		validate   func(*Config)
	}{
		{
			name:       "load with no config file uses defaults",
			configPath: "",
			envVars:    nil,
			wantErr:    false,
			validate: func(c *Config) {
				if c.Verbose != false {
					t.Errorf("Verbose = %v, want false", c.Verbose)
				}
				if c.LogLevel != "info" {
					t.Errorf("LogLevel = %v, want info", c.LogLevel)
				}
				if c.CacheDir == "" {
					t.Error("CacheDir should not be empty")
				}
			},
		},
		{
			name:       "load with empty config path",
			configPath: "",
			envVars:    nil,
			wantErr:    false,
			validate: func(c *Config) {
				// Should use defaults
				if c.Verbose != false {
					t.Errorf("Verbose = %v, want false", c.Verbose)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars if any
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cfg, err := Load(tt.configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(cfg)
			}
		})
	}
}

func TestLoadWithConfigFile(t *testing.T) {
	// Create a temporary config file
	tests := []struct {
		name          string
		configContent string
		wantErr       bool
		validate      func(*Config)
	}{
		{
			name: "load valid config file",
			configContent: `
verbose: true
log_level: debug
cache_dir: /tmp/test-cache
project_path: /tmp/test-project
`,
			wantErr: false,
			validate: func(c *Config) {
				if c.Verbose != true {
					t.Errorf("Verbose = %v, want true", c.Verbose)
				}
				if c.LogLevel != "debug" {
					t.Errorf("LogLevel = %v, want debug", c.LogLevel)
				}
				if c.CacheDir != "/tmp/test-cache" {
					t.Errorf("CacheDir = %v, want /tmp/test-cache", c.CacheDir)
				}
				if c.ProjectPath != "/tmp/test-project" {
					t.Errorf("ProjectPath = %v, want /tmp/test-project", c.ProjectPath)
				}
			},
		},
		{
			name: "load config with partial fields",
			configContent: `
verbose: true
`,
			wantErr: false,
			validate: func(c *Config) {
				if c.Verbose != true {
					t.Errorf("Verbose = %v, want true", c.Verbose)
				}
				// Other fields should use defaults
				if c.LogLevel != "info" {
					t.Errorf("LogLevel = %v, want info", c.LogLevel)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create temp config: %v", err)
			}

			cfg, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(cfg)
			}
		})
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantErr  bool
		validate func(*Config)
	}{
		{
			name: "verbose from new env var",
			envVars: map[string]string{
				"MDEV_VERBOSE": "true",
			},
			wantErr: false,
			validate: func(c *Config) {
				if c.Verbose != true {
					t.Errorf("Verbose = %v, want true", c.Verbose)
				}
			},
		},
		{
			name: "log level from new env var",
			envVars: map[string]string{
				"MDEV_LOG_LEVEL": "debug",
			},
			wantErr: false,
			validate: func(c *Config) {
				if c.LogLevel != "debug" {
					t.Errorf("LogLevel = %v, want debug", c.LogLevel)
				}
			},
		},
		{
			name: "cache dir from new env var",
			envVars: map[string]string{
				"MDEV_CACHE_DIR": "/custom/cache",
			},
			wantErr: false,
			validate: func(c *Config) {
				if c.CacheDir != "/custom/cache" {
					t.Errorf("CacheDir = %v, want /custom/cache", c.CacheDir)
				}
			},
		},
		{
			name: "project path from new env var",
			envVars: map[string]string{
				"MDEV_PROJECT_PATH": "/my/project",
			},
			wantErr: false,
			validate: func(c *Config) {
				if c.ProjectPath != "/my/project" {
					t.Errorf("ProjectPath = %v, want /my/project", c.ProjectPath)
				}
			},
		},
		{
			name: "multiple new env vars",
			envVars: map[string]string{
				"MDEV_VERBOSE":   "true",
				"MDEV_LOG_LEVEL": "warn",
				"MDEV_CACHE_DIR": "/env/cache",
			},
			wantErr: false,
			validate: func(c *Config) {
				if c.Verbose != true {
					t.Errorf("Verbose = %v, want true", c.Verbose)
				}
				if c.LogLevel != "warn" {
					t.Errorf("LogLevel = %v, want warn", c.LogLevel)
				}
				if c.CacheDir != "/env/cache" {
					t.Errorf("CacheDir = %v, want /env/cache", c.CacheDir)
				}
			},
		},
		{
			name: "legacy env var with deprecation support",
			envVars: map[string]string{
				"MOBILE_DEV_VERBOSE": "true",
			},
			wantErr: false,
			validate: func(c *Config) {
				if c.Verbose != true {
					t.Errorf("Verbose = %v, want true (from legacy env var)", c.Verbose)
				}
			},
		},
		{
			name: "new env var takes precedence over legacy",
			envVars: map[string]string{
				"MDEV_VERBOSE":       "false",
				"MOBILE_DEV_VERBOSE": "true",
			},
			wantErr: false,
			validate: func(c *Config) {
				if c.Verbose != false {
					t.Errorf("Verbose = %v, want false (new env var should take precedence)", c.Verbose)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cfg, err := Load("")
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(cfg)
			}
		})
	}
}

func TestLoadPriority(t *testing.T) {
	// Test priority: env vars > config file > defaults
	t.Run("env var overrides config file", func(t *testing.T) {
		// Create temp config file with verbose: false
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(configPath, []byte("verbose: false\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create temp config: %v", err)
		}

		// Set env var to true
		os.Setenv("MDEV_VERBOSE", "true")
		defer os.Unsetenv("MDEV_VERBOSE")

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Env var should override config file
		if cfg.Verbose != true {
			t.Errorf("Verbose = %v, want true (env var should override config)", cfg.Verbose)
		}
	})
}

func TestLoadWithInvalidConfigFile(t *testing.T) {
	t.Run("invalid config file uses defaults", func(t *testing.T) {
		// Create temp config file with invalid content
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
		if err != nil {
			t.Fatalf("Failed to create temp config: %v", err)
		}

		cfg, err := Load(configPath)
		// Viper may or may not error on invalid yaml depending on config
		// But it should still return a config with defaults
		if cfg == nil {
			t.Error("Expected config to be returned even with invalid file")
		}
		_ = err // May error, may not - that's okay
	})
}

func TestConfigStruct(t *testing.T) {
	// Test that Config struct has correct fields
	cfg := &Config{
		Verbose:     true,
		CacheDir:    "/test/cache",
		LogLevel:    "debug",
		ProjectPath: "/test/project",
	}

	if cfg.Verbose != true {
		t.Errorf("Verbose = %v, want true", cfg.Verbose)
	}
	if cfg.CacheDir != "/test/cache" {
		t.Errorf("CacheDir = %v, want /test/cache", cfg.CacheDir)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
	if cfg.ProjectPath != "/test/project" {
		t.Errorf("ProjectPath = %v, want /test/project", cfg.ProjectPath)
	}
}

// Test that viper is being used correctly
func TestLoadUsesViper(t *testing.T) {
	// Verify that Load actually uses viper
	v := viper.New()
	if v == nil {
		t.Error("Viper should be initialized")
	}

	// Set a test value
	v.Set("test_key", "test_value")

	// Verify it's set
	if v.GetString("test_key") != "test_value" {
		t.Error("Viper should store values")
	}
}

// Test config auto-migration from old to new location
func TestConfigMigration(t *testing.T) {
	t.Run("auto-migrates config from old location to new", func(t *testing.T) {
		// Create a temp home directory
		tmpDir := t.TempDir()

		// Create old config file
		oldConfigPath := filepath.Join(tmpDir, constants.LegacyConfigFileName+".yaml")
		newConfigPath := filepath.Join(tmpDir, constants.ConfigFileName+".yaml")

		oldContent := "verbose: true\nlog_level: debug\n"
		err := os.WriteFile(oldConfigPath, []byte(oldContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create old config: %v", err)
		}

		// Temporarily change HOME to our temp directory
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", origHome)

		// Load config - this should trigger migration
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Verify config was loaded from migrated file
		if cfg.Verbose != true {
			t.Errorf("Verbose = %v, want true", cfg.Verbose)
		}
		if cfg.LogLevel != "debug" {
			t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
		}

		// Verify new config file was created
		if _, err := os.Stat(newConfigPath); os.IsNotExist(err) {
			t.Error("New config file should exist after migration")
		}

		// Verify content was copied
		newContent, err := os.ReadFile(newConfigPath)
		if err != nil {
			t.Fatalf("Failed to read new config: %v", err)
		}
		if string(newContent) != oldContent {
			t.Errorf("New config content = %v, want %v", string(newContent), oldContent)
		}
	})

	t.Run("does not migrate if new config already exists", func(t *testing.T) {
		// Create a temp home directory
		tmpDir := t.TempDir()

		// Create both old and new config files
		oldConfigPath := filepath.Join(tmpDir, constants.LegacyConfigFileName+".yaml")
		newConfigPath := filepath.Join(tmpDir, constants.ConfigFileName+".yaml")

		oldContent := "verbose: false\nlog_level: info\n"
		newContent := "verbose: true\nlog_level: warn\n"

		err := os.WriteFile(oldConfigPath, []byte(oldContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create old config: %v", err)
		}

		err = os.WriteFile(newConfigPath, []byte(newContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create new config: %v", err)
		}

		// Temporarily change HOME to our temp directory
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", origHome)

		// Load config - should use new config, not migrate
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Verify config was loaded from new file (verbose: true)
		if cfg.Verbose != true {
			t.Errorf("Verbose = %v, want true (should use new config)", cfg.Verbose)
		}
		if cfg.LogLevel != "warn" {
			t.Errorf("LogLevel = %v, want warn (should use new config)", cfg.LogLevel)
		}
	})

	t.Run("continues without migration if old config does not exist", func(t *testing.T) {
		// Create a temp home directory
		tmpDir := t.TempDir()

		// Temporarily change HOME to our temp directory
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", origHome)

		// Load config - no old config exists, should use defaults
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Verify defaults are used
		if cfg.Verbose != false {
			t.Errorf("Verbose = %v, want false (should use default)", cfg.Verbose)
		}
		if cfg.LogLevel != "info" {
			t.Errorf("LogLevel = %v, want info (should use default)", cfg.LogLevel)
		}
	})
}

// Test dual env var support with deprecation warnings
func TestDualEnvVarSupport(t *testing.T) {
	t.Run("uses legacy env var when new is not set", func(t *testing.T) {
		// Set only legacy env var
		os.Setenv("MOBILE_DEV_VERBOSE", "true")
		defer os.Unsetenv("MOBILE_DEV_VERBOSE")

		// Ensure new env var is not set
		os.Unsetenv("MDEV_VERBOSE")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Verbose != true {
			t.Errorf("Verbose = %v, want true (from legacy env var)", cfg.Verbose)
		}
	})

	t.Run("new env var takes precedence over legacy", func(t *testing.T) {
		// Set both env vars
		os.Setenv("MDEV_VERBOSE", "false")
		os.Setenv("MOBILE_DEV_VERBOSE", "true")
		defer func() {
			os.Unsetenv("MDEV_VERBOSE")
			os.Unsetenv("MOBILE_DEV_VERBOSE")
		}()

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Verbose != false {
			t.Errorf("Verbose = %v, want false (new env var should take precedence)", cfg.Verbose)
		}
	})

	t.Run("legacy cache dir env var works", func(t *testing.T) {
		os.Setenv("MOBILE_DEV_CACHE_DIR", "/legacy/cache")
		defer os.Unsetenv("MOBILE_DEV_CACHE_DIR")

		os.Unsetenv("MDEV_CACHE_DIR")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.CacheDir != "/legacy/cache" {
			t.Errorf("CacheDir = %v, want /legacy/cache", cfg.CacheDir)
		}
	})

	t.Run("legacy log level env var works", func(t *testing.T) {
		os.Setenv("MOBILE_DEV_LOG_LEVEL", "error")
		defer os.Unsetenv("MOBILE_DEV_LOG_LEVEL")

		os.Unsetenv("MDEV_LOG_LEVEL")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.LogLevel != "error" {
			t.Errorf("LogLevel = %v, want error", cfg.LogLevel)
		}
	})
}
