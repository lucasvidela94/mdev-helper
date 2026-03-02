package checker

import (
	"fmt"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

const (
	// MinFreeSpaceGB is the minimum recommended free space in GB.
	MinFreeSpaceGB = 10
	// MinFreeSpaceBytes is MinFreeSpaceGB in bytes.
	MinFreeSpaceBytes = MinFreeSpaceGB * 1024 * 1024 * 1024
)

// DiskChecker checks for available disk space.
type DiskChecker struct {
	path string
}

// NewDiskChecker creates a new disk checker for the specified path.
// If path is empty, it uses the current working directory.
func NewDiskChecker(path string) *DiskChecker {
	if path == "" {
		path = "."
	}
	return &DiskChecker{path: path}
}

// Name returns the checker name.
func (d *DiskChecker) Name() string {
	return "Disk Space"
}

// Category returns the checker category.
func (d *DiskChecker) Category() doctor.CheckCategory {
	return doctor.CategoryEnvironment
}

// Check performs the disk space check.
func (d *DiskChecker) Check() doctor.CheckResult {
	// Get free space
	freeBytes, totalBytes, err := getDiskSpace(d.path)
	if err != nil {
		return doctor.CheckResult{
			Name:    d.Name(),
			Status:  doctor.StatusError,
			Message: fmt.Sprintf("Failed to check disk space: %v", err),
			Details: map[string]interface{}{
				"path":  d.path,
				"error": err.Error(),
			},
		}
	}

	freeGB := float64(freeBytes) / (1024 * 1024 * 1024)
	totalGB := float64(totalBytes) / (1024 * 1024 * 1024)
	usedGB := totalGB - freeGB
	percentUsed := (usedGB / totalGB) * 100

	result := doctor.CheckResult{
		Name: d.Name(),
		Details: map[string]interface{}{
			"path":        d.path,
			"freeBytes":   freeBytes,
			"totalBytes":  totalBytes,
			"freeGB":      freeGB,
			"totalGB":     totalGB,
			"usedGB":      usedGB,
			"percentUsed": percentUsed,
			"minFreeGB":   MinFreeSpaceGB,
		},
	}

	if freeBytes < MinFreeSpaceBytes {
		result.Status = doctor.StatusWarning
		result.Message = fmt.Sprintf(
			"Low disk space: %.1f GB free (minimum recommended: %d GB)",
			freeGB, MinFreeSpaceGB,
		)
	} else {
		result.Status = doctor.StatusPassed
		result.Message = fmt.Sprintf(
			"Disk space OK: %.1f GB free of %.1f GB (%.1f%% used)",
			freeGB, totalGB, percentUsed,
		)
	}

	return result
}

// GetFreeSpaceGB returns the free space in GB for the given path.
func GetFreeSpaceGB(path string) (float64, error) {
	freeBytes, _, err := getDiskSpace(path)
	if err != nil {
		return 0, err
	}
	return float64(freeBytes) / (1024 * 1024 * 1024), nil
}
