package checker

import (
	"regexp"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

func TestToolVersionCheckerName(t *testing.T) {
	checker := NewToolVersionChecker()
	if checker.Name() != "Tool Versions" {
		t.Errorf("Expected name 'Tool Versions', got %s", checker.Name())
	}
}

func TestToolVersionCheckerCategory(t *testing.T) {
	checker := NewToolVersionChecker()
	if checker.Category() != doctor.CategoryTools {
		t.Errorf("Expected category 'tools', got %s", checker.Category())
	}
}

func TestDefaultRequirements(t *testing.T) {
	reqs := DefaultRequirements()

	if len(reqs) == 0 {
		t.Error("Expected default requirements to not be empty")
	}

	// Check that all requirements have necessary fields
	for _, req := range reqs {
		if req.Name == "" {
			t.Error("Expected requirement to have a name")
		}
		if req.MinVersion == "" {
			t.Errorf("Expected %s to have a minimum version", req.Name)
		}
		if len(req.Command) == 0 {
			t.Errorf("Expected %s to have a command", req.Name)
		}
		if req.VersionPattern == nil {
			t.Errorf("Expected %s to have a version pattern", req.Name)
		}
	}
}

func TestToolVersionCheckerCheck(t *testing.T) {
	checker := NewToolVersionChecker()
	result := checker.Check()

	// Should have a name
	if result.Name != "Tool Versions" {
		t.Errorf("Expected result name 'Tool Versions', got %s", result.Name)
	}

	// Should have valid status
	if result.Status != doctor.StatusPassed && result.Status != doctor.StatusWarning && result.Status != doctor.StatusError {
		t.Errorf("Expected valid status, got %s", result.Status)
	}

	// Should have a message
	if result.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Should have details
	if result.Details == nil {
		t.Fatal("Expected details to be present")
	}

	// Check required detail fields
	if _, ok := result.Details["toolsChecked"]; !ok {
		t.Error("Expected 'toolsChecked' in details")
	}
	if _, ok := result.Details["results"]; !ok {
		t.Error("Expected 'results' in details")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int // >0 if v1 > v2, 0 if equal, <0 if v1 < v2
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"1.1.0", "1.0.0", 1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.0", "1.1.0", -1},
		{"1.0.0", "2.0.0", -1},
		{"18.0.0", "17.0.0", 1},
		{"3.24.0", "3.0.0", 1},
		{"v18.0.0", "18.0.0", 0},
		{"V1.0.0", "1.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			result, err := compareVersions(tt.v1, tt.v2)
			if err != nil {
				t.Fatalf("compareVersions failed: %v", err)
			}

			if tt.expected > 0 && result <= 0 {
				t.Errorf("Expected %s > %s, got %d", tt.v1, tt.v2, result)
			}
			if tt.expected < 0 && result >= 0 {
				t.Errorf("Expected %s < %s, got %d", tt.v1, tt.v2, result)
			}
			if tt.expected == 0 && result != 0 {
				t.Errorf("Expected %s == %s, got %d", tt.v1, tt.v2, result)
			}
		})
	}
}

func TestCompareVersionsInvalid(t *testing.T) {
	_, err := compareVersions("invalid", "1.0.0")
	if err == nil {
		t.Error("Expected error for invalid version")
	}

	_, err = compareVersions("1.0.0", "invalid")
	if err == nil {
		t.Error("Expected error for invalid version")
	}
}

func TestParseVersionParts(t *testing.T) {
	tests := []struct {
		version  string
		expected []int
	}{
		{"1.0.0", []int{1, 0, 0}},
		{"18.20.5", []int{18, 20, 5}},
		{"v1.0.0", []int{1, 0, 0}},
		{"V2.0.0", []int{2, 0, 0}},
		{"1.0.0-alpha", []int{1, 0}}, // Stops at first non-numeric part
		{"20", []int{20}},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			parts, err := parseVersionParts(tt.version)
			if err != nil {
				t.Fatalf("parseVersionParts failed: %v", err)
			}

			if len(parts) != len(tt.expected) {
				t.Fatalf("Expected %d parts, got %d", len(tt.expected), len(parts))
			}

			for i, expected := range tt.expected {
				if parts[i] != expected {
					t.Errorf("Expected part %d to be %d, got %d", i, expected, parts[i])
				}
			}
		})
	}
}

func TestParseVersionPartsInvalid(t *testing.T) {
	_, err := parseVersionParts("abc")
	if err == nil {
		t.Error("Expected error for non-numeric version")
	}

	_, err = parseVersionParts("")
	if err == nil {
		t.Error("Expected error for empty version")
	}
}

func TestIsVersionCompatible(t *testing.T) {
	tests := []struct {
		version    string
		minVersion string
		expected   bool
	}{
		{"18.0.0", "17.0.0", true},
		{"17.0.0", "17.0.0", true},
		{"16.0.0", "17.0.0", false},
		{"3.24.0", "3.0.0", true},
		{"2.0.0", "3.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version+"_vs_"+tt.minVersion, func(t *testing.T) {
			result := IsVersionCompatible(tt.version, tt.minVersion)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsVersionCompatibleInvalid(t *testing.T) {
	// Invalid versions should return false
	result := IsVersionCompatible("invalid", "1.0.0")
	if result {
		t.Error("Expected false for invalid version")
	}
}

func TestCheckToolVersionWithMissingTool(t *testing.T) {
	req := ToolRequirement{
		Name:           "NonExistentTool",
		MinVersion:     "1.0.0",
		Command:        []string{"this_tool_definitely_does_not_exist_12345"},
		VersionArgs:    []string{"--version"},
		VersionPattern: regexp.MustCompile(`(\d+\.\d+\.\d+)`),
	}

	result := checkToolVersion(req)

	if result.Installed {
		t.Error("Expected tool to be marked as not installed")
	}
	if result.Status != "missing" {
		t.Errorf("Expected status 'missing', got '%s'", result.Status)
	}
}

func TestNewToolVersionCheckerWithRequirements(t *testing.T) {
	customReqs := []ToolRequirement{
		{
			Name:           "CustomTool",
			MinVersion:     "2.0.0",
			Command:        []string{"custom"},
			VersionArgs:    []string{"--version"},
			VersionPattern: regexp.MustCompile(`(\d+\.\d+\.\d+)`),
		},
	}

	checker := NewToolVersionCheckerWithRequirements(customReqs)

	if len(checker.requirements) != 1 {
		t.Errorf("Expected 1 requirement, got %d", len(checker.requirements))
	}

	if checker.requirements[0].Name != "CustomTool" {
		t.Errorf("Expected requirement name 'CustomTool', got '%s'", checker.requirements[0].Name)
	}
}

// BenchmarkCompareVersions benchmarks version comparison
func BenchmarkCompareVersions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = compareVersions("18.20.5", "17.0.0")
	}
}
