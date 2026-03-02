package doctor

import (
	"testing"
	"time"
)

func TestDoctorReportAddResult(t *testing.T) {
	report := &DoctorReport{}

	// Add a passed result
	report.AddResult(CategoryTools, CheckResult{
		Name:   "Test 1",
		Status: StatusPassed,
	})

	if report.Summary.TotalChecks != 1 {
		t.Errorf("Expected 1 total check, got %d", report.Summary.TotalChecks)
	}
	if report.Summary.Passed != 1 {
		t.Errorf("Expected 1 passed, got %d", report.Summary.Passed)
	}

	// Add a warning result
	report.AddResult(CategoryTools, CheckResult{
		Name:   "Test 2",
		Status: StatusWarning,
	})

	if report.Summary.TotalChecks != 2 {
		t.Errorf("Expected 2 total checks, got %d", report.Summary.TotalChecks)
	}
	if report.Summary.Warnings != 1 {
		t.Errorf("Expected 1 warning, got %d", report.Summary.Warnings)
	}
	if !report.Summary.HasWarnings {
		t.Error("Expected HasWarnings to be true")
	}

	// Add an error result
	report.AddResult(CategoryEnvironment, CheckResult{
		Name:   "Test 3",
		Status: StatusError,
	})

	if report.Summary.TotalChecks != 3 {
		t.Errorf("Expected 3 total checks, got %d", report.Summary.TotalChecks)
	}
	if report.Summary.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", report.Summary.Errors)
	}
	if !report.Summary.HasErrors {
		t.Error("Expected HasErrors to be true")
	}
}

func TestDoctorReportCalculateSummary(t *testing.T) {
	tests := []struct {
		name             string
		hasErrors        bool
		hasWarnings      bool
		expectedExitCode int
	}{
		{"All good", false, false, 0},
		{"Warnings only", false, true, 1},
		{"Errors present", true, false, 2},
		{"Errors and warnings", true, true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &DoctorReport{
				Summary: ReportSummary{
					HasErrors:   tt.hasErrors,
					HasWarnings: tt.hasWarnings,
				},
			}

			report.CalculateSummary()

			if report.Summary.ExitCode != tt.expectedExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectedExitCode, report.Summary.ExitCode)
			}
		})
	}
}

func TestDoctorReportGetCategory(t *testing.T) {
	report := &DoctorReport{}

	// Add results to different categories
	report.AddResult(CategoryTools, CheckResult{Name: "Tool Check", Status: StatusPassed})
	report.AddResult(CategoryEnvironment, CheckResult{Name: "Env Check", Status: StatusPassed})

	// Get existing category
	toolsCategory := report.GetCategory(CategoryTools)
	if toolsCategory == nil {
		t.Fatal("Expected to find tools category")
	}
	if toolsCategory.Category != CategoryTools {
		t.Errorf("Expected category 'tools', got '%s'", toolsCategory.Category)
	}
	if len(toolsCategory.Checks) != 1 {
		t.Errorf("Expected 1 check in tools category, got %d", len(toolsCategory.Checks))
	}

	// Get non-existent category
	nonExistent := report.GetCategory(CategoryPerformance)
	if nonExistent != nil {
		t.Error("Expected nil for non-existent category")
	}
}

func TestDoctorReportGetFailedChecks(t *testing.T) {
	report := &DoctorReport{}

	// Add various results
	report.AddResult(CategoryTools, CheckResult{Name: "Passed", Status: StatusPassed})
	report.AddResult(CategoryTools, CheckResult{Name: "Warning", Status: StatusWarning})
	report.AddResult(CategoryEnvironment, CheckResult{Name: "Error", Status: StatusError})

	failed := report.GetFailedChecks()

	// Should have 2 failed checks (warning + error)
	if len(failed) != 2 {
		t.Errorf("Expected 2 failed checks, got %d", len(failed))
	}

	// Verify the failed checks
	foundWarning := false
	foundError := false
	for _, check := range failed {
		if check.Name == "Warning" {
			foundWarning = true
		}
		if check.Name == "Error" {
			foundError = true
		}
	}

	if !foundWarning {
		t.Error("Expected to find warning check in failed checks")
	}
	if !foundError {
		t.Error("Expected to find error check in failed checks")
	}
}

func TestDoctorReportGetFailedChecksAllPassed(t *testing.T) {
	report := &DoctorReport{}

	// Add only passed results
	report.AddResult(CategoryTools, CheckResult{Name: "Passed 1", Status: StatusPassed})
	report.AddResult(CategoryTools, CheckResult{Name: "Passed 2", Status: StatusPassed})

	failed := report.GetFailedChecks()

	if len(failed) != 0 {
		t.Errorf("Expected 0 failed checks, got %d", len(failed))
	}
}

func TestCategoryResultCounts(t *testing.T) {
	report := &DoctorReport{}

	// Add multiple results to the same category
	report.AddResult(CategoryTools, CheckResult{Status: StatusPassed})
	report.AddResult(CategoryTools, CheckResult{Status: StatusPassed})
	report.AddResult(CategoryTools, CheckResult{Status: StatusWarning})
	report.AddResult(CategoryTools, CheckResult{Status: StatusError})

	toolsCategory := report.GetCategory(CategoryTools)
	if toolsCategory == nil {
		t.Fatal("Expected to find tools category")
	}

	if toolsCategory.PassedCount != 2 {
		t.Errorf("Expected 2 passed, got %d", toolsCategory.PassedCount)
	}
	if toolsCategory.WarningCount != 1 {
		t.Errorf("Expected 1 warning, got %d", toolsCategory.WarningCount)
	}
	if toolsCategory.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", toolsCategory.ErrorCount)
	}
}

func TestDoctorReportTimestamp(t *testing.T) {
	before := time.Now()

	report := &DoctorReport{
		Timestamp: time.Now(),
	}

	after := time.Now()

	if report.Timestamp.Before(before) {
		t.Error("Timestamp should not be before test start")
	}
	if report.Timestamp.After(after) {
		t.Error("Timestamp should not be after test end")
	}
}

func TestReportSummaryFields(t *testing.T) {
	summary := ReportSummary{
		TotalChecks: 10,
		Passed:      7,
		Warnings:    2,
		Errors:      1,
		ExitCode:    2,
		HasErrors:   true,
		HasWarnings: true,
	}

	if summary.TotalChecks != 10 {
		t.Errorf("Expected TotalChecks=10, got %d", summary.TotalChecks)
	}
	if summary.Passed != 7 {
		t.Errorf("Expected Passed=7, got %d", summary.Passed)
	}
	if summary.Warnings != 2 {
		t.Errorf("Expected Warnings=2, got %d", summary.Warnings)
	}
	if summary.Errors != 1 {
		t.Errorf("Expected Errors=1, got %d", summary.Errors)
	}
	if summary.ExitCode != 2 {
		t.Errorf("Expected ExitCode=2, got %d", summary.ExitCode)
	}
	if !summary.HasErrors {
		t.Error("Expected HasErrors=true")
	}
	if !summary.HasWarnings {
		t.Error("Expected HasWarnings=true")
	}
}
