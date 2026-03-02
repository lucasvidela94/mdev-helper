package pathtools

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	// Test that Detect returns a non-nil ToolsInfo
	info := Detect()
	if info == nil {
		t.Fatal("Detect() returned nil")
	}

	// All tool fields should be non-nil
	if info.ADB == nil {
		t.Error("Detect() returned nil ADB")
	}
	if info.Emulator == nil {
		t.Error("Detect() returned nil Emulator")
	}
	if info.SDKManager == nil {
		t.Error("Detect() returned nil SDKManager")
	}
	if info.Node == nil {
		t.Error("Detect() returned nil Node")
	}
	if info.Java == nil {
		t.Error("Detect() returned nil Java")
	}
}

func TestDetectTool(t *testing.T) {
	// Test detecting a tool that should exist (at least 'go' should be in PATH)
	info := detectTool("go", func(path string) string {
		return "test-version"
	})

	if info == nil {
		t.Fatal("detectTool() returned nil")
	}

	if info.Name != "go" {
		t.Errorf("detectTool() Name = %v, want 'go'", info.Name)
	}

	// Since 'go' should be in PATH during tests
	goPath, err := exec.LookPath("go")
	if err == nil {
		if !info.InPath {
			t.Error("detectTool() InPath = false, expected true for 'go'")
		}
		if info.Path != goPath {
			t.Errorf("detectTool() Path = %v, want %v", info.Path, goPath)
		}
		if info.Version != "test-version" {
			t.Errorf("detectTool() Version = %v, want 'test-version'", info.Version)
		}
	}
}

func TestDetectTool_NotFound(t *testing.T) {
	// Test detecting a tool that doesn't exist
	info := detectTool("nonexistent-tool-12345", nil)

	if info == nil {
		t.Fatal("detectTool() returned nil")
	}

	if info.InPath {
		t.Error("detectTool() InPath = true for non-existent tool")
	}

	if info.Path != "" {
		t.Errorf("detectTool() Path = %v, want empty string", info.Path)
	}
}

func TestDetectSDKManager(t *testing.T) {
	info := detectSDKManager()

	if info == nil {
		t.Fatal("detectSDKManager() returned nil")
	}

	if info.Name != "sdkmanager" {
		t.Errorf("detectSDKManager() Name = %v, want 'sdkmanager'", info.Name)
	}

	// Result depends on environment, but should not panic
	_ = info.InPath
	_ = info.Path
}

func TestDetectSDKManager_WithAndroidHome(t *testing.T) {
	// Skip if there's a real ANDROID_HOME with sdkmanager in PATH
	if _, err := exec.LookPath("sdkmanager"); err == nil {
		t.Skip("sdkmanager found in PATH, skipping test")
	}

	// Save original ANDROID_HOME
	originalAndroidHome := os.Getenv("ANDROID_HOME")
	originalSDKRoot := os.Getenv("ANDROID_SDK_ROOT")
	defer func() {
		os.Setenv("ANDROID_HOME", originalAndroidHome)
		os.Setenv("ANDROID_SDK_ROOT", originalSDKRoot)
	}()

	// Create a temporary directory structure
	tempDir := t.TempDir()
	os.Setenv("ANDROID_HOME", tempDir)
	os.Unsetenv("ANDROID_SDK_ROOT")

	// Create fake sdkmanager
	cmdlineToolsDir := filepath.Join(tempDir, "cmdline-tools", "latest", "bin")
	os.MkdirAll(cmdlineToolsDir, 0755)

	var sdkmanagerPath string
	if runtime.GOOS == "windows" {
		sdkmanagerPath = filepath.Join(cmdlineToolsDir, "sdkmanager.bat")
	} else {
		sdkmanagerPath = filepath.Join(cmdlineToolsDir, "sdkmanager")
	}

	// Create the file
	os.WriteFile(sdkmanagerPath, []byte("#!/bin/sh\necho test"), 0755)

	info := detectSDKManager()

	if info.Path != sdkmanagerPath {
		t.Errorf("detectSDKManager() Path = %v, want %v", info.Path, sdkmanagerPath)
	}
}

