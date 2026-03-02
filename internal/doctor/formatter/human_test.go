package formatter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

func TestHumanFormatterName(t *testing.T) {
	formatter := NewHumanFormatter(DefaultOptions())
	if formatter.Name() != "human" {
		t.Errorf("Expected name 'human', got %s", formatter.Name())
	}
}

func TestHumanFormatterContentType(t *testing.T) {
	formatter := NewHumanFormatter(DefaultOptions())
	if formatter.ContentType() != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got %s", formatter.ContentType())
	}
}

func TestHumanFormatterFormat(t *testing.T) {
	formatter := NewHumanFormatter(FormatOptions{UseColors: false, ShowPassed: true})

	report := createTestReport()
	var buf bytes.Buffer

	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	// Should contain header
	if !strings.Contains(output, "Environment Diagnosis") {
		t.Error("Expected output to contain 'Environment Diagnosis'")
	}

	// Should contain check names
	if !strings.Contains(output, "Test Check 1") {
		t.Error("Expected output to contain 'Test Check 1'")
	}

	// Should contain summary
	if !strings.Contains(output, "Summary") {
		t.Error("Expected output to contain 'Summary'")
	}
}

func TestHumanFormatterWithColors(t *testing.T) {
	formatter := NewHumanFormatter(FormatOptions{UseColors: true, ShowPassed: true})

	report := createTestReport()
	var buf bytes.Buffer

	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	// Should contain ANSI color codes
	if !strings.Contains(output, "\033[") {
		t.Error("Expected output to contain ANSI color codes")
	}
}

func TestHumanFormatterHidePassed(t *testing.T) {
	formatter := NewHumanFormatter(FormatOptions{UseColors: false, ShowPassed: false})

	report := createTestReport()
	var buf bytes.Buffer

	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	// Should not contain passed check message
	if strings.Contains(output, "Everything is fine") {
		t.Error("Expected passed check to be hidden")
	}

	// Should still contain error check
	if !strings.Contains(output, "Something is wrong") {
		t.Error("Expected error check to be shown")
	}
}

func TestGetStatusIcon(t *testing.T) {
	formatter := NewHumanFormatter(DefaultOptions())

	tests := []struct {
		status   doctor.CheckStatus
		expected string
	}{
		{doctor.StatusPassed, "✓"},
		{doctor.StatusWarning, "⚠"},
		{doctor.StatusError, "✗"},
		{doctor.StatusSkipped, "○"},
		{doctor.CheckStatus("unknown"), "?"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := formatter.getStatusIcon(tt.status)
			if result != tt.expected {
				t.Errorf("Expected icon '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetStatusColor(t *testing.T) {
	formatter := NewHumanFormatter(DefaultOptions())

	tests := []struct {
		status   doctor.CheckStatus
		expected string
	}{
		{doctor.StatusPassed, "32"},
		{doctor.StatusWarning, "33"},
		{doctor.StatusError, "31"},
		{doctor.StatusSkipped, "90"},
		{doctor.CheckStatus("unknown"), "0"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := formatter.getStatusColor(tt.status)
			if result != tt.expected {
				t.Errorf("Expected color '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestStatusIconsMap(t *testing.T) {
	if StatusIcons[doctor.StatusPassed] != "✓" {
		t.Error("Expected passed icon to be '✓'")
	}
	if StatusIcons[doctor.StatusWarning] != "⚠" {
		t.Error("Expected warning icon to be '⚠'")
	}
	if StatusIcons[doctor.StatusError] != "✗" {
		t.Error("Expected error icon to be '✗'")
	}
	if StatusIcons[doctor.StatusSkipped] != "○" {
		t.Error("Expected skipped icon to be '○'")
	}
}

func TestFormatCategoryHeader(t *testing.T) {
	formatter := NewHumanFormatter(FormatOptions{UseColors: false})
	result := formatter.formatCategoryHeader(doctor.CategoryTools)

	if !strings.Contains(result, "TOOLS") {
		t.Error("Expected category header to contain 'TOOLS'")
	}
}

func TestFormatSummaryWithErrors(t *testing.T) {
	formatter := NewHumanFormatter(FormatOptions{UseColors: false})

	summary := doctor.ReportSummary{
		TotalChecks: 5,
		Passed:      2,
		Warnings:    1,
		Errors:      2,
		ExitCode:    2,
		HasErrors:   true,
		HasWarnings: true,
	}

	result := formatter.formatSummary(summary)

	if !strings.Contains(result, "issues that need attention") {
		t.Error("Expected summary to mention issues")
	}

	if !strings.Contains(result, "5 total") {
		t.Error("Expected summary to show total checks")
	}
}

func TestFormatSummaryWithWarningsOnly(t *testing.T) {
	formatter := NewHumanFormatter(FormatOptions{UseColors: false})

	summary := doctor.ReportSummary{
		TotalChecks: 5,
		Passed:      4,
		Warnings:    1,
		Errors:      0,
		ExitCode:    1,
		HasErrors:   false,
		HasWarnings: true,
	}

	result := formatter.formatSummary(summary)

	if !strings.Contains(result, "warnings") {
		t.Error("Expected summary to mention warnings")
	}
}

func TestFormatSummaryAllGood(t *testing.T) {
	formatter := NewHumanFormatter(FormatOptions{UseColors: false})

	summary := doctor.ReportSummary{
		TotalChecks: 5,
		Passed:      5,
		Warnings:    0,
		Errors:      0,
		ExitCode:    0,
		HasErrors:   false,
		HasWarnings: false,
	}

	result := formatter.formatSummary(summary)

	if !strings.Contains(result, "healthy") {
		t.Error("Expected summary to mention healthy environment")
	}
}

// Helper function to create a test report
func createTestReport() *doctor.DoctorReport {
	report := &doctor.DoctorReport{}

	report.AddResult(doctor.CategoryTools, doctor.CheckResult{
		Name:    "Test Check 1",
		Status:  doctor.StatusPassed,
		Message: "Everything is fine",
	})

	report.AddResult(doctor.CategoryTools, doctor.CheckResult{
		Name:    "Test Check 2",
		Status:  doctor.StatusError,
		Message: "Something is wrong",
	})

	report.AddResult(doctor.CategoryEnvironment, doctor.CheckResult{
		Name:    "Test Check 3",
		Status:  doctor.StatusWarning,
		Message: "Pay attention to this",
	})

	report.CalculateSummary()
	return report
}
