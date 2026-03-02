// Package node provides Node.js detection functionality.
package node

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// Detect searches for Node.js installations on the system.
func Detect() *detector.NodeInfo {
	// Priority 1: Check .nvmrc in current directory or parents
	if nvmrc := findNvmrc(); nvmrc != "" {
		info := detectFromPath()
		if info != nil && info.IsValid {
			info.Path = nvmrc + " (via .nvmrc)"
			return info
		}
	}

	// Priority 2: Check .node-version in current directory or parents
	if nodeVersion := findNodeVersionFile(); nodeVersion != "" {
		info := detectFromPath()
		if info != nil && info.IsValid {
			info.Path = nodeVersion + " (via .node-version)"
			return info
		}
	}

	// Priority 3: Try to find node in PATH
	return detectFromPath()
}

// findNvmrc searches for .nvmrc file in current directory and parents.
func findNvmrc() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		nvmrcPath := filepath.Join(dir, ".nvmrc")
		if _, err := os.Stat(nvmrcPath); err == nil {
			return nvmrcPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// findNodeVersionFile searches for .node-version file.
func findNodeVersionFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		versionPath := filepath.Join(dir, ".node-version")
		if _, err := os.Stat(versionPath); err == nil {
			return versionPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// detectFromPath finds node in PATH.
func detectFromPath() *detector.NodeInfo {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		return &detector.NodeInfo{
			IsValid: false,
			Error:   "No Node.js found. Install Node.js or set up via mise/nvm",
		}
	}

	// Get version
	version, err := getNodeVersion(nodePath)
	if err != nil {
		return &detector.NodeInfo{
			Path:    nodePath,
			IsValid: false,
			Error:   "Could not determine Node.js version: " + err.Error(),
		}
	}

	return &detector.NodeInfo{
		Path:         nodePath,
		Version:      version.full,
		MajorVersion: version.major,
		IsValid:      true,
	}
}

// versionInfo holds parsed version information
type versionInfo struct {
	full  string
	major string
}

// getNodeVersion runs `node -v` and parses the output.
func getNodeVersion(nodePath string) (versionInfo, error) {
	cmd := exec.Command(nodePath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return versionInfo{}, err
	}

	versionStr := strings.TrimSpace(string(output))
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Extract major version
	major := versionStr
	if idx := strings.Index(versionStr, "."); idx > 0 {
		major = versionStr[:idx]
	}

	// Handle major-only version
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		// Could also check via node -p process.versions.node
	}

	return versionInfo{full: versionStr, major: major}, nil
}

// DetectProjectNodeRequirements checks package.json for engines requirements.
func DetectProjectNodeRequirements(projectPath string) (string, error) {
	packageJSONPath := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return "", fmt.Errorf("no package.json found")
	}

	// Simple parsing for "engines" field
	content := string(data)
	enginesIdx := strings.Index(content, `"engines"`)
	if enginesIdx == -1 {
		return "", fmt.Errorf("no engines field in package.json")
	}

	// Find node version requirement
	start := enginesIdx
	end := start + len(`"engines"`)

	// Find the node field
	nodeIdx := strings.Index(content[end:], `"node"`)
	if nodeIdx == -1 {
		return "", fmt.Errorf("no node version requirement")
	}

	nodeStart := end + nodeIdx + len(`"node"`)
	// Find the version string after "node":
	remaining := content[nodeStart:]

	// Simple extraction - find the version string between quotes
	quoteStart := strings.Index(remaining, `"`)
	quoteEnd := strings.Index(remaining[quoteStart+1:], `"`)

	if quoteStart == -1 || quoteEnd == -1 {
		return "", fmt.Errorf("could not parse node version")
	}

	version := remaining[quoteStart+1 : quoteStart+1+quoteEnd]
	return version, nil
}
