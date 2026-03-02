// Package updater provides self-update functionality for the mdev CLI.
package updater

import (
	"fmt"

	"golang.org/x/mod/semver"
)

// CompareVersions compares two semantic versions.
// Returns:
//   - 1 if v1 > v2
//   - 0 if v1 == v2
//   - -1 if v1 < v2
//
// Both versions must be valid semantic versions (prefixed with 'v').
// If a version doesn't have the 'v' prefix, it will be added.
func CompareVersions(v1, v2 string) int {
	v1 = normalizeVersion(v1)
	v2 = normalizeVersion(v2)

	return semver.Compare(v1, v2)
}

// IsNewerVersion returns true if candidate is newer than current.
// It handles pre-release versions correctly (won't suggest downgrading
// from a pre-release to a stable version).
func IsNewerVersion(current, candidate string) bool {
	current = normalizeVersion(current)
	candidate = normalizeVersion(candidate)

	// If current is a pre-release and candidate is not,
	// we need special handling
	currentIsPrerelease := semver.Prerelease(current) != ""
	candidateIsPrerelease := semver.Prerelease(candidate) != ""

	// Compare versions
	cmp := semver.Compare(current, candidate)

	if cmp < 0 {
		// candidate is newer
		return true
	}

	// If versions are equal but one is a pre-release
	if cmp == 0 {
		// If current is a pre-release and candidate is stable,
		// candidate is effectively "newer" (we're moving from pre-release to stable)
		if currentIsPrerelease && !candidateIsPrerelease {
			return true
		}
	}

	return false
}

// normalizeVersion ensures the version has the 'v' prefix required by semver.
func normalizeVersion(v string) string {
	if v == "" || v == "dev" {
		return "v0.0.0"
	}
	if v[0] != 'v' {
		return "v" + v
	}
	return v
}

// ValidateVersion checks if a string is a valid semantic version.
func ValidateVersion(v string) error {
	v = normalizeVersion(v)
	if !semver.IsValid(v) {
		return fmt.Errorf("invalid semantic version: %s", v)
	}
	return nil
}

// IsPrerelease returns true if the version is a pre-release.
func IsPrerelease(v string) bool {
	v = normalizeVersion(v)
	return semver.Prerelease(v) != ""
}
