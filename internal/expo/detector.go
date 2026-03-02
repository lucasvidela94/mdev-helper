// Package expo provides Expo project detection functionality.
package expo

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// Detect checks if the current directory is an Expo project and returns info.
func Detect(projectPath string) *detector.ExpoInfo {
	// Use current directory if no path specified
	if projectPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return &detector.ExpoInfo{
				ProjectPath: "",
				Error:       "Could not determine current directory",
			}
		}
		projectPath = wd
	}

	info := &detector.ExpoInfo{
		ProjectPath: projectPath,
	}

	// Check for app.json or expo.json
	appJSONPath := findExpoConfig(projectPath)
	if appJSONPath == "" {
		info.HasExpo = false
		return info
	}

	info.HasExpo = true

	// Read and parse the config file
	config, err := readExpoConfig(appJSONPath)
	if err != nil {
		info.HasExpo = false
		return info
	}

	// Determine workflow (Managed vs Bare)
	info.IsManaged = config.IsManaged()
	info.IsBare = config.IsBare()

	// Get Expo SDK version
	info.ExpoVersion = config.GetExpoVersion()

	return info
}

// findExpoConfig searches for expo config files.
func findExpoConfig(dir string) string {
	configFiles := []string{
		"app.json",
		"app.config.js",
		"app.config.ts",
		"expo.json",
		"expo.config.js",
		"expo.config.ts",
	}

	for _, configFile := range configFiles {
		path := filepath.Join(dir, configFile)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// expoConfig represents the parsed expo/app config.
type expoConfig struct {
	Expo    *expoSection   `json:"expo"`
	Name    string         `json:"name"`
	Slug    string         `json:"slug"`
	IOS     iosSection     `json:"ios,omitempty"`
	Android androidSection `json:"android,omitempty"`
}

// expoSection contains the expo-specific config.
type expoSection struct {
	Name       string `json:"name,omitempty"`
	Slug       string `json:"slug,omitempty"`
	SDKVersion string `json:""sdkVersion,omitempty"`
	Version    string `json:"version,omitempty"`
}

// iosSection contains iOS specific config.
type iosSection struct {
	BundleIdentifier string `json:"bundleIdentifier,omitempty"`
}

// androidSection contains Android specific config.
type androidSection struct {
	Package string `json:"package,omitempty"`
}

// readExpoConfig reads and parses an Expo config file.
func readExpoConfig(path string) (*expoConfig, error) {
	// Handle JS/TS config files - just check if they exist
	if filepath.Ext(path) == ".js" || filepath.Ext(path) == ".ts" {
		// For config files that are code, we can't easily parse them
		// Return minimal config to indicate Expo presence
		return &expoConfig{
			Expo: &expoSection{},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config expoConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// IsManaged returns true if this is an Expo Managed workflow project.
func (c *expoConfig) IsManaged() bool {
	// If there's no ios/android directories with native code, it's managed
	// Also check if expo.ios or expo.android keys are missing or minimal
	if c.Expo == nil {
		return false
	}

	// If there's no explicit bare field, default to managed
	return true
}

// IsBare returns true if this is an Expo Bare workflow project.
func (c *expoConfig) IsBare() bool {
	// Bare workflow has explicit ios/android directories
	return !c.IsManaged()
}

// GetExpoVersion returns the Expo SDK version.
func (c *expoConfig) GetExpoVersion() string {
	if c.Expo == nil {
		return ""
	}
	return c.Expo.SDKVersion
}

// HasNativeDirectories checks if the project has ios/android directories.
func HasNativeDirectories(projectPath string) (hasIOS, hasAndroid bool) {
	iosPath := filepath.Join(projectPath, "ios")
	androidPath := filepath.Join(projectPath, "android")

	if _, err := os.Stat(iosPath); err == nil {
		hasIOS = true
	}
	if _, err := os.Stat(androidPath); err == nil {
		hasAndroid = true
	}

	return
}

// DetectWorkflow determines if the project is Managed or Bare workflow.
func DetectWorkflow(projectPath string) (bool, bool) {
	hasIOS, hasAndroid := HasNativeDirectories(projectPath)

	// If native directories exist, it's Bare workflow
	// If no native directories, it's Managed workflow
	isBare := hasIOS || hasAndroid
	isManaged := !isBare

	return isManaged, isBare
}

// FindProjectRoot searches upward for Expo project root.
func FindProjectRoot(startPath string) string {
	dir := startPath

	for {
		if findExpoConfig(dir) != "" {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}
