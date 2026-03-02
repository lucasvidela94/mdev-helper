package suggestions

import (
	"runtime"
	"strings"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/detector/shell"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	if gen == nil {
		t.Fatal("NewGenerator() returned nil")
	}
	if gen.shellInfo == nil {
		t.Error("NewGenerator() shellInfo is nil")
	}
}

func TestNewGeneratorWithShell(t *testing.T) {
	shellInfo := &shell.ShellInfo{
		Type:       shell.Bash,
		Name:       "bash",
		ConfigFile: ".bashrc",
	}

	gen := NewGeneratorWithShell(shellInfo)
	if gen == nil {
		t.Fatal("NewGeneratorWithShell() returned nil")
	}
	if gen.shellInfo != shellInfo {
		t.Error("NewGeneratorWithShell() did not set shellInfo correctly")
	}
}

func TestGenerator_ForMissingAndroidHome(t *testing.T) {
	tests := []struct {
		name       string
		shellType  shell.ShellType
		path       string
		wantCmd    string
		wantConfig string
	}{
		{
			name:       "bash with default path",
			shellType:  shell.Bash,
			path:       "",
			wantCmd:    `export ANDROID_HOME="`,
			wantConfig: ".bashrc",
		},
		{
			name:       "zsh with custom path",
			shellType:  shell.Zsh,
			path:       "/custom/android/sdk",
			wantCmd:    `export ANDROID_HOME="/custom/android/sdk"`,
			wantConfig: ".zshrc",
		},
		{
			name:       "fish shell",
			shellType:  shell.Fish,
			path:       "/android/sdk",
			wantCmd:    "set -x ANDROID_HOME /android/sdk",
			wantConfig: "config.fish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shellInfo := &shell.ShellInfo{
				Type:           tt.shellType,
				ConfigFile:     tt.wantConfig,
				ConfigFilePath: "/home/user/" + tt.wantConfig,
			}
			gen := NewGeneratorWithShell(shellInfo)
			suggestion := gen.ForMissingAndroidHome(tt.path)

			if suggestion == nil {
				t.Fatal("ForMissingAndroidHome() returned nil")
			}

			if !strings.Contains(suggestion.Issue, "ANDROID_HOME") {
				t.Errorf("Issue should mention ANDROID_HOME, got: %s", suggestion.Issue)
			}

			if suggestion.Priority != "high" {
				t.Errorf("Priority = %v, want 'high'", suggestion.Priority)
			}

			if !strings.Contains(suggestion.Command, tt.wantCmd) {
				t.Errorf("Command = %v, should contain %v", suggestion.Command, tt.wantCmd)
			}
		})
	}
}

func TestGenerator_ForMissingJavaHome(t *testing.T) {
	shellInfo := &shell.ShellInfo{
		Type:           shell.Bash,
		ConfigFile:     ".bashrc",
		ConfigFilePath: "/home/user/.bashrc",
	}
	gen := NewGeneratorWithShell(shellInfo)
	suggestion := gen.ForMissingJavaHome("/usr/lib/jvm/java-17")

	if suggestion == nil {
		t.Fatal("ForMissingJavaHome() returned nil")
	}

	if !strings.Contains(suggestion.Issue, "JAVA_HOME") {
		t.Errorf("Issue should mention JAVA_HOME, got: %s", suggestion.Issue)
	}

	if suggestion.Priority != "high" {
		t.Errorf("Priority = %v, want 'high'", suggestion.Priority)
	}

	wantCmd := `export JAVA_HOME="/usr/lib/jvm/java-17"`
	if suggestion.Command != wantCmd {
		t.Errorf("Command = %v, want %v", suggestion.Command, wantCmd)
	}
}

func TestGenerator_ForMissingPathEntry(t *testing.T) {
	shellInfo := &shell.ShellInfo{
		Type:           shell.Bash,
		ConfigFile:     ".bashrc",
		ConfigFilePath: "/home/user/.bashrc",
	}
	gen := NewGeneratorWithShell(shellInfo)
	suggestion := gen.ForMissingPathEntry("adb", "/android/sdk/platform-tools")

	if suggestion == nil {
		t.Fatal("ForMissingPathEntry() returned nil")
	}

	if !strings.Contains(suggestion.Issue, "adb") {
		t.Errorf("Issue should mention adb, got: %s", suggestion.Issue)
	}

	if suggestion.Priority != "medium" {
		t.Errorf("Priority = %v, want 'medium'", suggestion.Priority)
	}

	wantCmd := `export PATH="/android/sdk/platform-tools:$PATH"`
	if suggestion.Command != wantCmd {
		t.Errorf("Command = %v, want %v", suggestion.Command, wantCmd)
	}
}

