// Package updater provides self-update functionality for the mdev CLI.
package updater

import (
	"context"
	"fmt"

	"github.com/sombi/mobile-dev-helper/internal/config"
)

// Updater handles version checking and self-updates.
type Updater struct {
	githubClient   *GitHubClient
	currentVersion string
}

// NewUpdater creates a new Updater instance.
func NewUpdater() *Updater {
	return &Updater{
		githubClient:   NewGitHubClient(),
		currentVersion: config.Version,
	}
}

// NewUpdaterWithVersion creates a new Updater with a specific version (for testing).
func NewUpdaterWithVersion(version string) *Updater {
	return &Updater{
		githubClient:   NewGitHubClient(),
		currentVersion: version,
	}
}

// CheckResult represents the result of a version check.
type CheckResult struct {
	UpdateAvailable bool
	CurrentVersion  string
	LatestVersion   string
	ReleaseURL      string
	ReleaseInfo     *ReleaseInfo
}

// Check checks if a newer version is available.
// Returns a CheckResult indicating whether an update is available.
func (u *Updater) Check(ctx context.Context) (*CheckResult, error) {
	// Fetch latest release from GitHub
	release, err := u.githubClient.GetLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := release.TagName
	currentVersion := u.currentVersion

	// Check if update is available
	updateAvailable := IsNewerVersion(currentVersion, latestVersion)

	result := &CheckResult{
		UpdateAvailable: updateAvailable,
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
		ReleaseURL:      release.HTMLURL,
		ReleaseInfo:     release.ToReleaseInfo(),
	}

	return result, nil
}

// IsUpToDate checks if the current version is up to date.
// Returns true if no update is available.
func (u *Updater) IsUpToDate(ctx context.Context) (bool, error) {
	result, err := u.Check(ctx)
	if err != nil {
		return false, err
	}
	return !result.UpdateAvailable, nil
}

// GetCurrentVersion returns the current version.
func (u *Updater) GetCurrentVersion() string {
	return u.currentVersion
}
