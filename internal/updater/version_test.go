package updater

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"equal versions", "v1.0.0", "v1.0.0", 0},
		{"v1 greater than v2", "v1.1.0", "v1.0.0", 1},
		{"v1 less than v2", "v1.0.0", "v1.1.0", -1},
		{"major version difference", "v2.0.0", "v1.9.9", 1},
		{"without v prefix", "1.0.0", "1.1.0", -1},
		{"dev version", "dev", "v1.0.0", -1},
		{"empty version", "", "v1.0.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		candidate string
		expected  bool
	}{
		{"newer version available", "v1.0.0", "v1.1.0", true},
		{"same version", "v1.0.0", "v1.0.0", false},
		{"older version", "v1.1.0", "v1.0.0", false},
		{"major version bump", "v1.0.0", "v2.0.0", true},
		{"pre-release to stable", "v1.1.0-beta.1", "v1.1.0", true},
		{"stable to pre-release", "v1.1.0", "v1.2.0-beta.1", true},
		{"dev version", "dev", "v1.0.0", true},
		{"without v prefix", "1.0.0", "1.1.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNewerVersion(tt.current, tt.candidate)
			if result != tt.expected {
				t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", tt.current, tt.candidate, result, tt.expected)
			}
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.0.0", "v1.0.0"},
		{"1.0.0", "v1.0.0"},
		{"dev", "v0.0.0"},
		{"", "v0.0.0"},
		{"v2.1.3-beta.1", "v2.1.3-beta.1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeVersion(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"valid version", "v1.0.0", false},
		{"valid with prerelease", "v1.0.0-beta.1", false},
		{"valid without v prefix", "1.0.0", false},
		{"invalid version", "not-a-version", true},
		{"empty version normalizes to valid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
			}
		})
	}
}

func TestIsPrerelease(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		{"v1.0.0", false},
		{"v1.0.0-beta.1", true},
		{"v1.0.0-alpha", true},
		{"v1.0.0-rc.1", true},
		{"1.0.0-beta.1", true},
		{"dev", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := IsPrerelease(tt.version)
			if result != tt.expected {
				t.Errorf("IsPrerelease(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}
