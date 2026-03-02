// Package sdk provides Android SDK detection functionality.
package sdk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// Standard paths for Android SDK by OS
var standardPaths = func() []string {
	var paths []string
	home := os.Getenv("HOME")

	switch runtime.GOOS {
	case "linux":
		paths = []string{
			filepath.Join(home, "Android", "Sdk"),
			"/opt/android-sdk",
			"/usr/local/android-sdk",
		}
	case "darwin":
		paths = []string{
			filepath.Join(home, "Library", "Android", "sdk"),
			"/opt/android-sdk",
		}
	case "windows":
		paths = []string{
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Android", "Sdk"),
			filepath.Join(os.Getenv("PROGRAMFILES"), "Android", "Sdk"),
			"C:\\Android\\Sdk",
		}
	}
	return paths
}()

// Detect searches for Android SDK on the system.
func Detect() *detector.SDKInfo {
	// Priority 1: Check ANDROID_HOME
	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome != "" {
		info := checkSDKPath(androidHome)
		if info != nil {
			return info
		}
	}

	// Priority 2: Check ANDROID_SDK_ROOT
	androidSDKRoot := os.Getenv("ANDROID_SDK_ROOT")
	if androidSDKRoot != "" && androidSDKRoot != androidHome {
		info := checkSDKPath(androidSDKRoot)
		if info != nil {
			return info
		}
	}

	// Priority 3: Check standard directories
	for _, path := range standardPaths {
		info := checkSDKPath(path)
		if info != nil {
			return info
		}
	}

	return &detector.SDKInfo{
		IsValid: false,
		Error:   "No Android SDK found. Install Android SDK or set ANDROID_HOME",
	}
}

// checkSDKPath checks if a given path contains a valid Android SDK.
func checkSDKPath(path string) *detector.SDKInfo {
	// Check if path exists
	if _, err := os.Stat(path); err != nil {
		return nil
	}

	info := &detector.SDKInfo{
		Path:    path,
		IsValid: true,
	}

	// Get SDK version (from platforms if available)
	if platforms, err := getPlatforms(path); err == nil {
		info.Platforms = platforms
		if len(platforms) > 0 {
			info.Version = platforms[0]
		}
	}

	// Get build tools
	if buildTools, err := getBuildTools(path); err == nil {
		info.BuildTools = buildTools
	}

	// Get NDK if available
	if ndk, err := getNDKVersion(path); err == nil {
		info.NDK = ndk
	}

	// Get command line tools
	if cmdlineTools, err := getCommandLineToolsVersion(path); err == nil {
		info.CommandLineTools = cmdlineTools
	}

	// Validate that we have at least platforms or build-tools
	if len(info.Platforms) == 0 && len(info.BuildTools) == 0 {
		return nil
	}

	return info
}

// getPlatforms returns list of installed Android platform versions.
func getPlatforms(sdkPath string) ([]string, error) {
	platformsPath := filepath.Join(sdkPath, "platforms")
	entries, err := os.ReadDir(platformsPath)
	if err != nil {
		return nil, err
	}

	var platforms []string
	for _, entry := range entries {
		if entry.IsDir() {
			platforms = append(platforms, entry.Name())
		}
	}
	return platforms, nil
}

// getBuildTools returns list of installed build tools versions.
func getBuildTools(sdkPath string) ([]string, error) {
	buildToolsPath := filepath.Join(sdkPath, "build-tools")
	entries, err := os.ReadDir(buildToolsPath)
	if err != nil {
		return nil, err
	}

	var tools []string
	for _, entry := range entries {
		if entry.IsDir() {
			tools = append(tools, entry.Name())
		}
	}
	return tools, nil
}

// getNDKVersion returns the installed NDK version.
func getNDKVersion(sdkPath string) (string, error) {
	ndkPath := filepath.Join(sdkPath, "ndk")
	entries, err := os.ReadDir(ndkPath)
	if err != nil {
		return "", err
	}

	// Return the first NDK version found (most recent)
	for _, entry := range entries {
		if entry.IsDir() {
			return entry.Name(), nil
		}
	}
	return "", fmt.Errorf("no NDK found")
}

// getCommandLineToolsVersion returns the command line tools version.
func getCommandLineToolsVersion(sdkPath string) (string, error) {
	// Check for cmdline-tools directory
	cmdlinePath := filepath.Join(sdkPath, "cmdline-tools")
	entries, err := os.ReadDir(cmdlinePath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "latest") {
			return entry.Name(), nil
		}
	}

	// Check tools/bin for sdkmanager
	toolsPath := filepath.Join(sdkPath, "tools", "bin", "sdkmanager")
	if runtime.GOOS == "windows" {
		toolsPath += ".bat"
	}

	if _, err := os.Stat(toolsPath); err == nil {
		// Try to get version
		cmd := exec.Command(toolsPath, "--version")
		output, err := cmd.Output()
		if err == nil {
			version := strings.TrimSpace(string(output))
			if version != "" {
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("command line tools not found")
}
