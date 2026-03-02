// Package version provides version information for the CLI.
// The version is intended to be set at build time via ldflags.
// Example: go build -ldflags "-X main.version=1.0.0"
package version

// Version is the current version of the CLI.
// Defaults to "dev" if not set at build time.
var Version = "dev"
