// Package updater provides self-update functionality for the mdev CLI.
package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/inconshreveable/go-update"
)

// ProgressCallback is called during download to report progress.
type ProgressCallback func(downloaded, total int64, speed float64)

// DownloadOptions contains options for downloading.
type DownloadOptions struct {
	// Asset is the release asset to download.
	Asset *Asset

	// ProgressCallback is called periodically during download.
	ProgressCallback ProgressCallback

	// UpdateInterval is how often to call the progress callback.
	UpdateInterval time.Duration
}

// DownloadResult contains the result of a download.
type DownloadResult struct {
	// Path is the path to the downloaded file.
	Path string

	// Size is the size of the downloaded file in bytes.
	Size int64

	// Checksum is the SHA256 checksum of the downloaded file.
	Checksum string

	// Duration is how long the download took.
	Duration time.Duration
}

// Downloader handles downloading release assets.
type Downloader struct {
	httpClient *http.Client
}

// NewDownloader creates a new Downloader.
func NewDownloader() *Downloader {
	return &Downloader{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Longer timeout for binary downloads
		},
	}
}

// Download downloads a release asset to a temporary location.
// Returns the path to the downloaded file.
func (d *Downloader) Download(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	if opts.Asset == nil {
		return nil, fmt.Errorf("no asset specified for download")
	}

	if opts.UpdateInterval == 0 {
		opts.UpdateInterval = 100 * time.Millisecond
	}

	// Create temporary file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("mdev-update-%d", time.Now().Unix()))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.Asset.BrowserDownloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("User-Agent", "mdev-cli/1.0")

	startTime := time.Now()

	// Perform download
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create output file
	out, err := os.Create(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	// Get content length for progress tracking
	totalSize := resp.ContentLength

	// Copy with progress tracking
	var downloaded int64
	lastUpdate := time.Now()

	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				os.Remove(tempFile)
				return nil, fmt.Errorf("failed to write download: %w", writeErr)
			}
			downloaded += int64(n)

			// Report progress
			if opts.ProgressCallback != nil && time.Since(lastUpdate) >= opts.UpdateInterval {
				elapsed := time.Since(startTime).Seconds()
				speed := float64(0)
				if elapsed > 0 {
					speed = float64(downloaded) / elapsed
				}
				opts.ProgressCallback(downloaded, totalSize, speed)
				lastUpdate = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tempFile)
			return nil, fmt.Errorf("failed to read download: %w", err)
		}
	}

	duration := time.Since(startTime)

	// Final progress callback
	if opts.ProgressCallback != nil {
		opts.ProgressCallback(downloaded, totalSize, float64(downloaded)/duration.Seconds())
	}

	return &DownloadResult{
		Path:     tempFile,
		Size:     downloaded,
		Duration: duration,
	}, nil
}

// VerifyBinary performs basic verification of the downloaded binary.
// It checks that the file exists and has a reasonable size.
func (d *Downloader) VerifyBinary(path string, minSize int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("downloaded file is a directory")
	}

	if info.Size() < minSize {
		return fmt.Errorf("downloaded file is too small (%d bytes), expected at least %d bytes", info.Size(), minSize)
	}

	return nil
}

// ApplyUpdate applies the downloaded binary using go-update.
// This performs an atomic replacement of the current binary.
func (d *Downloader) ApplyUpdate(binaryPath string) error {
	// Open the new binary
	newBinary, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to open new binary: %w", err)
	}
	defer newBinary.Close()

	// Apply the update atomically
	err = update.Apply(newBinary, update.Options{})
	if err != nil {
		// Attempt rollback if update fails
		if rollbackErr := update.RollbackError(err); rollbackErr != nil {
			return fmt.Errorf("update failed and rollback failed: %v (original error: %w)", rollbackErr, err)
		}
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

// Cleanup removes the temporary download file.
func (d *Downloader) Cleanup(path string) {
	if path != "" {
		os.Remove(path)
	}
}

// FormatBytes formats a byte count into a human-readable string.
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatSpeed formats a speed in bytes/second into a human-readable string.
func FormatSpeed(bytesPerSec float64) string {
	return FormatBytes(int64(bytesPerSec)) + "/s"
}
