// Package updater provides self-update functionality for the mdev CLI.
package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	// GitHubAPIBase is the base URL for GitHub API.
	GitHubAPIBase = "https://api.github.com"

	// GitHubRepoOwner is the owner of the repository.
	GitHubRepoOwner = "sombi"

	// GitHubRepoName is the name of the repository.
	GitHubRepoName = "mobile-dev-helper"

	// RequestTimeout is the timeout for HTTP requests.
	RequestTimeout = 30 * time.Second
)

// GitHubRelease represents a GitHub release response.
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	Assets      []Asset   `json:"assets"`
	TarballURL  string    `json:"tarball_url"`
	ZipballURL  string    `json:"zipball_url"`
}

// Asset represents a release asset.
type Asset struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Size               int64     `json:"size"`
	DownloadCount      int       `json:"download_count"`
	BrowserDownloadURL string    `json:"browser_download_url"`
	ContentType        string    `json:"content_type"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// GitHubClient provides methods to interact with the GitHub API.
type GitHubClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		baseURL: GitHubAPIBase,
		httpClient: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetLatestRelease fetches the latest release from GitHub.
// Returns the release information or an error if the request fails.
func (c *GitHubClient) GetLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.baseURL, GitHubRepoOwner, GitHubRepoName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for GitHub API
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "mdev-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("GitHub API rate limit exceeded. Please try again later")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release response: %w", err)
	}

	return &release, nil
}

// FindAssetForPlatform finds the appropriate binary asset for the current platform.
// It matches assets based on GOOS and GOARCH.
func (r *GitHubRelease) FindAssetForPlatform() *Asset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Build expected asset name patterns
	// e.g., mdev-linux-amd64, mdev-darwin-arm64, mdev-windows-amd64.exe
	patterns := []string{
		fmt.Sprintf("mdev-%s-%s", goos, goarch),
		fmt.Sprintf("mdev-%s-%s.exe", goos, goarch),
	}

	for _, asset := range r.Assets {
		for _, pattern := range patterns {
			if strings.Contains(asset.Name, pattern) {
				return &asset
			}
		}
	}

	return nil
}

// ReleaseInfo contains processed information about a release.
type ReleaseInfo struct {
	Version      string
	Name         string
	Changelog    string
	URL          string
	PublishedAt  time.Time
	IsPrerelease bool
	Asset        *Asset
}

// ToReleaseInfo converts a GitHubRelease to ReleaseInfo.
func (r *GitHubRelease) ToReleaseInfo() *ReleaseInfo {
	asset := r.FindAssetForPlatform()

	return &ReleaseInfo{
		Version:      r.TagName,
		Name:         r.Name,
		Changelog:    r.Body,
		URL:          r.HTMLURL,
		PublishedAt:  r.PublishedAt,
		IsPrerelease: r.Prerelease,
		Asset:        asset,
	}
}
