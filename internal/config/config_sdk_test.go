package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/constants"
)

func TestConfigStructWithSDKPaths(t *testing.T) {
	// Test that Config struct has SDK path fields
	cfg := &Config{
		Verbose:     true,
		AndroidHome: "/test/android-sdk",
		JavaHome:    "/test/jdk",
		FlutterHome: "/test/flutter",
	}

	if cfg.AndroidHome != "/test/android-sdk" {
		t.Errorf("AndroidHome = %v, want /test/android-sdk", cfg.AndroidHome)
	}
	if cfg.JavaHome != "/test/jdk" {
		t.Errorf("JavaHome = %v, want /test/jdk", cfg.JavaHome)
	}
	if cfg.FlutterHome != "/test/flutter" {
		t.Errorf("FlutterHome = %v, want /test/flutter", cfg.FlutterHome)
	}
}

func TestDefaultConfigWithSDKPaths(t *testing.T) {
	cfg := DefaultConfig()

	// SDK paths should be empty by default
	if cfg.AndroidHome != "" {
		t.Errorf("AndroidHome = %v, want empty string", cfg.AndroidHome)
	}
	if cfg.JavaHome != "" {
		t.Errorf("JavaHome = %v, want empty string", cfg.JavaHome)
	}
	if cfg.FlutterHome != "" {
		t.Errorf("FlutterHome = %v, want empty string", cfg.FlutterHome)
	}
}

