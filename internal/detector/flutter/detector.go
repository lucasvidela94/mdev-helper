// Package flutter provides Flutter SDK detection functionality.
package flutter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// Standard paths for Flutter installations by OS
var standardPaths = func() []string {
	home := os.Getenv("HOME")

	switch runtime.GOOS {
	case "linux":
		return []string{
			filepath.Join(home, "flutter"),
			"/opt/flutter",
			"/usr/local/flutter",
			"/snap/flutter/current",
		}
	case "darwin":
		return []string{
			filepath.Join(home, "flutter"),
			"/opt/flutter",
			"/usr/local/flutter",
			filepath.Join(home, "development", "flutter"),
		}
	case "windows":
		return []string{
			`C:\flutter`,
			`C:\src\flutter`,
			filepath.Join(os.Getenv("USERPROFILE"), "flutter"),
		}
	}
	return nil
}()

// Detect searches for Flutter SDK installations on the system.
func Detect() *detector.FlutterInfo {
	// Priority 1: Check FLUTTER_HOME
	flutterHome := os.Getenv("FLUTTER_HOME")
	if flutterHome != "" {
		info := checkFlutterPath(flutterHome)
		if info != nil {
			info.FlutterHome = flutterHome
			return info
		}
	}

	// Priority 2: Try to find flutter in PATH
	info := detectFromPath()
	if info != nil && info.IsValid {
		return info
	}

	// Priority 3: Check standard installation paths
	for _, path := range standardPaths {
		info := checkFlutterPath(path)
		if info != nil {
			info.FlutterHome = flutterHome
			if flutterHome == "" {
				info.Warning = "FLUTTER_HOME not set"
			}
			return info
		}
	}

	// Return not found
	return &detector.FlutterInfo{
		IsValid: false,
		Error:   "No Flutter SDK found. Install Flutter or set FLUTTER_HOME",
	}
}

// checkFlutterPath checks if a given path contains a valid Flutter SDK.
func checkFlutterPath(path string) *detector.FlutterInfo {
	flutterBin := filepath.Join(path, "bin", "flutter")
	if runtime.GOOS == "windows" {
		flutterBin += ".bat"
	}

	if _, err := os.Stat(flutterBin); err != nil {
		return nil
	}

	info := &detector.FlutterInfo{
		Path:    path,
		IsValid: true,
	}

	// Get version information
	version, err := getFlutterVersion(flutterBin)
	if err == nil {
		info.Version = version.version
		info.Channel = version.channel
		info.FrameworkRev = version.frameworkRev
		info.EngineRev = version.engineRev
	}

	// Get bundled Dart SDK path
	dartPath := filepath.Join(path, "bin", "cache", "dart-sdk")
	if _, err := os.Stat(dartPath); err == nil {
		info.DartPath = dartPath
	}

	// Run flutter doctor
	doctor, err := runFlutterDoctor(flutterBin)
	if err == nil {
		info.Doctor = doctor
	}

	return info
}

// detectFromPath tries to find flutter in PATH.
func detectFromPath() *detector.FlutterInfo {
	flutterPath, err := exec.LookPath("flutter")
	if err != nil {
		return nil
	}

	// Resolve symlinks to get actual path
	realPath, err := filepath.EvalSymlinks(flutterPath)
	if err != nil {
		realPath = flutterPath
	}

	// Get the Flutter SDK home from the flutter binary path (bin/flutter -> bin/ -> flutter/)
	flutterHome := filepath.Dir(filepath.Dir(realPath))

	info := checkFlutterPath(flutterHome)
	if info != nil {
		info.FlutterHome = os.Getenv("FLUTTER_HOME")
		if info.FlutterHome == "" {
			info.Warning = "FLUTTER_HOME not set"
		}
		return info
	}

	return nil
}

// versionInfo holds parsed version information
type versionInfo struct {
	version      string
	channel      string
	frameworkRev string
	engineRev    string
}

