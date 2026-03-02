package formatter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

func TestJSONFormatterName(t *testing.T) {
	formatter := NewJSONFormatter()
	if formatter.Name() != "json" {
		t.Errorf("Expected name 'json', got %s", formatter.Name())
	}
}

func TestJSONFormatterContentType(t *testing.T) {
	formatter := NewJSONFormatter()
	if formatter.ContentType() != "application/json" {
		t.Errorf("Expected content type 'application/json', got %s", formatter.ContentType())
	}
}

func TestJSONFormatterFormat(t *testing.T) {
	formatter := NewJSONFormatter()

	report := createTestJSONReport()
	var buf bytes.Buffer

	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Should contain timestamp
	if _, ok := result["timestamp"]; !ok {
		t.Error("Expected JSON to contain 'timestamp'")
	}

	// Should contain categories
	if _, ok := result["categories"]; !ok {
		t.Error("Expected JSON to contain 'categories'")
	}

	// Should contain summary
	if _, ok := result["summary"]; !ok {
		t.Error("Expected JSON to contain 'summary'")
	}
}

func TestJSONFormatterPretty(t *testing.T) {
	formatter := NewJSONFormatter()

	report := createTestJSONReport()
	var buf bytes.Buffer

	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	// Pretty printed JSON should contain newlines and indentation
	if !strings.Contains(output, "\n") {
		t.Error("Expected pretty printed JSON to contain newlines")
	}

	if !strings.Contains(output, "  ") {
		t.Error("Expected pretty printed JSON to contain indentation")
	}
}

func TestJSONFormatterCompact(t *testing.T) {
	formatter := NewCompactJSONFormatter()

	report := createTestJSONReport()
	var buf bytes.Buffer

	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	// Compact JSON should not contain newlines (except final)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 1 {
		t.Error("Expected compact JSON to be on a single line")
	}
}

func TestJSONFormatterSetPretty(t *testing.T) {
	formatter := NewJSONFormatter()
	formatter.SetPretty(false)

	report := createTestJSONReport()
	var buf bytes.Buffer

	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()

	// Should now be compact
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 1 {
		t.Error("Expected JSON to be compact after SetPretty(false)")
	}
}

func TestJSONFormatterStructure(t *testing.T) {
	formatter := NewJSONFormatter()

	report := createTestJSONReport()
	var buf bytes.Buffer

	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Check categories
	if len(output.Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(output.Categories))
	}

	// Check first category
	if output.Categories[0].Category != "tools" {
		t.Errorf("Expected first category to be 'tools', got '%s'", output.Categories[0].Category)
	}

	// Check summary
	if output.Summary.TotalChecks != 3 {
		t.Errorf("Expected 3 total checks, got %d", output.Summary.TotalChecks)
	}

	if output.Summary.Passed != 1 {
		t.Errorf("Expected 1 passed, got %d", output.Summary.Passed)
	}

	if output.Summary.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", output.Summary.Errors)
	}

	if output.Summary.Warnings != 1 {
		t.Errorf("Expected 1 warning, got %d", output.Summary.Warnings)
	}

	if output.Summary.ExitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", output.Summary.ExitCode)
	}
}

func TestJSONFormatterEmptyReport(t *testing.T) {
	formatter := NewJSONFormatter()

	report := &doctor.DoctorReport{}
	report.CalculateSummary()

	var buf bytes.Buffer
	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Should still be valid - categories might be nil or empty array
	categories, ok := result["categories"].([]interface{})
	if ok && len(categories) != 0 {
		t.Errorf("Expected 0 categories, got %d", len(categories))
	}
}

func TestJSONFormatterCheckDetails(t *testing.T) {
	formatter := NewJSONFormatter()

	report := &doctor.DoctorReport{}
	report.AddResult(doctor.CategoryTools, doctor.CheckResult{
		Name:    "Detailed Check",
		Status:  doctor.StatusPassed,
		Message: "Check passed",
		Details: map[string]interface{}{
			"version": "1.0.0",
			"count":   42,
			"enabled": true,
		},
	})
	report.CalculateSummary()

	var buf bytes.Buffer
	err := formatter.Format(report, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Check that details are included
	if len(output.Categories) == 0 {
		t.Fatal("Expected at least one category")
	}

	if len(output.Categories[0].Checks) == 0 {
		t.Fatal("Expected at least one check")
	}

	details := output.Categories[0].Checks[0].Details
	if details == nil {
		t.Fatal("Expected details to be present")
	}

	if details["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %v", details["version"])
	}
}

// Helper function to create a test report
func createTestJSONReport() *doctor.DoctorReport {
	report := &doctor.DoctorReport{}

	report.AddResult(doctor.CategoryTools, doctor.CheckResult{
		Name:    "Tool Check",
		Status:  doctor.StatusPassed,
		Message: "Tool is installed",
	})

	report.AddResult(doctor.CategoryTools, doctor.CheckResult{
		Name:    "Version Check",
		Status:  doctor.StatusError,
		Message: "Version is too old",
	})

	report.AddResult(doctor.CategoryEnvironment, doctor.CheckResult{
		Name:    "Env Check",
		Status:  doctor.StatusWarning,
		Message: "Environment variable not set",
	})

	report.CalculateSummary()
	return report
}
