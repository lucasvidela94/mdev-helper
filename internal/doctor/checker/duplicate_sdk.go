package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

// DuplicateSDKChecker checks for multiple Android SDK installations.
type DuplicateSDKChecker struct {
	// AdditionalPaths allows specifying extra paths to check beyond defaults
	AdditionalPaths []string
}

// NewDuplicateSDKChecker creates a new duplicate SDK checker.
func NewDuplicateSDKChecker() *DuplicateSDKChecker {
	return &DuplicateSDKChecker{
		AdditionalPaths: []string{},
	}
}

// Name returns the checker name.
func (d *DuplicateSDKChecker) Name() string {
	return "Duplicate SDK Detection"
}

// Category returns the checker category.
func (d *DuplicateSDKChecker) Category() doctor.CheckCategory {
	return doctor.CategoryEnvironment
}

// Check performs the duplicate SDK detection.
func (d *DuplicateSDKChecker) Check() doctor.CheckResult {
	// Find all potential SDK locations
	locations := d.findSDKLocations()

	// Validate each location
	validSDKs := make([]SDKLocation, 0)
	for _, loc := range locations {
		if isValidSDKPath(loc.Path) {
			validSDKs = append(validSDKs, loc)
		}
	}

	result := doctor.CheckResult{
		Name: d.Name(),
		Details: map[string]interface{}{
			"locationsChecked": len(locations),
			"validSDKs":        validSDKs,
		},
	}

	if len(validSDKs) > 1 {
		result.Status = doctor.StatusWarning
		paths := make([]string, len(validSDKs))
		for i, sdk := range validSDKs {
			paths[i] = sdk.Path
		}
		result.Message = fmt.Sprintf(
			"Multiple Android SDK installations detected (%d): %v. This may cause conflicts.",
			len(validSDKs), paths,
		)
	} else if len(validSDKs) == 1 {
		result.Status = doctor.StatusPassed
		result.Message = fmt.Sprintf("Single Android SDK installation found at: %s", validSDKs[0].Path)
	} else {
		result.Status = doctor.StatusPassed
		result.Message = "No Android SDK installations detected"
	}

	return result
}

// SDKLocation represents a discovered SDK location.
type SDKLocation struct {
	Path   string `json:"path"`
	Source string `json:"source"` // How this location was discovered (env, default, etc.)
}

// findSDKLocations searches for potential SDK locations.
func (d *DuplicateSDKChecker) findSDKLocations() []SDKLocation {
	locations := make([]SDKLocation, 0)
	seen := make(map[string]bool)

	// Check ANDROID_HOME
	if androidHome := os.Getenv("ANDROID_HOME"); androidHome != "" {
		if !seen[androidHome] {
			locations = append(locations, SDKLocation{
				Path:   androidHome,
				Source: "ANDROID_HOME",
			})
			seen[androidHome] = true
		}
	}

	// Check ANDROID_SDK_ROOT (legacy)
	if sdkRoot := os.Getenv("ANDROID_SDK_ROOT"); sdkRoot != "" {
		if !seen[sdkRoot] {
			locations = append(locations, SDKLocation{
				Path:   sdkRoot,
				Source: "ANDROID_SDK_ROOT",
			})
			seen[sdkRoot] = true
		}
	}

	// Check default installation paths by OS
	defaultPaths := getDefaultSDKPaths()
	for _, path := range defaultPaths {
		if !seen[path] {
			locations = append(locations, SDKLocation{
				Path:   path,
				Source: "default",
			})
			seen[path] = true
		}
	}

	// Check additional user-specified paths
	for _, path := range d.AdditionalPaths {
		if !seen[path] {
			locations = append(locations, SDKLocation{
				Path:   path,
				Source: "user",
			})
			seen[path] = true
		}
	}

	// Check common alternative locations
	altPaths := getAlternativeSDKPaths()
	for _, path := range altPaths {
		if !seen[path] {
			locations = append(locations, SDKLocation{
				Path:   path,
				Source: "alternative",
			})
			seen[path] = true
		}
	}

	return locations
}

// getDefaultSDKPaths returns the default SDK installation paths by OS.
func getDefaultSDKPaths() []string {
	home := os.Getenv("HOME")

	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(home, "Library/Android/sdk"),
		}
	case "linux":
		return []string{
			filepath.Join(home, "Android/Sdk"),
			"/usr/lib/android-sdk",
			"/opt/android-sdk",
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		return []string{
			filepath.Join(localAppData, "Android", "Sdk"),
			`C:\Users\%USERNAME%\AppData\Local\Android\Sdk`,
		}
	default:
		return []string{
			filepath.Join(home, "Android/Sdk"),
		}
	}
}

// getAlternativeSDKPaths returns alternative/common SDK locations.
func getAlternativeSDKPaths() []string {
	home := os.Getenv("HOME")

	paths := []string{
		"/opt/android-sdk",
		"/usr/local/android-sdk",
		filepath.Join(home, ".android-sdk"),
		filepath.Join(home, "sdk"),
	}

	// Check for SDKs installed via package managers
	if runtime.GOOS == "linux" {
		paths = append(paths,
			"/usr/share/android-sdk",
			"/var/lib/android-sdk",
		)
	}

	return paths
}

// isValidSDKPath checks if a path contains a valid Android SDK.
func isValidSDKPath(path string) bool {
	// Expand environment variables
	path = os.ExpandEnv(path)

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}

	// Check for required SDK directories
	requiredDirs := []string{"platforms", "build-tools"}
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(path, dir)
		if _, err := os.Stat(dirPath); err != nil {
			return false
		}
	}

	return true
}

// GetSDKPlatforms returns the list of installed platform versions.
func GetSDKPlatforms(sdkPath string) ([]string, error) {
	platformsDir := filepath.Join(sdkPath, "platforms")

	entries, err := os.ReadDir(platformsDir)
	if err != nil {
		return nil, err
	}

	platforms := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "android-") {
			platforms = append(platforms, entry.Name())
		}
	}

	return platforms, nil
}

// GetSDKBuildTools returns the list of installed build-tools versions.
func GetSDKBuildTools(sdkPath string) ([]string, error) {
	buildToolsDir := filepath.Join(sdkPath, "build-tools")

	entries, err := os.ReadDir(buildToolsDir)
	if err != nil {
		return nil, err
	}

	versions := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	return versions, nil
}

// DetectConflicts analyzes SDK locations and returns potential conflicts.
func DetectConflicts(sdks []SDKLocation) []string {
	conflicts := make([]string, 0)

	if len(sdks) <= 1 {
		return conflicts
	}

	// Check for different SDKs with same platforms
	platformSets := make(map[string][]string)
	for _, sdk := range sdks {
		platforms, err := GetSDKPlatforms(sdk.Path)
		if err != nil {
			continue
		}
		for _, platform := range platforms {
			platformSets[platform] = append(platformSets[platform], sdk.Path)
		}
	}

	// Report platforms installed in multiple SDKs
	for platform, paths := range platformSets {
		if len(paths) > 1 {
			conflicts = append(conflicts, fmt.Sprintf(
				"Platform %s installed in multiple SDKs: %v",
				platform, paths,
			))
		}
	}

	return conflicts
}
