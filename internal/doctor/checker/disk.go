package checker

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

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

// getDiskSpace returns free and total space for the given path.
// This implementation works for Unix-like systems.
func getDiskSpace(path string) (freeBytes, totalBytes uint64, err error) {
	// Ensure path exists
	if _, err := os.Stat(path); err != nil {
		return 0, 0, err
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, err
	}

	// Calculate free and total space
	// Blocks available to non-superuser * block size
	freeBytes = stat.Bavail * uint64(stat.Bsize)
	totalBytes = stat.Blocks * uint64(stat.Bsize)

	return freeBytes, totalBytes, nil
}

// GetFreeSpaceGB returns the free space in GB for the given path.
func GetFreeSpaceGB(path string) (float64, error) {
	freeBytes, _, err := getDiskSpace(path)
	if err != nil {
		return 0, err
	}
	return float64(freeBytes) / (1024 * 1024 * 1024), nil
}

// init registers the disk checker on non-Windows systems.
func init() {
	if runtime.GOOS != "windows" {
		// Disk space checker is registered by default on Unix systems
		// Windows would need a different implementation
	}
}
