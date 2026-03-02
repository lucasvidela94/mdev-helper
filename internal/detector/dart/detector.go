// Package dart provides Dart SDK detection functionality.
package dart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// Standard paths for Dart SDK installations by OS
var standardPaths = func() []string {
	home := os.Getenv("HOME")

	switch runtime.GOOS {
	case "linux":
		return []string{
			"/usr/lib/dart",
			"/opt/dart",
			"/usr/local/dart",
			filepath.Join(home, "dart-sdk"),
		}
	case "darwin":
		return []string{
			"/usr/local/dart",
			"/opt/dart",
			filepath.Join(home, "dart-sdk"),
			filepath.Join(home, "development", "dart-sdk"),
		}
	case "windows":
		return []string{
			`C:\dart-sdk`,
			filepath.Join(os.Getenv("LOCALAPPDATA"), "dart-sdk"),
		}
	}
	return nil
}()

// Detect searches for Dart SDK installations on the system.
// If flutterPath is provided, it will check for bundled Dart first.
func Detect(flutterPath string) *detector.DartInfo {
	// Priority 1: Check for Dart bundled with Flutter
	if flutterPath != "" {
		bundledDart := filepath.Join(flutterPath, "bin", "cache", "dart-sdk")
		if info := checkDartPath(bundledDart); info != nil {
			info.IsBundled = true
			info.Flutter = flutterPath
			return info
		}
	}

	// Priority 2: Check DART_SDK environment variable
	dartSDK := os.Getenv("DART_SDK")
	if dartSDK != "" {
		if info := checkDartPath(dartSDK); info != nil {
			return info
		}
	}

	// Priority 3: Try to find dart in PATH
	info := detectFromPath()
	if info != nil && info.IsValid {
		return info
	}

	// Priority 4: Check standard installation paths
	for _, path := range standardPaths {
		if info := checkDartPath(path); info != nil {
			return info
		}
	}

	// Return not found
	return &detector.DartInfo{
		IsValid: false,
		Error:   "No Dart SDK found. Install Dart or use Flutter (which includes Dart)",
	}
}

// checkDartPath checks if a given path contains a valid Dart SDK.
func checkDartPath(path string) *detector.DartInfo {
	dartBin := filepath.Join(path, "bin", "dart")
	if runtime.GOOS == "windows" {
		dartBin += ".exe"
	}

	if _, err := os.Stat(dartBin); err != nil {
		return nil
	}

	info := &detector.DartInfo{
		Path:    path,
		IsValid: true,
	}

	// Get version
	version, err := getDartVersion(dartBin)
	if err == nil {
		info.Version = version
	}

	return info
}

// detectFromPath tries to find dart in PATH.
func detectFromPath() *detector.DartInfo {
	dartPath, err := exec.LookPath("dart")
	if err != nil {
		return nil
	}

	// Resolve symlinks to get actual path
	realPath, err := filepath.EvalSymlinks(dartPath)
	if err != nil {
		realPath = dartPath
	}

	// Get the Dart SDK home from the dart binary path (bin/dart -> bin/ -> dart-sdk/)
	dartHome := filepath.Dir(filepath.Dir(realPath))

	info := checkDartPath(dartHome)
	if info != nil {
		return info
	}

	// Fallback: just return path info with version from PATH dart
	version, _ := getDartVersion(dartPath)
	return &detector.DartInfo{
		Path:    filepath.Dir(realPath),
		Version: version,
		IsValid: true,
	}
}

// getDartVersion runs `dart --version` and parses the output.
func getDartVersion(dartBin string) (string, error) {
	cmd := exec.Command(dartBin, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	versionStr := strings.TrimSpace(string(output))

	// Dart version output format: "Dart SDK version: 3.5.0 (stable) ..."
	// or just "3.5.0" in newer versions

	if strings.HasPrefix(versionStr, "Dart SDK version: ") {
		parts := strings.Fields(versionStr)
		if len(parts) >= 4 {
			return parts[3], nil
		}
	}

	// Try to extract version number from any format
	fields := strings.Fields(versionStr)
	for _, field := range fields {
		// Look for something that looks like a version number
		if strings.Contains(field, ".") && !strings.Contains(field, ":") {
			// Clean up any trailing punctuation
			field = strings.TrimRight(field, "(),")
			return field, nil
		}
	}

	return "", fmt.Errorf("could not parse version from: %s", versionStr)
}
