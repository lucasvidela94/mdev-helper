package shell

import (
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	// Test that Detect returns a non-nil ShellInfo
	info := Detect()
	if info == nil {
		t.Fatal("Detect() returned nil")
	}

	// Test that Type is set
	if info.Type == "" {
		t.Error("Detect() returned ShellInfo with empty Type")
	}

	// Test that Name is set
	if info.Name == "" {
		t.Error("Detect() returned ShellInfo with empty Name")
	}
}

func TestDetectFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantType ShellType
		wantName string
	}{
		{
			name:     "bash on Unix",
			path:     "/bin/bash",
			wantType: Bash,
			wantName: "bash",
		},
		{
			name:     "zsh on Unix",
			path:     "/bin/zsh",
			wantType: Zsh,
			wantName: "zsh",
		},
		{
			name:     "fish on Unix",
			path:     "/usr/bin/fish",
			wantType: Fish,
			wantName: "fish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := detectFromPath(tt.path)
			if info.Type != tt.wantType {
				t.Errorf("detectFromPath() Type = %v, want %v", info.Type, tt.wantType)
			}
			if info.Name != tt.wantName {
				t.Errorf("detectFromPath() Name = %v, want %v", info.Name, tt.wantName)
			}
		})
	}
}

func TestShellInfo_GetShellExportCommand(t *testing.T) {
	tests := []struct {
		name     string
		shell    ShellType
		variable string
		value    string
		want     string
	}{
		{
			name:     "bash export",
			shell:    Bash,
			variable: "ANDROID_HOME",
			value:    "/path/to/sdk",
			want:     `export ANDROID_HOME="/path/to/sdk"`,
		},
		{
			name:     "zsh export",
			shell:    Zsh,
			variable: "JAVA_HOME",
			value:    "/path/to/jdk",
			want:     `export JAVA_HOME="/path/to/jdk"`,
		},
		{
			name:     "fish export",
			shell:    Fish,
			variable: "PATH",
			value:    "/new/path",
			want:     "set -x PATH /new/path",
		},
		{
			name:     "powershell export",
			shell:    PowerShell,
			variable: "ANDROID_HOME",
			value:    `C:\Android\Sdk`,
			want:     `$env:ANDROID_HOME = "C:\Android\Sdk"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ShellInfo{Type: tt.shell}
			got := info.GetShellExportCommand(tt.variable, tt.value)
			if got != tt.want {
				t.Errorf("GetShellExportCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShellInfo_GetShellAppendPathCommand(t *testing.T) {
	tests := []struct {
		name  string
		shell ShellType
		path  string
		want  string
	}{
		{
			name:  "bash append path",
			shell: Bash,
			path:  "/new/path",
			want:  `export PATH="/new/path:$PATH"`,
		},
		{
			name:  "zsh append path",
			shell: Zsh,
			path:  "/new/path",
			want:  `export PATH="/new/path:$PATH"`,
		},
		{
			name:  "fish append path",
			shell: Fish,
			path:  "/new/path",
			want:  "set -x PATH /new/path $PATH",
		},
		{
			name:  "powershell append path",
			shell: PowerShell,
			path:  `C:\Android\Sdk\platform-tools`,
			want:  "$env:PATH = \"C:\\Android\\Sdk\\platform-tools;$env:PATH\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ShellInfo{Type: tt.shell}
			got := info.GetShellAppendPathCommand(tt.path)
			if got != tt.want {
				t.Errorf("GetShellAppendPathCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShellInfo_IsUnixShell(t *testing.T) {
	tests := []struct {
		name  string
		shell ShellType
		want  bool
	}{
		{"bash is Unix", Bash, true},
		{"zsh is Unix", Zsh, true},
		{"fish is Unix", Fish, true},
		{"powershell is not Unix", PowerShell, false},
		{"cmd is not Unix", Cmd, false},
		{"unknown is not Unix", Unknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ShellInfo{Type: tt.shell}
			if got := info.IsUnixShell(); got != tt.want {
				t.Errorf("IsUnixShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShellInfo_IsWindowsShell(t *testing.T) {
	tests := []struct {
		name  string
		shell ShellType
		want  bool
	}{
		{"bash is not Windows", Bash, false},
		{"zsh is not Windows", Zsh, false},
		{"fish is not Windows", Fish, false},
		{"powershell is Windows", PowerShell, true},
		{"cmd is Windows", Cmd, true},
		{"unknown is not Windows", Unknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ShellInfo{Type: tt.shell}
			if got := info.IsWindowsShell(); got != tt.want {
				t.Errorf("IsWindowsShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConfigFilePath(t *testing.T) {
	// Test that getConfigFilePath returns a non-empty path
	path := getConfigFilePath(".bashrc")
	if path == "" {
		t.Error("getConfigFilePath() returned empty string")
	}

	// Test that it includes the filename
	if !contains(path, ".bashrc") {
		t.Errorf("getConfigFilePath() = %v, should contain '.bashrc'", path)
	}
}

func TestGetFishConfigPath(t *testing.T) {
	path := getFishConfigPath()
	if path == "" {
		t.Error("getFishConfigPath() returned empty string")
	}

	if !contains(path, "config.fish") {
		t.Errorf("getFishConfigPath() = %v, should contain 'config.fish'", path)
	}
}

func TestGetPowerShellConfigPath(t *testing.T) {
	path := getPowerShellConfigPath()
	if path == "" {
		t.Error("getPowerShellConfigPath() returned empty string")
	}

	if !contains(path, "Microsoft.PowerShell_profile.ps1") {
		t.Errorf("getPowerShellConfigPath() = %v, should contain 'Microsoft.PowerShell_profile.ps1'", path)
	}

	// On Windows, path should contain Documents
	if runtime.GOOS == "windows" {
		if !contains(path, "Documents") {
			t.Errorf("getPowerShellConfigPath() = %v, should contain 'Documents' on Windows", path)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
