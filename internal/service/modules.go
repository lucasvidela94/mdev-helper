// Package service provides node_modules cleaning functionality.
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sombi/mobile-dev-helper/internal/detector/packagemgr"
)

// ModulesService handles cleaning and reinstalling node_modules.
type ModulesService struct {
	projectPath string
	detector    *packagemgr.Detector
}

// NewModulesService creates a new ModulesService.
func NewModulesService(projectPath string) *ModulesService {
	return &ModulesService{
		projectPath: projectPath,
		detector:    packagemgr.NewDetector(projectPath),
	}
}

// CleanResult represents the result of a modules operation.
type ModulesResult struct {
	Success      bool   `json:"success"`
	Manager      string `json:"manager"`
	Action       string `json:"action"`
	SizeFreed    string `json:"sizeFreed,omitempty"`
	Error        string `json:"error,omitempty"`
	ReinstallCmd string `json:"reinstallCmd,omitempty"`
}

// Clean removes node_modules without reinstalling.
func (s *ModulesService) Clean() (*ModulesResult, error) {
	return s.cleanInternal(false)
}

// CleanAndReinstall removes node_modules and reinstalls dependencies.
func (s *ModulesService) CleanAndReinstall() (*ModulesResult, error) {
	return s.cleanInternal(true)
}

// cleanInternal performs the actual cleaning with optional reinstall.
func (s *ModulesService) cleanInternal(reinstall bool) (*ModulesResult, error) {
	if s.projectPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("could not determine current directory: %w", err)
		}
		s.projectPath = wd
		s.detector = packagemgr.NewDetector(wd)
	}

	// Detect package manager
	managerInfo := s.detector.Detect()
	if !managerInfo.IsDetected {
		return &ModulesResult{
			Success: false,
			Error:   fmt.Sprintf("could not detect package manager: %s", managerInfo.Error),
		}, nil
	}

	// Check if node_modules exists
	nodeModulesPath := filepath.Join(s.projectPath, "node_modules")
	info, err := os.Stat(nodeModulesPath)

	var sizeFreed string
	if err == nil && info.IsDir() {
		// Calculate size before deletion
		size := getDirSize(nodeModulesPath)
		sizeFreed = formatSize(size)
	}

	// Remove node_modules
	if err := os.RemoveAll(nodeModulesPath); err != nil {
		return &ModulesResult{
			Success: false,
			Manager: managerInfo.Manager.String(),
			Action:  "remove node_modules",
			Error:   fmt.Sprintf("failed to remove node_modules: %v", err),
		}, nil
	}

	result := &ModulesResult{
		Success:   true,
		Manager:   managerInfo.Manager.String(),
		Action:    "remove node_modules",
		SizeFreed: sizeFreed,
	}

	// Reinstall if requested
	if reinstall {
		result.Action = "remove and reinstall node_modules"
		result.ReinstallCmd = managerInfo.InstallCmd

		cmd := exec.Command("sh", "-c", managerInfo.InstallCmd)
		cmd.Dir = s.projectPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to reinstall dependencies: %v", err)
			return result, nil
		}
	}

	return result, nil
}

// GetManagerInfo returns information about the detected package manager.
func (s *ModulesService) GetManagerInfo() *packagemgr.ManagerInfo {
	return s.detector.Detect()
}

// CleanPackageManagerCache cleans the package manager's global cache.
func (s *ModulesService) CleanPackageManagerCache() (*ModulesResult, error) {
	managerInfo := s.detector.Detect()
	if !managerInfo.IsDetected {
		return &ModulesResult{
			Success: false,
			Error:   fmt.Sprintf("could not detect package manager: %s", managerInfo.Error),
		}, nil
	}

	cleanCmd := managerInfo.Manager.GetCleanCommand()

	cmd := exec.Command("sh", "-c", cleanCmd)
	cmd.Dir = s.projectPath
	output, err := cmd.CombinedOutput()

	result := &ModulesResult{
		Success: err == nil,
		Manager: managerInfo.Manager.String(),
		Action:  "clean package manager cache",
	}

	if err != nil {
		result.Error = fmt.Sprintf("failed to clean cache: %v (output: %s)", err, string(output))
	}

	return result, nil
}
