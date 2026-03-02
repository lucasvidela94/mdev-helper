// Package config provides version information for the CLI.
// These variables are set at build time via ldflags.
package config

// Version information variables.
// These are set at build time via ldflags in the Makefile.
var (
	// Version is the current version of the CLI (e.g., "v1.2.3").
	// Defaults to "dev" if not set at build time.
	Version = "dev"

	// BuildTime is the timestamp when the binary was built (RFC3339 format).
	// Defaults to "unknown" if not set at build time.
	BuildTime = "unknown"

	// GitCommit is the git commit SHA (short form, 7 characters).
	// Defaults to "unknown" if not set at build time.
	GitCommit = "unknown"
)

// IsDevBuild returns true if the binary was built without version ldflags.
func IsDevBuild() bool {
	return Version == "dev"
}

// VersionInfo returns a formatted version string for display.
func VersionInfo() string {
	return Version
}

// FullVersionInfo returns the complete version information.
func FullVersionInfo() string {
	return Version
}
