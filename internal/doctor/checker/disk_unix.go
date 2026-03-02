//go:build !windows
// +build !windows

package checker

import (
	"os"
	"syscall"
)

// getDiskSpace returns free and total space for the given path.
// Unix implementation using syscall.Statfs.
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
