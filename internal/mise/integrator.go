// Package mise provides integration with mise CLI tool.
package mise

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// IsInstalled checks if mise is installed on the system.
func IsInstalled() bool {
	_, err := exec.LookPath("mise")
	return err == nil
}

// GetVersion returns the installed mise version.
func GetVersion() (string, error) {
	if !IsInstalled() {
		return "", fmt.Errorf("mise not installed")
	}

	cmd := exec.Command("mise", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// ListTools returns a list of tools managed by mise.
func ListTools() ([]string, error) {
	if !IsInstalled() {
		return nil, fmt.Errorf("mise not installed")
	}

	cmd := exec.Command("mise", "ls", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse JSON output (simple parsing for tool names)
	content := string(output)
	var tools []string

	// Simple extraction - find tool names
	// Format: [{"name":"node","version":"20.0.0",...},...]
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"name"`) {
			// Extract name between quotes
			start := strings.Index(line, `"name"`)
			remain := line[start+6:]
			colonIdx := strings.Index(remain, ":")
			if colonIdx > 0 {
				value := strings.TrimSpace(remain[:colonIdx])
				if len(value) > 1 && value[0] == '"' {
					name := value[1 : len(value)-1]
					tools = append(tools, name)
				}
			}
		}
	}

	return tools, nil
}

// GetToolInfo returns information about a specific tool installed via mise.
func GetToolInfo(toolName string) (*detector.MiseInfo, error) {
	if !IsInstalled() {
		return nil, fmt.Errorf("mise not installed")
	}

	// Check if tool is installed
	cmd := exec.Command("mise", "current", toolName)
	output, err := cmd.Output()
	if err != nil {
		// Tool not installed via mise
		return &detector.MiseInfo{
			Name:        toolName,
			IsInstalled: false,
		}, nil
	}

	version := strings.TrimSpace(string(output))

	// Get tool path
	whichCmd := exec.Command("mise", "exec", "--", "which", toolName)
	pathOutput, _ := whichCmd.Output()
	path := strings.TrimSpace(string(pathOutput))

	return &detector.MiseInfo{
		Name:        toolName,
		Version:     version,
		IsInstalled: true,
		Path:        path,
	}, nil
}

// InstallTool installs a tool using mise.
func InstallTool(toolName, version string) error {
	if !IsInstalled() {
		return fmt.Errorf("mise not installed")
	}

	var args []string
	if version != "" {
		args = []string{"install", fmt.Sprintf("%s@%s", toolName, version)}
	} else {
		args = []string{"install", toolName}
	}

	cmd := exec.Command("mise", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// UseTool activates a tool version in the current directory.
func UseTool(toolName, version string) error {
	if !IsInstalled() {
		return fmt.Errorf("mise not installed")
	}

	args := []string{"use", "-g"}
	if version != "" {
		args = append(args, fmt.Sprintf("%s@%s", toolName, version))
	} else {
		args = append(args, toolName)
	}

	cmd := exec.Command("mise", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