func TestSave(t *testing.T) {
	t.Run("save creates config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test-config.yaml")

		cfg := &Config{
			Verbose:     true,
			LogLevel:    "debug",
			AndroidHome: "/test/android",
			JavaHome:    "/test/java",
			FlutterHome: "/test/flutter",
			ConfigPath:  configPath,
		}

		err := cfg.Save()
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file should exist after Save()")
		}

		// Load and verify content
		loadedCfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if loadedCfg.Verbose != true {
			t.Errorf("Verbose = %v, want true", loadedCfg.Verbose)
		}
		if loadedCfg.LogLevel != "debug" {
			t.Errorf("LogLevel = %v, want debug", loadedCfg.LogLevel)
		}
		if loadedCfg.AndroidHome != "/test/android" {
			t.Errorf("AndroidHome = %v, want /test/android", loadedCfg.AndroidHome)
		}
		if loadedCfg.JavaHome != "/test/java" {
			t.Errorf("JavaHome = %v, want /test/java", loadedCfg.JavaHome)
		}
		if loadedCfg.FlutterHome != "/test/flutter" {
			t.Errorf("FlutterHome = %v, want /test/flutter", loadedCfg.FlutterHome)
		}
	})

	t.Run("save uses default location when ConfigPath is empty", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Temporarily change HOME
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", origHome)

		cfg := &Config{
			Verbose:  true,
			LogLevel: "warn",
		}
		// ConfigPath is empty, should use default

		err := cfg.Save()
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify file was created at default location
		defaultPath := filepath.Join(tmpDir, constants.ConfigFileName+".yaml")
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			t.Error("Config file should exist at default location after Save()")
		}
	})
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid directory",
			path:    t.TempDir(),
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "non-existent path",
			path:    "/nonexistent/path/12345",
			wantErr: true,
		},
		{
			name:    "file instead of directory",
			path:    func() string { f, _ := os.CreateTemp("", "test"); f.Close(); return f.Name() }(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEffectiveValue(t *testing.T) {
	tests := []struct {
		name      string
		configVal string
		envVar    string
		envVal    string
		want      string
	}{
		{
			name:      "android_home from config",
			configVal: "/config/android",
			envVar:    "",
			envVal:    "",
			want:      "/config/android",
		},
		{
			name:      "android_home from ANDROID_HOME env var",
			configVal: "/config/android",
			envVar:    "ANDROID_HOME",
			envVal:    "/env/android",
			want:      "/env/android",
		},
		{
			name:      "java_home from config",
			configVal: "/config/java",
			envVar:    "",
			envVal:    "",
			want:      "/config/java",
		},
		{
			name:      "java_home from JAVA_HOME env var",
			configVal: "/config/java",
			envVar:    "JAVA_HOME",
			envVal:    "/env/java",
			want:      "/env/java",
		},
		{
			name:      "flutter_home from config",
			configVal: "/config/flutter",
			envVar:    "",
			envVal:    "",
			want:      "/config/flutter",
		},
		{
			name:      "flutter_home from FLUTTER_HOME env var",
			configVal: "/config/flutter",
			envVar:    "FLUTTER_HOME",
			envVal:    "/env/flutter",
			want:      "/env/flutter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars first
			os.Unsetenv("ANDROID_HOME")
			os.Unsetenv("MDEV_ANDROID_HOME")
			os.Unsetenv("JAVA_HOME")
			os.Unsetenv("MDEV_JAVA_HOME")
			os.Unsetenv("FLUTTER_HOME")
			os.Unsetenv("MDEV_FLUTTER_HOME")

			// Set env var if specified
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
				defer os.Unsetenv(tt.envVar)
			}

			cfg := &Config{
				AndroidHome: "/config/android",
				JavaHome:    "/config/java",
				FlutterHome: "/config/flutter",
			}

			// Determine which key to test based on env var name
			var key string
			if tt.envVar == "ANDROID_HOME" || tt.name == "android_home from config" {
				key = "android_home"
			} else if tt.envVar == "JAVA_HOME" || tt.name == "java_home from config" {
				key = "java_home"
			} else if tt.envVar == "FLUTTER_HOME" || tt.name == "flutter_home from config" {
				key = "flutter_home"
			}

			got, err := cfg.GetEffectiveValue(key)
			if err != nil {
				t.Errorf("GetEffectiveValue() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetEffectiveValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEffectiveValueUnknownKey(t *testing.T) {
	cfg := &Config{}
	_, err := cfg.GetEffectiveValue("unknown_key")
	if err == nil {
		t.Error("GetEffectiveValue() should return error for unknown key")
	}
}

func TestGetEffectiveValueOtherKeys(t *testing.T) {
	cfg := &Config{
		Verbose:     true,
		DryRun:      false,
		CacheDir:    "/test/cache",
		LogLevel:    "debug",
		ProjectPath: "/test/project",
	}

	tests := []struct {
		key  string
		want string
	}{
		{"verbose", "true"},
		{"dry_run", "false"},
		{"cache_dir", "/test/cache"},
		{"log_level", "debug"},
		{"project_path", "/test/project"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := cfg.GetEffectiveValue(tt.key)
			if err != nil {
				t.Errorf("GetEffectiveValue() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetEffectiveValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadWithSDKPathsFromConfigFile(t *testing.T) {
	configContent := `
verbose: true
android_home: /custom/android-sdk
java_home: /custom/jdk
flutter_home: /custom/flutter
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AndroidHome != "/custom/android-sdk" {
		t.Errorf("AndroidHome = %v, want /custom/android-sdk", cfg.AndroidHome)
	}
	if cfg.JavaHome != "/custom/jdk" {
		t.Errorf("JavaHome = %v, want /custom/jdk", cfg.JavaHome)
	}
	if cfg.FlutterHome != "/custom/flutter" {
		t.Errorf("FlutterHome = %v, want /custom/flutter", cfg.FlutterHome)
	}
}

func TestLoadWithSDKEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(*Config)
	}{
		{
			name: "android_home from MDEV_ANDROID_HOME",
			envVars: map[string]string{
				"MDEV_ANDROID_HOME": "/env/android",
			},
			validate: func(c *Config) {
				val, _ := c.GetEffectiveValue("android_home")
				if val != "/env/android" {
					t.Errorf("android_home = %v, want /env/android", val)
				}
			},
		},
		{
			name: "java_home from MDEV_JAVA_HOME",
			envVars: map[string]string{
				"MDEV_JAVA_HOME": "/env/java",
			},
			validate: func(c *Config) {
				val, _ := c.GetEffectiveValue("java_home")
				if val != "/env/java" {
					t.Errorf("java_home = %v, want /env/java", val)
				}
			},
		},
		{
			name: "flutter_home from MDEV_FLUTTER_HOME",
			envVars: map[string]string{
				"MDEV_FLUTTER_HOME": "/env/flutter",
			},
			validate: func(c *Config) {
				val, _ := c.GetEffectiveValue("flutter_home")
				if val != "/env/flutter" {
					t.Errorf("flutter_home = %v, want /env/flutter", val)
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
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if tt.validate != nil {
				tt.validate(cfg)
			}
		})
	}
}

func TestLegacySDKEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(*Config)
	}{
		{
			name: "legacy MOBILE_DEV_ANDROID_HOME",
			envVars: map[string]string{
				"MOBILE_DEV_ANDROID_HOME": "/legacy/android",
			},
			validate: func(c *Config) {
				val, _ := c.GetEffectiveValue("android_home")
				if val != "/legacy/android" {
					t.Errorf("android_home = %v, want /legacy/android", val)
				}
			},
		},
		{
			name: "legacy MOBILE_DEV_JAVA_HOME",
			envVars: map[string]string{
				"MOBILE_DEV_JAVA_HOME": "/legacy/java",
			},
			validate: func(c *Config) {
				val, _ := c.GetEffectiveValue("java_home")
				if val != "/legacy/java" {
					t.Errorf("java_home = %v, want /legacy/java", val)
				}
			},
		},
		{
			name: "legacy MOBILE_DEV_FLUTTER_HOME",
			envVars: map[string]string{
				"MOBILE_DEV_FLUTTER_HOME": "/legacy/flutter",
			},
			validate: func(c *Config) {
				val, _ := c.GetEffectiveValue("flutter_home")
				if val != "/legacy/flutter" {
					t.Errorf("flutter_home = %v, want /legacy/flutter", val)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear new env vars first
			os.Unsetenv("MDEV_ANDROID_HOME")
			os.Unsetenv("MDEV_JAVA_HOME")
			os.Unsetenv("MDEV_FLUTTER_HOME")

			// Set legacy env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cfg, err := Load("")
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if tt.validate != nil {
				tt.validate(cfg)
			}
		})
	}
}

func TestNewEnvVarTakesPrecedenceOverLegacySDK(t *testing.T) {
	// Set both new and legacy env vars
	os.Setenv("MDEV_ANDROID_HOME", "/new/android")
	os.Setenv("MOBILE_DEV_ANDROID_HOME", "/legacy/android")
	defer func() {
		os.Unsetenv("MDEV_ANDROID_HOME")
		os.Unsetenv("MOBILE_DEV_ANDROID_HOME")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	val, _ := cfg.GetEffectiveValue("android_home")
	if val != "/new/android" {
		t.Errorf("android_home = %v, want /new/android (new env var should take precedence)", val)
	}
}

func TestStandardEnvVarTakesPrecedence(t *testing.T) {
	// Set both standard and MDEV_ env vars
	os.Setenv("ANDROID_HOME", "/standard/android")
	os.Setenv("MDEV_ANDROID_HOME", "/mdev/android")
	defer func() {
		os.Unsetenv("ANDROID_HOME")
		os.Unsetenv("MDEV_ANDROID_HOME")
	}()

	cfg := &Config{AndroidHome: "/config/android"}

	val, _ := cfg.GetEffectiveValue("android_home")
	if val != "/standard/android" {
		t.Errorf("android_home = %v, want /standard/android (standard env var should take precedence)", val)
	}
}
