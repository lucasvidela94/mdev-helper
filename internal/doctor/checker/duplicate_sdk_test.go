package checker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

func TestDuplicateSDKCheckerName(t *testing.T) {
	checker := NewDuplicateSDKChecker()
	if checker.Name() != "Duplicate SDK Detection" {
		t.Errorf("Expected name 'Duplicate SDK Detection', got %s", checker.Name())
	}
}

func TestDuplicateSDKCheckerCategory(t *testing.T) {
	checker := NewDuplicateSDKChecker()
	if checker.Category() != doctor.CategoryEnvironment {
		t.Errorf("Expected category 'environment', got %s", checker.Category())
	}
}

func TestDuplicateSDKCheckerCheck(t *testing.T) {
	checker := NewDuplicateSDKChecker()
	result := checker.Check()

	// Should have a name
	if result.Name != "Duplicate SDK Detection" {
		t.Errorf("Expected result name 'Duplicate SDK Detection', got %s", result.Name)
	}

	// Should have valid status
	if result.Status != doctor.StatusPassed && result.Status != doctor.StatusWarning {
		t.Errorf("Expected status 'passed' or 'warning', got %s", result.Status)
	}

	// Should have a message
	if result.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Should have details
	if result.Details == nil {
		t.Fatal("Expected details to be present")
	}
}

func TestIsValidSDKPath(t *testing.T) {
	// Create a temporary directory structure that looks like an SDK
	tempDir := t.TempDir()

	// Without required directories, should be invalid
	if isValidSDKPath(tempDir) {
		t.Error("Expected invalid SDK path without required directories")
	}

	// Create required directories
	os.MkdirAll(filepath.Join(tempDir, "platforms"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "build-tools"), 0755)

	// Now should be valid
	if !isValidSDKPath(tempDir) {
		t.Error("Expected valid SDK path with required directories")
	}
}

func TestIsValidSDKPathNonExistent(t *testing.T) {
	if isValidSDKPath("/nonexistent/path/that/does/not/exist") {
		t.Error("Expected non-existent path to be invalid")
	}
}

func TestIsValidSDKPathNotDirectory(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "sdk")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	if isValidSDKPath(tempFile.Name()) {
		t.Error("Expected file path to be invalid (not a directory)")
	}
}

func TestGetSDKPlatforms(t *testing.T) {
	// Create a temporary SDK structure
	tempDir := t.TempDir()
	platformsDir := filepath.Join(tempDir, "platforms")
	os.MkdirAll(filepath.Join(platformsDir, "android-30"), 0755)
	os.MkdirAll(filepath.Join(platformsDir, "android-31"), 0755)
	os.MkdirAll(filepath.Join(platformsDir, "android-32"), 0755)
	os.MkdirAll(filepath.Join(platformsDir, "not-a-platform"), 0755) // Should be ignored

	platforms, err := GetSDKPlatforms(tempDir)
	if err != nil {
		t.Fatalf("GetSDKPlatforms failed: %v", err)
	}

	if len(platforms) != 3 {
		t.Errorf("Expected 3 platforms, got %d", len(platforms))
	}

	// Check that we got the right platforms
	platformMap := make(map[string]bool)
	for _, p := range platforms {
		platformMap[p] = true
	}

	if !platformMap["android-30"] {
		t.Error("Expected android-30 in platforms")
	}
	if !platformMap["android-31"] {
		t.Error("Expected android-31 in platforms")
	}
	if !platformMap["android-32"] {
		t.Error("Expected android-32 in platforms")
	}
	if platformMap["not-a-platform"] {
		t.Error("Did not expect 'not-a-platform' in platforms")
	}
}