func TestGenerator_ForMissingSDKPlatforms(t *testing.T) {
	gen := NewGenerator()
	suggestion := gen.ForMissingSDKPlatforms()

	if suggestion == nil {
		t.Fatal("ForMissingSDKPlatforms() returned nil")
	}

	if !strings.Contains(suggestion.Issue, "platforms") {
		t.Errorf("Issue should mention platforms, got: %s", suggestion.Issue)
	}

	if suggestion.Priority != "high" {
		t.Errorf("Priority = %v, want 'high'", suggestion.Priority)
	}

	if suggestion.Command == "" {
		t.Error("Command should not be empty")
	}
}

func TestGenerator_ForMissingBuildTools(t *testing.T) {
	gen := NewGenerator()
	suggestion := gen.ForMissingBuildTools()

	if suggestion == nil {
		t.Fatal("ForMissingBuildTools() returned nil")
	}

	if !strings.Contains(suggestion.Issue, "build-tools") {
		t.Errorf("Issue should mention build-tools, got: %s", suggestion.Issue)
	}

	if suggestion.Priority != "high" {
		t.Errorf("Priority = %v, want 'high'", suggestion.Priority)
	}
}

func TestGenerator_ForMiseNotInstalled(t *testing.T) {
	gen := NewGenerator()
	suggestion := gen.ForMiseNotInstalled()

	if suggestion == nil {
		t.Fatal("ForMiseNotInstalled() returned nil")
	}

	if !strings.Contains(suggestion.Issue, "mise") {
		t.Errorf("Issue should mention mise, got: %s", suggestion.Issue)
	}

	if suggestion.Priority != "low" {
		t.Errorf("Priority = %v, want 'low'", suggestion.Priority)
	}

	// Command should be appropriate for the OS
	if runtime.GOOS == "darwin" {
		if suggestion.Command != "brew install mise" {
			t.Errorf("Command = %v, want 'brew install mise'", suggestion.Command)
		}
	} else if runtime.GOOS == "linux" {
		if suggestion.Command != "curl https://mise.run | sh" {
			t.Errorf("Command = %v, want 'curl https://mise.run | sh'", suggestion.Command)
		}
	}
}

func TestGenerator_ForMiseToolNotInstalled(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		tool    string
		version string
		wantCmd string
	}{
		{
			name:    "with version",
			tool:    "java",
			version: "17",
			wantCmd: "mise install java@17",
		},
		{
			name:    "without version",
			tool:    "node",
			version: "",
			wantCmd: "mise install node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := gen.ForMiseToolNotInstalled(tt.tool, tt.version)

			if suggestion == nil {
				t.Fatal("ForMiseToolNotInstalled() returned nil")
			}

			if !strings.Contains(suggestion.Issue, tt.tool) {
				t.Errorf("Issue should mention %s, got: %s", tt.tool, suggestion.Issue)
			}

			if suggestion.Command != tt.wantCmd {
				t.Errorf("Command = %v, want %v", suggestion.Command, tt.wantCmd)
			}

			if suggestion.Priority != "medium" {
				t.Errorf("Priority = %v, want 'medium'", suggestion.Priority)
			}
		})
	}
}

func TestFormatSuggestion(t *testing.T) {
	suggestion := &Suggestion{
		Issue:      "Test issue",
		Solution:   "Test solution",
		Command:    "test command",
		ConfigFile: "/home/user/.bashrc",
		Priority:   "high",
	}

	formatted := FormatSuggestion(suggestion)

	if !strings.Contains(formatted, "Test issue") {
		t.Error("Formatted suggestion should contain the issue")
	}
	if !strings.Contains(formatted, "Test solution") {
		t.Error("Formatted suggestion should contain the solution")
	}
	if !strings.Contains(formatted, "test command") {
		t.Error("Formatted suggestion should contain the command")
	}
	if !strings.Contains(formatted, "/home/user/.bashrc") {
		t.Error("Formatted suggestion should contain the config file")
	}
}

func TestGetPriorityColor(t *testing.T) {
	tests := []struct {
		priority string
		want     string
	}{
		{"high", "\033[31m"},
		{"medium", "\033[33m"},
		{"low", "\033[36m"},
		{"unknown", "\033[0m"},
		{"", "\033[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			got := GetPriorityColor(tt.priority)
			if got != tt.want {
				t.Errorf("GetPriorityColor(%q) = %q, want %q", tt.priority, got, tt.want)
			}
		})
	}
}

func TestResetColor(t *testing.T) {
	if got := ResetColor(); got != "\033[0m" {
		t.Errorf("ResetColor() = %q, want %q", got, "\033[0m")
	}
}

func TestGenerator_getDefaultAndroidHome(t *testing.T) {
	gen := NewGenerator()
	path := gen.getDefaultAndroidHome()

	if path == "" {
		t.Error("getDefaultAndroidHome() returned empty string")
	}

	// Path should contain "Android" or appropriate OS-specific path
	if runtime.GOOS == "darwin" {
		if !strings.Contains(path, "Android") {
			t.Errorf("getDefaultAndroidHome() = %v, should contain 'Android'", path)
		}
	}
}
