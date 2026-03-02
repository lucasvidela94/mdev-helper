// Package gradlew provides Gradle wrapper detection functionality.
package gradlew

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// Detect searches for gradlew in the current project or specified path.
func Detect(projectPath string) *detector.GradleInfo {
	// Use current directory if no path specified
	if projectPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return &detector.GradleInfo{
				IsValid: false,
				Error:   "Could not determine current directory",
			}
		}
		projectPath = wd
	}

	// Look for gradlew in the project path
	gradlewPath := findGradlew(projectPath)
	if gradlewPath == "" {
		return &detector.GradleInfo{
			Path:       projectPath,
			IsValid:    false,
			HasWrapper: false,
			Error:      "No gradlew found in project directory",
		}
	}

	// Check if gradlew is executable or needs repair
	info := &detector.GradleInfo{
		Path:       gradlewPath,
		HasWrapper: true,
	}

	// Try to get Gradle version from wrapper properties
	gradleVersion := getGradleVersionFromWrapper(projectPath)
	if gradleVersion != "" {
		info.GradleVersion = gradleVersion
	}

	// Test if gradlew is executable
	if err := testGradlew(gradlewPath); err != nil {
		info.IsValid = false
		info.Error = "gradlew is not executable or broken: " + err.Error()
		return info
	}

	info.IsValid = true
	info.JavaHome = os.Getenv("JAVA_HOME")

	return info
}

// findGradlew searches for gradlew in the given directory.
func findGradlew(dir string) string {
	// Check for gradlew in the root
	gradlewPath := filepath.Join(dir, "gradlew")
	if _, err := os.Stat(gradlewPath); err == nil {
		return gradlewPath
	}

	// Also check common subdirectories
	for _, subdir := range []string{"android", "android/app", "app"} {
		gradlewPath = filepath.Join(dir, subdir, "gradlew")
		if _, err := os.Stat(gradlewPath); err == nil {
			return gradlewPath
		}
	}

	return ""
}

// getGradleVersionFromWrapper reads gradle-wrapper.properties to get the Gradle version.
func getGradleVersionFromWrapper(projectPath string) string {
	// Look for wrapper properties in standard locations
	locations := []string{
		filepath.Join(projectPath, "gradle", "wrapper", "gradle-wrapper.properties"),
		filepath.Join(projectPath, "gradle-wrapper.properties"),
	}

	for _, propsPath := range locations {
		data, err := os.ReadFile(propsPath)
		if err != nil {
			continue
		}

		// Parse distributionUrl
		// Example: distributionUrl=https\://services.gradle.org/distributions/gradle-8.2-all.zip
		re := regexp.MustCompile(`distributionUrl=.*?gradle-([0-9.]+)`)
		matches := re.FindStringSubmatch(string(data))
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// testGradlew tests if gradlew can be executed.
func testGradlew(gradlewPath string) error {
	// Make it executable if on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(gradlewPath, 0755); err != nil {
			return fmt.Errorf("failed to make gradlew executable: %w", err)
		}
	}

	// Try running --version
	cmd := exec.Command(gradlewPath, "--version")
	cmd.Dir = filepath.Dir(gradlewPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If it fails, check if it's a permissions issue
		if strings.Contains(string(output), "Permission denied") {
			return fmt.Errorf("permission denied")
		}
		return fmt.Errorf("gradlew execution failed: %w", err)
	}

	return nil
}

// DetectFromWrapperScripts searches for gradlew scripts in parent directories.
func DetectFromWrapperScripts(startPath string) *detector.GradleInfo {
	dir := startPath

	for {
		gradlewPath := findGradlew(dir)
		if gradlewPath != "" {
			return Detect(dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return &detector.GradleInfo{
		IsValid:    false,
		HasWrapper: false,
		Error:      "No gradlew found in directory tree",
	}
}
