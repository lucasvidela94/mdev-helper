// Package suggestions provides shell-specific configuration suggestions.
package suggestions

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/detector/shell"
)

// Suggestion represents a configuration suggestion for the user.
type Suggestion struct {
	Issue      string `json:"issue"`
	Solution   string `json:"solution"`
	Command    string `json:"command,omitempty"`
	ConfigFile string `json:"configFile,omitempty"`
	Priority   string `json:"priority"` // high, medium, low
}

// Generator creates suggestions based on environment issues.
type Generator struct {
	shellInfo *shell.ShellInfo
}

// NewGenerator creates a new suggestion generator.
func NewGenerator() *Generator {
	return &Generator{
		shellInfo: shell.Detect(),
	}
}

// NewGeneratorWithShell creates a suggestion generator with a specific shell.
func NewGeneratorWithShell(shellInfo *shell.ShellInfo) *Generator {
	return &Generator{
		shellInfo: shellInfo,
	}
}

// ForMissingAndroidHome generates a suggestion for missing ANDROID_HOME.
func (g *Generator) ForMissingAndroidHome(possiblePath string) *Suggestion {
	if possiblePath == "" {
		possiblePath = g.getDefaultAndroidHome()
	}

	cmd := g.shellInfo.GetShellExportCommand("ANDROID_HOME", possiblePath)

	return &Suggestion{
		Issue:      "ANDROID_HOME environment variable is not set",
		Solution:   fmt.Sprintf("Add ANDROID_HOME pointing to your Android SDK at %s", possiblePath),
		Command:    cmd,
		ConfigFile: g.shellInfo.ConfigFilePath,
		Priority:   "high",
	}
}

// ForMissingJavaHome generates a suggestion for missing JAVA_HOME.
func (g *Generator) ForMissingJavaHome(javaPath string) *Suggestion {
	cmd := g.shellInfo.GetShellExportCommand("JAVA_HOME", javaPath)

	return &Suggestion{
		Issue:      "JAVA_HOME environment variable is not set",
		Solution:   fmt.Sprintf("Add JAVA_HOME pointing to your JDK installation at %s", javaPath),
		Command:    cmd,
		ConfigFile: g.shellInfo.ConfigFilePath,
		Priority:   "high",
	}
}

// ForMissingPathEntry generates a suggestion for adding a directory to PATH.
func (g *Generator) ForMissingPathEntry(tool, pathEntry string) *Suggestion {
	cmd := g.shellInfo.GetShellAppendPathCommand(pathEntry)

	return &Suggestion{
		Issue:      fmt.Sprintf("%s is not in your PATH", tool),
		Solution:   fmt.Sprintf("Add %s to your PATH", pathEntry),
		Command:    cmd,
		ConfigFile: g.shellInfo.ConfigFilePath,
		Priority:   "medium",
	}
}

// ForMissingSDKPlatforms generates a suggestion for missing SDK platforms.
func (g *Generator) ForMissingSDKPlatforms() *Suggestion {
	return &Suggestion{
		Issue:    "No Android SDK platforms installed",
		Solution: "Install at least one Android platform using sdkmanager",
		Command:  "sdkmanager \"platforms;android-34\"",
		Priority: "high",
	}
}

// ForMissingBuildTools generates a suggestion for missing build tools.
func (g *Generator) ForMissingBuildTools() *Suggestion {
	return &Suggestion{
		Issue:    "No Android SDK build-tools installed",
		Solution: "Install build-tools using sdkmanager",
		Command:  "sdkmanager \"build-tools;34.0.0\"",
		Priority: "high",
	}
}

// ForMiseNotInstalled generates a suggestion for installing mise.
func (g *Generator) ForMiseNotInstalled() *Suggestion {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "brew install mise"
	case "linux":
		cmd = "curl https://mise.run | sh"
	default:
		cmd = "See https://mise.jdx.dev/ for installation instructions"
	}

	return &Suggestion{
		Issue:    "mise is not installed",
		Solution: "Install mise for version management of Node.js, Java, and other tools",
		Command:  cmd,
		Priority: "low",
	}
}

// ForMiseToolNotInstalled generates a suggestion for installing a tool via mise.
func (g *Generator) ForMiseToolNotInstalled(tool, version string) *Suggestion {
	cmd := fmt.Sprintf("mise install %s@%s", tool, version)
	if version == "" {
		cmd = fmt.Sprintf("mise install %s", tool)
	}

	return &Suggestion{
		Issue:    fmt.Sprintf("%s is not installed via mise", tool),
		Solution: fmt.Sprintf("Install %s using mise", tool),
		Command:  cmd,
		Priority: "medium",
	}
}

// getDefaultAndroidHome returns the default Android SDK path for the current OS.
func (g *Generator) getDefaultAndroidHome() string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Android", "sdk")
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return `C:\Users\` + os.Getenv("USERNAME") + `\AppData\Local\Android\Sdk`
		}
		return filepath.Join(localAppData, "Android", "Sdk")
	default: // linux and others
		return filepath.Join(home, "Android", "Sdk")
	}
}

// FormatSuggestion formats a suggestion for display.
func FormatSuggestion(s *Suggestion) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Issue: %s\n", s.Issue))
	b.WriteString(fmt.Sprintf("Solution: %s\n", s.Solution))

	if s.Command != "" {
		b.WriteString(fmt.Sprintf("Command: %s\n", s.Command))
	}

	if s.ConfigFile != "" {
		b.WriteString(fmt.Sprintf("Add to: %s\n", s.ConfigFile))
	}

	return b.String()
}

// GetPriorityColor returns an ANSI color code for the priority level.
func GetPriorityColor(priority string) string {
	switch priority {
	case "high":
		return "\033[31m" // Red
	case "medium":
		return "\033[33m" // Yellow
	case "low":
		return "\033[36m" // Cyan
	default:
		return "\033[0m" // Reset
	}
}

// ResetColor returns the ANSI reset code.
func ResetColor() string {
	return "\033[0m"
}
