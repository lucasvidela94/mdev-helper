// Package packagemgr provides package manager detection functionality.
package packagemgr

import (
	"fmt"
	"os"
	"path/filepath"
)

// PackageManager represents a detected package manager.
type PackageManager string

const (
	// NPM represents npm package manager.
	NPM PackageManager = "npm"
	// Yarn represents yarn package manager.
	Yarn PackageManager = "yarn"
	// PNPM represents pnpm package manager.
	PNPM PackageManager = "pnpm"
	// Bun represents bun package manager.
	Bun PackageManager = "bun"
	// Unknown represents an unknown package manager.
	Unknown PackageManager = "unknown"
)

// ManagerInfo contains information about a detected package manager.
type ManagerInfo struct {
	Manager    PackageManager `json:"manager"`         // Detected package manager
	InstallCmd string         `json:"installCmd"`      // Install command (e.g., "npm install")
	LockFile   string         `json:"lockFile"`        // Path to lock file
	IsDetected bool           `json:"isDetected"`      // Whether a package manager was detected
	Error      string         `json:"error,omitempty"` // Error message if detection failed
}

// Detector handles package manager detection.
type Detector struct {
	projectPath string
}

// NewDetector creates a new package manager detector.
func NewDetector(projectPath string) *Detector {
	return &Detector{
		projectPath: projectPath,
	}
}

// Detect detects the package manager used in the project.
// Priority order: bun > pnpm > yarn > npm (newer/faster tools preferred)
func (d *Detector) Detect() *ManagerInfo {
	if d.projectPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return &ManagerInfo{
				Manager:    Unknown,
				IsDetected: false,
				Error:      fmt.Sprintf("could not determine current directory: %v", err),
			}
		}
		d.projectPath = wd
	}

	// Check for lock files in priority order: bun > pnpm > yarn > npm
	lockFiles := []struct {
		filename string
		manager  PackageManager
		install  string
	}{
		{"bun.lockb", Bun, "bun install"},
		{"pnpm-lock.yaml", PNPM, "pnpm install"},
		{"yarn.lock", Yarn, "yarn install"},
		{"package-lock.json", NPM, "npm install"},
	}

	for _, lf := range lockFiles {
		lockPath := filepath.Join(d.projectPath, lf.filename)
		if _, err := os.Stat(lockPath); err == nil {
			return &ManagerInfo{
				Manager:    lf.manager,
				InstallCmd: lf.install,
				LockFile:   lockPath,
				IsDetected: true,
			}
		}
	}

	// No lock file found - check for package.json to default to npm
	packageJSONPath := filepath.Join(d.projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err == nil {
		return &ManagerInfo{
			Manager:    NPM,
			InstallCmd: "npm install",
			LockFile:   "",
			IsDetected: true,
			Error:      "no lock file found, defaulting to npm",
		}
	}

	return &ManagerInfo{
		Manager:    Unknown,
		IsDetected: false,
		Error:      "no package.json or lock files found",
	}
}

// String returns the string representation of the package manager.
func (pm PackageManager) String() string {
	return string(pm)
}

// GetInstallCommand returns the install command for a package manager.
func (pm PackageManager) GetInstallCommand() string {
	switch pm {
	case NPM:
		return "npm install"
	case Yarn:
		return "yarn install"
	case PNPM:
		return "pnpm install"
	case Bun:
		return "bun install"
	default:
		return "npm install"
	}
}

// GetCleanCommand returns the cache clean command for a package manager.
func (pm PackageManager) GetCleanCommand() string {
	switch pm {
	case NPM:
		return "npm cache clean --force"
	case Yarn:
		return "yarn cache clean"
	case PNPM:
		return "pnpm store prune"
	case Bun:
		return "bun pm cache rm"
	default:
		return "npm cache clean --force"
	}
}
