// Package service provides repair functionality for broken tools.
package service

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// findGradlewInProject searches for gradlew in the project directory and common subdirectories.
// Returns the full path to gradlew if found, or an empty string if not found.
func findGradlewInProject(projectPath string) string {
	if projectPath == "" {
		var err error
		projectPath, err = os.Getwd()
		if err != nil {
			return ""
		}
	}

	// Try root first
	gradlewPath := filepath.Join(projectPath, "gradlew")
	if _, err := os.Stat(gradlewPath); err == nil {
		return gradlewPath
	}

	// Try searching in subdirectories
	for _, subdir := range []string{"android", "android/app"} {
		path := filepath.Join(projectPath, subdir, "gradlew")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// RepairService handles repair operations for development tools.
type RepairService struct {
	projectPath string
}

// NewRepairService creates a new RepairService.
func NewRepairService(projectPath string) *RepairService {
	return &RepairService{
		projectPath: projectPath,
	}
}

// RepairGradlew attempts to repair a broken gradlew wrapper.
func (s *RepairService) RepairGradlew() *detector.RepairResult {
	result := &detector.RepairResult{
		Tool:   "gradlew",
		Action: "repair",
	}

	// Find gradlew using the helper function
	gradlewPath := findGradlewInProject(s.projectPath)
	if gradlewPath == "" {
		result.Success = false
		result.Error = "gradlew not found in project"
		return result
	}

	// Repair step 1: Make executable (Unix)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(gradlewPath, 0755); err != nil {
			result.Success = false
			result.Error = "Failed to make gradlew executable: " + err.Error()
			return result
		}
		result.Action = "chmod"
		result.Message = "Made gradlew executable"
	}

	// Repair step 2: Download wrapper if missing
	wrapperPropsPath := filepath.Join(filepath.Dir(gradlewPath), "gradle", "wrapper", "gradle-wrapper.properties")
	if _, err := os.Stat(wrapperPropsPath); os.IsNotExist(err) {
		// Try to download wrapper
		if err := s.downloadGradleWrapper(gradlewPath); err != nil {
			result.Success = false
			result.Error = "Failed to download wrapper: " + err.Error()
			return result
		}
		result.Action = "download"
		result.Message = "Downloaded Gradle wrapper"
	}

	// Verify gradlew works
	cmd := exec.Command(gradlewPath, "--version")
	cmd.Dir = filepath.Dir(gradlewPath)
	if err := cmd.Run(); err != nil {
		result.Success = false
		result.Error = "gradlew still broken after repair: " + err.Error()
		return result
	}

	result.Success = true
	result.Message = "gradlew repaired successfully"
	return result
}

// downloadGradleWrapper downloads the Gradle wrapper JAR.
func (s *RepairService) downloadGradleWrapper(gradlewPath string) error {
	// Create wrapper directory
	wrapperDir := filepath.Join(filepath.Dir(gradlewPath), "gradle", "wrapper")
	if err := os.MkdirAll(wrapperDir, 0755); err != nil {
		return err
	}

	// Download gradle-wrapper.jar
	jarURL := "https://raw.githubusercontent.com/gradle/gradle/master/gradle/wrapper/gradle-wrapper.jar"
	jarPath := filepath.Join(wrapperDir, "gradle-wrapper.jar")

	resp, err := http.Get(jarURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: %s", resp.Status)
	}

	// Create file
	out, err := os.Create(jarPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.ReadFrom(resp.Body)
	return err
}

// ValidateGradlew validates if gradlew is working.
func (s *RepairService) ValidateGradlew() (*detector.RepairResult, error) {
	result := &detector.RepairResult{
		Tool:   "gradlew",
		Action: "validate",
	}

	// Find gradlew using the helper function (searches subdirectories too)
	gradlewPath := findGradlewInProject(s.projectPath)
	if gradlewPath == "" {
		result.Success = false
		result.Error = "gradlew not found"
		return result, nil
	}

	// Make executable
	if runtime.GOOS != "windows" {
		if err := os.Chmod(gradlewPath, 0755); err != nil {
			result.Success = false
			result.Error = "Failed to make executable: " + err.Error()
			return result, nil
		}
	}

	// Test execution
	cmd := exec.Command(gradlewPath, "--version")
	cmd.Dir = filepath.Dir(gradlewPath)
	if err := cmd.Run(); err != nil {
		result.Success = false
		result.Error = "gradlew execution failed: " + err.Error()
		return result, nil
	}

	result.Success = true
	result.Message = "gradlew is valid"
	return result, nil
}
