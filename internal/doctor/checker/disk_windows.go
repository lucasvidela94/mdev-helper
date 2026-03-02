//go:build windows
// +build windows

package checker

import (
	"os"

	"golang.org/x/sys/windows"
)

// getDiskSpace returns free and total space for the given path.
// Windows implementation using GetDiskFreeSpaceEx.
func getDiskSpace(path string) (freeBytes, totalBytes uint64, err error) {
	// Ensure path exists
	if _, err := os.Stat(path); err != nil {
		return 0, 0, err
	}

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64

	err = windows.GetDiskFreeSpaceEx(
		windows.StringToUTF16Ptr(path),
		&freeBytesAvailable,
		&totalNumberOfBytes,
		&totalNumberOfFreeBytes,
	)
	if err != nil {
		return 0, 0, err
	}

	return freeBytesAvailable, totalNumberOfBytes, nil
}