func TestGetSDKPlatformsInvalidPath(t *testing.T) {
	_, err := GetSDKPlatforms("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestGetSDKBuildTools(t *testing.T) {
	// Create a temporary SDK structure
	tempDir := t.TempDir()
	buildToolsDir := filepath.Join(tempDir, "build-tools")
	os.MkdirAll(filepath.Join(buildToolsDir, "30.0.0"), 0755)
	os.MkdirAll(filepath.Join(buildToolsDir, "31.0.0"), 0755)
	os.MkdirAll(filepath.Join(buildToolsDir, "32.0.0"), 0755)

	versions, err := GetSDKBuildTools(tempDir)
	if err != nil {
		t.Fatalf("GetSDKBuildTools failed: %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("Expected 3 build tool versions, got %d", len(versions))
	}
}

func TestGetSDKBuildToolsInvalidPath(t *testing.T) {
	_, err := GetSDKBuildTools("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestDetectConflicts(t *testing.T) {
	// Create two temporary SDK structures with overlapping platforms
	tempDir1 := t.TempDir()
	platformsDir1 := filepath.Join(tempDir1, "platforms")
	os.MkdirAll(filepath.Join(platformsDir1, "android-30"), 0755)
	os.MkdirAll(filepath.Join(platformsDir1, "android-31"), 0755)
	os.MkdirAll(filepath.Join(tempDir1, "build-tools"), 0755)

	tempDir2 := t.TempDir()
	platformsDir2 := filepath.Join(tempDir2, "platforms")
	os.MkdirAll(filepath.Join(platformsDir2, "android-30"), 0755) // Overlapping
	os.MkdirAll(filepath.Join(platformsDir2, "android-32"), 0755)
	os.MkdirAll(filepath.Join(tempDir2, "build-tools"), 0755)

	sdks := []SDKLocation{
		{Path: tempDir1, Source: "test1"},
		{Path: tempDir2, Source: "test2"},
	}

	conflicts := DetectConflicts(sdks)

	// Should detect conflict for android-30
	if len(conflicts) == 0 {
		t.Error("Expected conflicts to be detected")
	}

	foundAndroid30Conflict := false
	for _, conflict := range conflicts {
		if contains(conflict, "android-30") {
			foundAndroid30Conflict = true
			break
		}
	}

	if !foundAndroid30Conflict {
		t.Error("Expected conflict for android-30 platform")
	}
}

func TestDetectConflictsNoConflict(t *testing.T) {
	// Create two temporary SDK structures with no overlapping platforms
	tempDir1 := t.TempDir()
	platformsDir1 := filepath.Join(tempDir1, "platforms")
	os.MkdirAll(filepath.Join(platformsDir1, "android-30"), 0755)
	os.MkdirAll(filepath.Join(tempDir1, "build-tools"), 0755)

	tempDir2 := t.TempDir()
	platformsDir2 := filepath.Join(tempDir2, "platforms")
	os.MkdirAll(filepath.Join(platformsDir2, "android-31"), 0755)
	os.MkdirAll(filepath.Join(tempDir2, "build-tools"), 0755)

	sdks := []SDKLocation{
		{Path: tempDir1, Source: "test1"},
		{Path: tempDir2, Source: "test2"},
	}

	conflicts := DetectConflicts(sdks)

	// Should not detect any conflicts
	if len(conflicts) > 0 {
		t.Errorf("Expected no conflicts, got %d: %v", len(conflicts), conflicts)
	}
}

func TestDetectConflictsSingleSDK(t *testing.T) {
	sdks := []SDKLocation{
		{Path: "/some/path", Source: "test"},
	}

	conflicts := DetectConflicts(sdks)

	// Should not detect any conflicts with only one SDK
	if len(conflicts) > 0 {
		t.Errorf("Expected no conflicts with single SDK, got %d", len(conflicts))
	}
}

func TestDuplicateSDKCheckerWithAdditionalPaths(t *testing.T) {
	// Create a temporary SDK structure
	tempDir := t.TempDir()
	os.MkdirAll(filepath.Join(tempDir, "platforms"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "build-tools"), 0755)

	checker := NewDuplicateSDKChecker()
	checker.AdditionalPaths = []string{tempDir}

	result := checker.Check()

	// Should include the additional path in details
	if result.Details == nil {
		t.Fatal("Expected details to be present")
	}

	validSDKs, ok := result.Details["validSDKs"].([]SDKLocation)
	if !ok {
		t.Fatal("Expected validSDKs to be a []SDKLocation")
	}

	// Should find our additional SDK
	found := false
	for _, sdk := range validSDKs {
		if sdk.Path == tempDir {
			found = true
			if sdk.Source != "user" {
				t.Errorf("Expected source 'user', got '%s'", sdk.Source)
			}
			break
		}
	}

	if !found {
		t.Error("Expected additional SDK path to be found")
	}
}

func TestGetDefaultSDKPaths(t *testing.T) {
	paths := getDefaultSDKPaths()

	if len(paths) == 0 {
		t.Error("Expected default SDK paths to not be empty")
	}

	// All paths should be non-empty
	for _, path := range paths {
		if path == "" {
			t.Error("Expected default path to not be empty")
		}
	}
}

func TestGetAlternativeSDKPaths(t *testing.T) {
	paths := getAlternativeSDKPaths()

	if len(paths) == 0 {
		t.Error("Expected alternative SDK paths to not be empty")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
