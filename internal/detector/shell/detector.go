// Package shell provides shell detection functionality.
package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ShellType represents the type of shell.
type ShellType string

const (
	Bash       ShellType = "bash"
	Zsh        ShellType = "zsh"
	Fish       ShellType = "fish"
	PowerShell ShellType = "powershell"
	Cmd        ShellType = "cmd"
	Unknown    ShellType = "unknown"
)

// ShellInfo contains information about the detected shell.
type ShellInfo struct {
	Type           ShellType `json:"type"`
	Name           string    `json:"name"`
	Path           string    `json:"path"`
	ConfigFile     string    `json:"configFile"`
	ConfigFilePath string    `json:"configFilePath"`
}

// Detect identifies the current shell and returns its information.
func Detect() *ShellInfo {
	// Priority 1: Check SHELL environment variable (Unix systems)
	if shellPath := os.Getenv("SHELL"); shellPath != "" && runtime.GOOS != "windows" {
		return detectFromPath(shellPath)
	}

	// Priority 2: Check Windows-specific environment variables
	if runtime.GOOS == "windows" {
		return detectWindowsShell()
	}

	// Priority 3: Try to detect from parent process or fallback
	return detectFallback()
}

// detectFromPath determines shell type from the executable path.
func detectFromPath(shellPath string) *ShellInfo {
	base := filepath.Base(shellPath)
	// Remove .exe extension if present
	base = strings.TrimSuffix(base, ".exe")

	info := &ShellInfo{
		Path: shellPath,
		Name: base,
	}

	switch {
	case strings.Contains(base, "bash"):
		info.Type = Bash
		info.ConfigFile = ".bashrc"
		info.ConfigFilePath = getConfigFilePath(".bashrc")
	case strings.Contains(base, "zsh"):
		info.Type = Zsh
		info.ConfigFile = ".zshrc"
		info.ConfigFilePath = getConfigFilePath(".zshrc")
	case strings.Contains(base, "fish"):
		info.Type = Fish
		info.ConfigFile = "config.fish"
		info.ConfigFilePath = getFishConfigPath()
	default:
		info.Type = Unknown
	}

	return info
}

// detectWindowsShell detects the shell on Windows systems.
func detectWindowsShell() *ShellInfo {
	// Check for PowerShell Core (pwsh)
	if _, err := os.Stat(`C:\Program Files\PowerShell\7\pwsh.exe`); err == nil {
		return &ShellInfo{
			Type:           PowerShell,
			Name:           "pwsh",
			Path:           `C:\Program Files\PowerShell\7\pwsh.exe`,
			ConfigFile:     "Microsoft.PowerShell_profile.ps1",
			ConfigFilePath: getPowerShellConfigPath(),
		}
	}

	// Check for Windows PowerShell
	if _, err := os.Stat(`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`); err == nil {
		return &ShellInfo{
			Type:           PowerShell,
			Name:           "powershell",
			Path:           `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
			ConfigFile:     "Microsoft.PowerShell_profile.ps1",
			ConfigFilePath: getPowerShellConfigPath(),
		}
	}

	// Default to Command Prompt
	return &ShellInfo{
		Type:           Cmd,
		Name:           "cmd",
		Path:           `C:\Windows\System32\cmd.exe`,
		ConfigFile:     "",
		ConfigFilePath: "",
	}
}

// detectFallback provides a fallback detection method.
func detectFallback() *ShellInfo {
	// Try common shell paths
	shellPaths := []string{
		"/bin/bash",
		"/usr/bin/bash",
		"/bin/zsh",
		"/usr/bin/zsh",
		"/usr/local/bin/bash",
		"/usr/local/bin/zsh",
		"/bin/fish",
		"/usr/bin/fish",
	}

	for _, path := range shellPaths {
		if _, err := os.Stat(path); err == nil {
			return detectFromPath(path)
		}
	}

	return &ShellInfo{
		Type: Unknown,
		Name: "unknown",
	}
}

// getConfigFilePath returns the full path to a config file in the home directory.
func getConfigFilePath(filename string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, filename)
}

// getFishConfigPath returns the path to fish config file.
func getFishConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "fish", "config.fish")
}

// getPowerShellConfigPath returns the path to PowerShell profile.
func getPowerShellConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// PowerShell Core profile path
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	}

	// macOS/Linux PowerShell Core
	return filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
}

// GetShellExportCommand returns the appropriate export command for the shell.
func (s *ShellInfo) GetShellExportCommand(variable, value string) string {
	switch s.Type {
	case Fish:
		return "set -x " + variable + " " + value
	case PowerShell:
		return "$env:" + variable + " = \"" + value + "\""
	case Cmd:
		return "set " + variable + "=" + value
	default: // bash, zsh, and others
		return "export " + variable + "=\"" + value + "\""
	}
}

// GetShellAppendPathCommand returns the appropriate command to append to PATH.
func (s *ShellInfo) GetShellAppendPathCommand(path string) string {
	switch s.Type {
	case Fish:
		return "set -x PATH " + path + " $PATH"
	case PowerShell:
		return "$env:PATH = \"" + path + ";$env:PATH\""
	case Cmd:
		return "set PATH=" + path + ";%PATH%"
	default: // bash, zsh, and others
		return "export PATH=\"" + path + ":$PATH\""
	}
}

// IsUnixShell returns true if the shell is a Unix-style shell.
func (s *ShellInfo) IsUnixShell() bool {
	return s.Type == Bash || s.Type == Zsh || s.Type == Fish
}

// IsWindowsShell returns true if the shell is a Windows shell.
func (s *ShellInfo) IsWindowsShell() bool {
	return s.Type == PowerShell || s.Type == Cmd
}
