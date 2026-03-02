package logs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/logs"
)

func TestMetroStreamer_Name(t *testing.T) {
	streamer := NewMetroStreamer("")
	if streamer.Name() != "metro" {
		t.Errorf("Name() = %v, want metro", streamer.Name())
	}
}

func TestMetroStreamer_IsReactNativeProject(t *testing.T) {
	// Create a temporary directory with a mock React Native project
	tmpDir := t.TempDir()

	// Create package.json with react-native
	packageJSON := `{
		"name": "test-app",
		"dependencies": {
			"react-native": "0.72.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	streamer := NewMetroStreamer(tmpDir)
	if !streamer.IsReactNativeProject() {
		t.Error("Expected IsReactNativeProject() to return true")
	}
}

func TestMetroStreamer_IsReactNativeProject_NoRN(t *testing.T) {
	// Create a temporary directory without React Native
	tmpDir := t.TempDir()

	// Create package.json without react-native
	packageJSON := `{
		"name": "test-app",
		"dependencies": {
			"lodash": "4.17.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	streamer := NewMetroStreamer(tmpDir)
	if streamer.IsReactNativeProject() {
		t.Error("Expected IsReactNativeProject() to return false")
	}
}

func TestMetroStreamer_IsReactNativeProject_WithMetroConfig(t *testing.T) {
	// Create a temporary directory with metro.config.js
	tmpDir := t.TempDir()

	// Create metro.config.js
	if err := os.WriteFile(filepath.Join(tmpDir, "metro.config.js"), []byte("module.exports = {};"), 0644); err != nil {
		t.Fatalf("Failed to create metro.config.js: %v", err)
	}

	streamer := NewMetroStreamer(tmpDir)
	if !streamer.IsReactNativeProject() {
		t.Error("Expected IsReactNativeProject() to return true with metro.config.js")
	}
}

func TestMetroStreamer_parseMetroLine(t *testing.T) {
	streamer := NewMetroStreamer("")

	tests := []struct {
		name          string
		line          string
		expectedMsg   string
		expectedLevel logs.LogLevel
	}{
		{
			name:          "info level",
			line:          "[INFO] Metro bundler started",
			expectedMsg:   "Metro bundler started",
			expectedLevel: logs.LevelInfo,
		},
		{
			name:          "error level",
			line:          "[ERROR] Build failed",
			expectedMsg:   "Build failed",
			expectedLevel: logs.LevelError,
		},
		{
			name:          "warning level",
			line:          "[WARN] Deprecated API",
			expectedMsg:   "Deprecated API",
			expectedLevel: logs.LevelWarning,
		},
		{
			name:          "debug level",
			line:          "[DEBUG] Processing file",
			expectedMsg:   "Processing file",
			expectedLevel: logs.LevelDebug,
		},
		{
			name:          "plain line",
			line:          "Just a plain message",
			expectedMsg:   "Just a plain message",
			expectedLevel: logs.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := streamer.parseMetroLine(tt.line)

			if entry.Source != "metro" {
				t.Errorf("Expected source 'metro', got %s", entry.Source)
			}

			if entry.Message != tt.expectedMsg {
				t.Errorf("Expected message %q, got %q", tt.expectedMsg, entry.Message)
			}

			if entry.Level != tt.expectedLevel {
				t.Errorf("Expected level %v, got %v", tt.expectedLevel, entry.Level)
			}

			if entry.Tag != "Metro" {
				t.Errorf("Expected tag 'Metro', got %s", entry.Tag)
			}
		})
	}
}

func TestMetroStreamer_parseLevel(t *testing.T) {
	streamer := NewMetroStreamer("")

	tests := []struct {
		level    string
		expected logs.LogLevel
	}{
		{"DEBUG", logs.LevelDebug},
		{"LOG", logs.LevelDebug},
		{"INFO", logs.LevelInfo},
		{"WARN", logs.LevelWarning},
		{"WARNING", logs.LevelWarning},
		{"ERROR", logs.LevelError},
		{"UNKNOWN", logs.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			got := streamer.parseLevel(tt.level)
			if got != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}
