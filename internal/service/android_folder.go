// Package service provides Android folder cleaning functionality.
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/expo"
)

// AndroidFolderService handles cleaning the android/ folder with safety checks.
type AndroidFolderService struct {
	projectPath string
}

// AndroidFolderResult represents the result of an android folder operation.
type AndroidFolderResult struct {
	Success       bool     `json:"success"`
	Action        string   `json:"action"`
	SizeFreed     string   `json:"sizeFreed,omitempty"`
	Warnings      []string `json:"warnings,omitempty"`
	HasGitChanges bool     `json:"hasGitChanges"`
	IsExpoManaged bool     `json:"isExpoManaged"`
	Error         string   `json:"error,omitempty"`
}

// NewAndroidFolderService creates a new AndroidFolderService.
func NewAndroidFolderService(projectPath string) *AndroidFolderService {
	return &AndroidFolderService{
		projectPath: projectPath,
	}
}

// Clean removes the android/ folder with safety checks.
func (s *AndroidFolderService) Clean(dryRun bool) (*AndroidFolderResult, error) {
	return s.cleanInternal(dryRun, false)
}

// CleanWithReinstall removes the android/ folder and regenerates it (for Expo projects).
func (s *AndroidFolderService) CleanWithReinstall(dryRun bool) (*AndroidFolderResult, error) {
	return s.cleanInternal(dryRun, true)
}

// cleanInternal performs the actual cleaning with optional regeneration.
func (s *AndroidFolderService) cleanInternal(dryRun, reinstall bool) (*AndroidFolderResult, error) {
	if s.projectPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("could not determine current directory: %w", err)
		}
		s.projectPath = wd
	}

	result := &AndroidFolderResult{
		Success:  true,
		Warnings: []string{},
	}

	// Check if android directory exists
	androidPath := filepath.Join(s.projectPath, "android")
	info, err := os.Stat(androidPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Success = false
			result.Error = "android/ directory does not exist"
			return result, nil
		}
		return nil, fmt.Errorf("could not stat android directory: %w", err)
	}

	if !info.IsDir() {
		result.Success = false
		result.Error = "android/ exists but is not a directory"
		return result, nil
	}

	// Detect project type and workflow
	expoInfo := expo.Detect(s.projectPath)
	result.IsExpoManaged = expoInfo.HasExpo && expoInfo.IsManaged

	// Check for uncommitted git changes in android/
	result.HasGitChanges = s.hasGitChanges(androidPath)
	if result.HasGitChanges {
		result.Warnings = append(result.Warnings, "android/ folder has uncommitted git changes. Consider committing or stashing them first.")
	}

	// Add workflow-specific warnings
	if expoInfo.HasExpo {
		if expoInfo.IsManaged {
			result.Warnings = append(result.Warnings, "This appears to be an Expo Managed workflow project. The android/ folder is auto-generated. You can safely delete it and run 'npx expo prebuild' to regenerate.")
		} else {
			result.Warnings = append(result.Warnings, "This appears to be an Expo Bare workflow project. The android/ folder contains native code modifications. Make sure you have committed any custom changes before deleting.")
		}
	} else {
		result.Warnings = append(result.Warnings, "This appears to be a React Native Bare project. The android/ folder contains native code. Make sure you have committed any custom changes before deleting.")
	}

	// Calculate size
	size := getDirSize(androidPath)
	result.SizeFreed = formatSize(size)

	if dryRun {
		result.Action = "dry-run: would remove android/ folder"
		return result, nil
	}

	// Remove android directory
	result.Action = "remove android/ folder"
	if err := os.RemoveAll(androidPath); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to remove android/ folder: %v", err)
		return result, nil
	}

	// Regenerate if requested (for Expo projects)
	if reinstall && expoInfo.HasExpo {
		result.Action = "remove and regenerate android/ folder"

		cmd := exec.Command("npx", "expo", "prebuild", "--platform", "android")
		cmd.Dir = s.projectPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to regenerate android/ folder: %v. You can manually run 'npx expo prebuild' later.", err))
		}
	}

	return result, nil
}

// hasGitChanges checks if the android directory has uncommitted git changes.
func (s *AndroidFolderService) hasGitChanges(androidPath string) bool {
	// Check if we're in a git repository
	gitCmd := exec.Command("git", "-C", s.projectPath, "rev-parse", "--git-dir")
	if err := gitCmd.Run(); err != nil {
		// Not a git repository
		return false
	}

	// Check for uncommitted changes in android/ directory
	cmd := exec.Command("git", "-C", s.projectPath, "status", "--porcelain", "android/")
	output, err := cmd.Output()
	if err != nil {
		// If git status fails, assume no changes to be safe
		return false
	}

	// If output is not empty, there are uncommitted changes
	return strings.TrimSpace(string(output)) != ""
}

// GetInfo returns information about the android folder without modifying it.
func (s *AndroidFolderService) GetInfo() (*AndroidFolderResult, error) {
	if s.projectPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("could not determine current directory: %w", err)
		}
		s.projectPath = wd
	}

	result := &AndroidFolderResult{
		Success:  true,
		Warnings: []string{},
	}

	// Check if android directory exists
	androidPath := filepath.Join(s.projectPath, "android")
	info, err := os.Stat(androidPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Success = false
			result.Error = "android/ directory does not exist"
			return result, nil
		}
		return nil, fmt.Errorf("could not stat android directory: %w", err)
	}

	if !info.IsDir() {
		result.Success = false
		result.Error = "android/ exists but is not a directory"
		return result, nil
	}

	// Detect project type
	expoInfo := expo.Detect(s.projectPath)
	result.IsExpoManaged = expoInfo.HasExpo && expoInfo.IsManaged
	result.HasGitChanges = s.hasGitChanges(androidPath)

	// Calculate size
	size := getDirSize(androidPath)
	result.SizeFreed = formatSize(size)

	return result, nil
}