// getFlutterVersion runs `flutter --version` and parses the output.
func getFlutterVersion(flutterBin string) (versionInfo, error) {
	cmd := exec.Command(flutterBin, "--version", "--machine")
	output, err := cmd.Output()
	if err != nil {
		return versionInfo{}, err
	}

	// Parse JSON output
	var result struct {
		FrameworkVersion string `json:"frameworkVersion"`
		Channel          string `json:"channel"`
		FrameworkCommit  string `json:"frameworkCommit"`
		EngineCommit     string `json:"engineCommit"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		// Fallback to text parsing if JSON fails
		return parseVersionText(string(output))
	}

	return versionInfo{
		version:      result.FrameworkVersion,
		channel:      result.Channel,
		frameworkRev: result.FrameworkCommit,
		engineRev:    result.EngineCommit,
	}, nil
}

// parseVersionText parses text output from flutter --version (fallback).
func parseVersionText(output string) (versionInfo, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	info := versionInfo{}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "Flutter ") {
			// Format: "Flutter 3.24.0 • channel stable • ..."
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info.version = parts[1]
			}
			if len(parts) >= 5 && parts[2] == "•" {
				info.channel = parts[4]
			}
		}
	}

	if info.version == "" {
		return versionInfo{}, fmt.Errorf("could not parse version from output")
	}

	return info, nil
}

// runFlutterDoctor runs `flutter doctor --machine` and parses the output.
func runFlutterDoctor(flutterBin string) (*detector.FlutterDoctor, error) {
	cmd := exec.Command(flutterBin, "doctor", "--machine")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse JSON output
	var categories []struct {
		Name   string `json:"name"`
		Status bool   `json:"status"`
		Issues []struct {
			Description string `json:"description"`
			IsError     bool   `json:"isError"`
		} `json:"issues"`
	}

	if err := json.Unmarshal(output, &categories); err != nil {
		// Fallback: return empty doctor info
		return &detector.FlutterDoctor{
			Categories: []detector.DoctorCategory{},
			Issues:     []string{},
		}, nil
	}

	doctor := &detector.FlutterDoctor{
		Categories: make([]detector.DoctorCategory, 0, len(categories)),
		Issues:     []string{},
	}

	for _, cat := range categories {
		status := "ok"
		if !cat.Status {
			status = "missing"
		}
		if len(cat.Issues) > 0 {
			status = "partial"
		}

		doctor.Categories = append(doctor.Categories, detector.DoctorCategory{
			Name:   cat.Name,
			Status: status,
		})

		for _, issue := range cat.Issues {
			doctor.Issues = append(doctor.Issues, issue.Description)
		}
	}

	return doctor, nil
}

// DetectProject checks if the given path is a Flutter project.
func DetectProject(projectPath string) *detector.FlutterProjectInfo {
	pubspecPath := filepath.Join(projectPath, "pubspec.yaml")

	data, err := os.ReadFile(pubspecPath)
	if err != nil {
		return &detector.FlutterProjectInfo{
			Path:    projectPath,
			IsValid: false,
			Error:   "No pubspec.yaml found",
		}
	}

	content := string(data)

	// Check if it has flutter dependency
	if !strings.Contains(content, "flutter:") {
		return &detector.FlutterProjectInfo{
			Path:    projectPath,
			IsValid: false,
			Error:   "Not a Flutter project (no flutter dependency in pubspec.yaml)",
		}
	}

	info := &detector.FlutterProjectInfo{
		Path:         projectPath,
		IsValid:      true,
		Dependencies: []string{},
	}

	// Parse basic info from pubspec.yaml
	lines := strings.Split(content, "\n")
	inDependencies := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse name
		if strings.HasPrefix(line, "name:") {
			info.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}

		// Parse version
		if strings.HasPrefix(line, "version:") {
			info.Version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		}

		// Check for Flutter SDK constraint
		if strings.Contains(line, "sdk: flutter") {
			info.FlutterSDK = "any"
		}

		// Parse dependencies section
		if line == "dependencies:" {
			inDependencies = true
			continue
		}

		if inDependencies {
			// End of dependencies section (new top-level key or empty line with no indent)
			if line == "" || (!strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "-")) {
				if strings.HasSuffix(line, ":") && !strings.Contains(line, ": ") {
					inDependencies = false
					continue
				}
			}

			// Extract dependency name
			if line != "" && !strings.HasPrefix(line, "#") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 0 {
					depName := strings.TrimSpace(parts[0])
					if depName != "" && depName != "flutter" && depName != "sdk" {
						info.Dependencies = append(info.Dependencies, depName)
					}
				}
			}
		}
	}

	return info
}
