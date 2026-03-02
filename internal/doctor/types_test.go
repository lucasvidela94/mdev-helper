package doctor

import (
	"testing"
)

func TestCheckStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   CheckStatus
		expected string
	}{
		{"Passed", StatusPassed, "passed"},
		{"Warning", StatusWarning, "warning"},
		{"Error", StatusError, "error"},
		{"Skipped", StatusSkipped, "skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}

func TestCheckCategoryConstants(t *testing.T) {
	tests := []struct {
		name     string
		category CheckCategory
		expected string
	}{
		{"Tools", CategoryTools, "tools"},
		{"Environment", CategoryEnvironment, "environment"},
		{"Performance", CategoryPerformance, "performance"},
		{"Project", CategoryProject, "project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.category) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.category)
			}
		})
	}
}

func TestCheckResultIsPassed(t *testing.T) {
	tests := []struct {
		name     string
		status   CheckStatus
		expected bool
	}{
		{"Passed returns true", StatusPassed, true},
		{"Warning returns true", StatusWarning, true},
		{"Error returns false", StatusError, false},
		{"Skipped returns false", StatusSkipped, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckResult{Status: tt.status}
			if result.IsPassed() != tt.expected {
				t.Errorf("IsPassed() = %v, expected %v", result.IsPassed(), tt.expected)
			}
		})
	}
}

func TestCheckResultIsCritical(t *testing.T) {
	tests := []struct {
		name     string
		status   CheckStatus
		expected bool
	}{
		{"Error returns true", StatusError, true},
		{"Passed returns false", StatusPassed, false},
		{"Warning returns false", StatusWarning, false},
		{"Skipped returns false", StatusSkipped, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckResult{Status: tt.status}
			if result.IsCritical() != tt.expected {
				t.Errorf("IsCritical() = %v, expected %v", result.IsCritical(), tt.expected)
			}
		})
	}
}

func TestCheckResultHasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		status   CheckStatus
		expected bool
	}{
		{"Warning returns true", StatusWarning, true},
		{"Passed returns false", StatusPassed, false},
		{"Error returns false", StatusError, false},
		{"Skipped returns false", StatusSkipped, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckResult{Status: tt.status}
			if result.HasWarnings() != tt.expected {
				t.Errorf("HasWarnings() = %v, expected %v", result.HasWarnings(), tt.expected)
			}
		})
	}
}