func TestToolsInfo_HasMissingTools(t *testing.T) {
	tests := []struct {
		name  string
		tools *ToolsInfo
		want  bool
	}{
		{
			name: "no missing tools",
			tools: &ToolsInfo{
				ADB:        &ToolInfo{InPath: true},
				Emulator:   &ToolInfo{InPath: true},
				SDKManager: &ToolInfo{InPath: true},
			},
			want: false,
		},
		{
			name: "adb missing",
			tools: &ToolsInfo{
				ADB:        &ToolInfo{InPath: false},
				Emulator:   &ToolInfo{InPath: true},
				SDKManager: &ToolInfo{InPath: true},
			},
			want: true,
		},
		{
			name: "emulator missing",
			tools: &ToolsInfo{
				ADB:        &ToolInfo{InPath: true},
				Emulator:   &ToolInfo{InPath: false},
				SDKManager: &ToolInfo{InPath: true},
			},
			want: true,
		},
		{
			name: "sdkmanager missing",
			tools: &ToolsInfo{
				ADB:        &ToolInfo{InPath: true},
				Emulator:   &ToolInfo{InPath: true},
				SDKManager: &ToolInfo{InPath: false},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tools.HasMissingTools(); got != tt.want {
				t.Errorf("HasMissingTools() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolsInfo_GetMissingTools(t *testing.T) {
	tools := &ToolsInfo{
		ADB:        &ToolInfo{Name: "adb", InPath: false},
		Emulator:   &ToolInfo{Name: "emulator", InPath: true},
		SDKManager: &ToolInfo{Name: "sdkmanager", InPath: false},
	}

	missing := tools.GetMissingTools()

	if len(missing) != 2 {
		t.Errorf("GetMissingTools() returned %d tools, want 2", len(missing))
	}

	// Check that adb and sdkmanager are in the list
	hasADB := false
	hasSDKManager := false
	for _, tool := range missing {
		if tool == "adb" {
			hasADB = true
		}
		if tool == "sdkmanager" {
			hasSDKManager = true
		}
	}

	if !hasADB {
		t.Error("GetMissingTools() should include 'adb'")
	}
	if !hasSDKManager {
		t.Error("GetMissingTools() should include 'sdkmanager'")
	}
}

func TestCheckAllTools(t *testing.T) {
	results := CheckAllTools()

	if len(results) != 5 {
		t.Errorf("CheckAllTools() returned %d results, want 5", len(results))
	}

	// Check that all expected tools are present
	expectedTools := map[string]bool{
		"adb":        false,
		"emulator":   false,
		"sdkmanager": false,
		"node":       false,
		"java":       false,
	}

	for _, result := range results {
		if _, exists := expectedTools[result.Tool]; !exists {
			t.Errorf("Unexpected tool in results: %s", result.Tool)
		}
		expectedTools[result.Tool] = true
	}

	for tool, found := range expectedTools {
		if !found {
			t.Errorf("Missing tool in results: %s", tool)
		}
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single line",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "windows line endings",
			input:    "line1\r\nline2\r\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "trailing newline",
			input:    "line1\nline2\n",
			expected: []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("splitLines() returned %d lines, want %d", len(got), len(tt.expected))
				return
			}
			for i, line := range got {
				if line != tt.expected[i] {
					t.Errorf("splitLines()[%d] = %v, want %v", i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestExtractVersionFromOutput(t *testing.T) {
	// Test with simple output
	output := "version 1.2.3\nmore info"
	version := extractVersionFromOutput(output)

	if version != "version 1.2.3" {
		t.Errorf("extractVersionFromOutput() = %v, want 'version 1.2.3'", version)
	}

	// Test with empty output
	version = extractVersionFromOutput("")
	if version != "" {
		t.Errorf("extractVersionFromOutput() = %v, want empty string", version)
	}
}
