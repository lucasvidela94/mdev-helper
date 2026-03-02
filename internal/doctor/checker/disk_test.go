package checker

import (
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

func TestDiskCheckerName(t *testing.T) {
	checker := NewDiskChecker(".")
	if checker.Name() != "Disk Space" {
		t.Errorf("Expected name 'Disk Space', got %s", checker.Name())
	}
}

func TestDiskCheckerCategory(t *testing.T) {
	checker := NewDiskChecker(".")
	if checker.Category() != doctor.CategoryEnvironment {
		t.Errorf("Expected category 'environment', got %s", checker.Category())
	}
}

func TestDiskCheckerCheck(t *testing.T) {
	// Use temp directory for testing
	tempDir := t.TempDir()
	checker := NewDiskChecker(tempDir)

	result := checker.Check()

	// Should have a name
	if result.Name != "Disk Space" {
		t.Errorf("Expected result name 'Disk Space', got %s", result.Name)
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
		t.Error("Expected details to be present")
	}

	// Check required detail fields
	requiredFields := []string{"path", "freeBytes", "totalBytes", "freeGB", "totalGB", "usedGB", "percentUsed", "minFreeGB"}
	for _, field := range requiredFields {
		if _, ok := result.Details[field]; !ok {
			t.Errorf("Expected detail field '%s' to be present", field)
		}
	}
}

func TestDiskCheckerWithEmptyPath(t *testing.T) {
	// Empty path should default to current directory
	checker := NewDiskChecker("")
	if checker.path != "." {
		t.Errorf("Expected path to be '.', got %s", checker.path)
	}
}

func TestDiskCheckerWithInvalidPath(t *testing.T) {
	// Use a path that doesn't exist
	checker := NewDiskChecker("/nonexistent/path/that/does/not/exist")

	result := checker.Check()

	// Should return error status
	if result.Status != doctor.StatusError {
		t.Errorf("Expected status 'error' for invalid path, got %s", result.Status)
	}

	// Should have error message
	if result.Message == "" {
		t.Error("Expected error message for invalid path")
	}
}

func TestGetDiskSpace(t *testing.T) {
	tempDir := t.TempDir()

	freeBytes, totalBytes, err := getDiskSpace(tempDir)
	if err != nil {
		t.Fatalf("getDiskSpace failed: %v", err)
	}

	// Should return positive values
	if freeBytes == 0 {
		t.Error("Expected freeBytes to be greater than 0")
	}
	if totalBytes == 0 {
		t.Error("Expected totalBytes to be greater than 0")
	}
	if freeBytes > totalBytes {
		t.Error("Expected freeBytes to be less than or equal to totalBytes")
	}
}

func TestGetDiskSpaceInvalidPath(t *testing.T) {
	_, _, err := getDiskSpace("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestGetFreeSpaceGB(t *testing.T) {
	tempDir := t.TempDir()

	freeGB, err := GetFreeSpaceGB(tempDir)
	if err != nil {
		t.Fatalf("GetFreeSpaceGB failed: %v", err)
	}

	// Should return positive value
	if freeGB <= 0 {
		t.Errorf("Expected freeGB to be greater than 0, got %f", freeGB)
	}
}

func TestGetFreeSpaceGBInvalidPath(t *testing.T) {
	_, err := GetFreeSpaceGB("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestDiskCheckerLowSpaceScenario(t *testing.T) {
	// This test verifies the checker's behavior when low on space
	// We can't easily simulate low disk space, but we can verify the logic

	checker := NewDiskChecker(".")
	result := checker.Check()

	// Verify the result has the expected structure
	if result.Details == nil {
		t.Fatal("Expected details to be present")
	}

	freeGB, ok := result.Details["freeGB"].(float64)
	if !ok {
		t.Fatal("Expected freeGB to be a float64")
	}

	minFreeGB := result.Details["minFreeGB"].(int)

	// Verify status matches the free space
	if freeGB < float64(minFreeGB) {
		if result.Status != doctor.StatusWarning {
			t.Errorf("Expected warning when free space (%.1f GB) is below minimum (%d GB)", freeGB, minFreeGB)
		}
		if result.Message == "" {
			t.Error("Expected warning message for low disk space")
		}
	} else {
		if result.Status != doctor.StatusPassed {
			t.Errorf("Expected passed when free space (%.1f GB) is above minimum (%d GB)", freeGB, minFreeGB)
		}
	}
}

// TestDiskCheckerResultConsistency verifies that the result is consistent
func TestDiskCheckerResultConsistency(t *testing.T) {
	tempDir := t.TempDir()
	checker := NewDiskChecker(tempDir)

	result1 := checker.Check()
	result2 := checker.Check()

	// Status should be the same
	if result1.Status != result2.Status {
		t.Errorf("Status inconsistent: first=%s, second=%s", result1.Status, result2.Status)
	}

	// Free bytes should be the same (or very close)
	free1 := result1.Details["freeBytes"].(uint64)
	free2 := result2.Details["freeBytes"].(uint64)

	// Allow for small changes due to background activity
	diff := int64(free1) - int64(free2)
	if diff < 0 {
		diff = -diff
	}

	// Allow up to 1MB difference
	if diff > 1024*1024 {
		t.Errorf("Free space inconsistent: first=%d, second=%d, diff=%d", free1, free2, diff)
	}
}

// BenchmarkDiskChecker benchmarks the disk checker
func BenchmarkDiskChecker(b *testing.B) {
	tempDir := b.TempDir()
	checker := NewDiskChecker(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checker.Check()
	}
}
