// Package pathtools provides detection of tools in PATH.
package pathtools

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// ToolsInfo contains information about tools found in PATH.
type ToolsInfo struct {
	ADB        *ToolInfo `json:"adb,omitempty"`
	Emulator   *ToolInfo `json:"emulator,omitempty"`
	SDKManager *ToolInfo `json:"sdkManager,omitempty"`
	Node       *ToolInfo `json:"node,omitempty"`
	Java       *ToolInfo `json:"java,omitempty"`
}

// ToolInfo contains information about a specific tool.
type ToolInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
	InPath  bool   `json:"inPath"`
}

// Detect searches for all required tools in PATH.
func Detect() *ToolsInfo {
	return &ToolsInfo{
		ADB:        detectTool("adb", getADBVersion),
		Emulator:   detectTool("emulator", nil),
		SDKManager: detectSDKManager(),
		Node:       detectTool("node", getNodeVersion),
		Java:       detectTool("java", getJavaVersion),
	}
}

// detectTool attempts to find a tool in PATH and optionally get its version.
func detectTool(name string, versionFunc func(string) string) *ToolInfo {
	info := &ToolInfo{
		Name:   name,
		InPath: false,
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return info
	}

	info.InPath = true
	info.Path = path

	if versionFunc != nil {
		info.Version = versionFunc(path)
	}

	return info
}

// detectSDKManager attempts to find sdkmanager in various locations.
func detectSDKManager() *ToolInfo {
	info := &ToolInfo{
		Name:   "sdkmanager",
		InPath: false,
	}

	// Try to find sdkmanager in PATH first
	path, err := exec.LookPath("sdkmanager")
	if err == nil {
		info.InPath = true
		info.Path = path
		return info
	}

	// Try common sdkmanager locations in Android SDK
	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome == "" {
		androidHome = os.Getenv("ANDROID_SDK_ROOT")
	}

	if androidHome != "" {
		// Check cmdline-tools locations
		possiblePaths := []string{
			androidHome + "/cmdline-tools/latest/bin/sdkmanager",
			androidHome + "/cmdline-tools/bin/sdkmanager",
			androidHome + "/tools/bin/sdkmanager",
		}

		if runtime.GOOS == "windows" {
			for i, p := range possiblePaths {
				possiblePaths[i] = p + ".bat"
			}
		}

		for _, p := range possiblePaths {
			if _, err := os.Stat(p); err == nil {
				info.InPath = false // Not in PATH but found
				info.Path = p
				return info
			}
		}
	}

	return info
}

// getADBVersion gets the version of adb.
func getADBVersion(path string) string {
	cmd := exec.Command(path, "version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse "Android Debug Bridge version 1.0.x"
	// Return first line which contains version
	lines := string(output)
	if len(lines) > 0 {
		// Extract version number
		return extractVersionFromOutput(lines)
	}
	return ""
}

// getNodeVersion gets the version of node.
func getNodeVersion(path string) string {
	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// getJavaVersion gets the version of java.
func getJavaVersion(path string) string {
	cmd := exec.Command(path, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return extractVersionFromOutput(string(output))
}

// extractVersionFromOutput extracts a version string from command output.
func extractVersionFromOutput(output string) string {
	// Simple extraction - look for version patterns
	// This is a basic implementation
	lines := splitLines(output)
	if len(lines) > 0 {
		return lines[0]
	}
	return ""
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
	var current string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else if r != '\r' {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// HasMissingTools returns true if any required tools are missing from PATH.
func (t *ToolsInfo) HasMissingTools() bool {
	return (t.ADB != nil && !t.ADB.InPath) ||
		(t.Emulator != nil && !t.Emulator.InPath) ||
		(t.SDKManager != nil && !t.SDKManager.InPath)
}

// GetMissingTools returns a list of tool names that are missing from PATH.
func (t *ToolsInfo) GetMissingTools() []string {
	var missing []string
	if t.ADB != nil && !t.ADB.InPath {
		missing = append(missing, "adb")
	}
	if t.Emulator != nil && !t.Emulator.InPath {
		missing = append(missing, "emulator")
	}
	if t.SDKManager != nil && !t.SDKManager.InPath {
		missing = append(missing, "sdkmanager")
	}
	return missing
}

// CheckResult contains the result of checking if a tool is available.
type CheckResult struct {
	Tool    string `json:"tool"`
	Found   bool   `json:"found"`
	Path    string `json:"path,omitempty"`
	Version string `json:"version,omitempty"`
}

// CheckAllTools checks all tools and returns detailed results.
func CheckAllTools() []CheckResult {
	tools := Detect()
	return []CheckResult{
		{Tool: "adb", Found: tools.ADB.InPath, Path: tools.ADB.Path, Version: tools.ADB.Version},
		{Tool: "emulator", Found: tools.Emulator.InPath, Path: tools.Emulator.Path, Version: tools.Emulator.Version},
		{Tool: "sdkmanager", Found: tools.SDKManager.InPath || tools.SDKManager.Path != "", Path: tools.SDKManager.Path, Version: tools.SDKManager.Version},
		{Tool: "node", Found: tools.Node.InPath, Path: tools.Node.Path, Version: tools.Node.Version},
		{Tool: "java", Found: tools.Java.InPath, Path: tools.Java.Path, Version: tools.Java.Version},
	}
}

// AddToReport adds PATH tools information to an EnvironmentReport.
func AddToReport(report *detector.EnvironmentReport) {
	tools := Detect()

	// Check for missing tools and add warnings
	if !tools.ADB.InPath {
		if androidHome := os.Getenv("ANDROID_HOME"); androidHome != "" {
			report.Warnings = append(report.Warnings, "adb not found in PATH. Add $ANDROID_HOME/platform-tools to PATH")
		} else {
			report.Warnings = append(report.Warnings, "adb not found in PATH. Install Android SDK platform-tools")
		}
	}

	if !tools.Emulator.InPath {
		if androidHome := os.Getenv("ANDROID_HOME"); androidHome != "" {
			report.Warnings = append(report.Warnings, "emulator not found in PATH. Add $ANDROID_HOME/emulator to PATH")
		}
	}

	if !tools.SDKManager.InPath && tools.SDKManager.Path == "" {
		report.Warnings = append(report.Warnings, "sdkmanager not found. Install Android SDK command line tools")
	}
}
