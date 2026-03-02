package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

// HumanFormatter formats reports for human readability with colors and icons.
type HumanFormatter struct {
	options FormatOptions
}

// NewHumanFormatter creates a new human-readable formatter.
func NewHumanFormatter(options FormatOptions) *HumanFormatter {
	return &HumanFormatter{options: options}
}

// Name returns the formatter name.
func (h *HumanFormatter) Name() string {
	return "human"
}

// ContentType returns the MIME type.
func (h *HumanFormatter) ContentType() string {
	return "text/plain"
}

// Format formats the report for human consumption.
func (h *HumanFormatter) Format(report *doctor.DoctorReport, w io.Writer) error {
	var output strings.Builder

	// Header
	output.WriteString(h.formatHeader("Environment Diagnosis"))
	output.WriteString("\n")

	// Categories
	for _, category := range report.Categories {
		if len(category.Checks) == 0 {
			continue
		}

		// Category header
		output.WriteString(h.formatCategoryHeader(category.Category))

		// Checks
		for _, check := range category.Checks {
			if !h.options.ShowPassed && check.Status == doctor.StatusPassed {
				continue
			}
			output.WriteString(h.formatCheck(check))
		}

		output.WriteString("\n")
	}

	// Summary
	output.WriteString(h.formatSummary(report.Summary))

	_, err := w.Write([]byte(output.String()))
	return err
}

// formatHeader formats the main header.
func (h *HumanFormatter) formatHeader(title string) string {
	if h.options.UseColors {
		return fmt.Sprintf("\033[1m=== %s ===\033[0m", title)
	}
	return fmt.Sprintf("=== %s ===", title)
}

// formatCategoryHeader formats a category header.
func (h *HumanFormatter) formatCategoryHeader(category doctor.CheckCategory) string {
	categoryName := strings.ToUpper(string(category))
	if h.options.UseColors {
		return fmt.Sprintf("\n\033[1;34m%s:\033[0m\n", categoryName)
	}
	return fmt.Sprintf("\n%s:\n", categoryName)
}

// formatCheck formats a single check result.
func (h *HumanFormatter) formatCheck(check doctor.CheckResult) string {
	icon := h.getStatusIcon(check.Status)
	color := h.getStatusColor(check.Status)

	var result strings.Builder
	if h.options.UseColors {
		result.WriteString(fmt.Sprintf("  %s \033[%sm%s\033[0m: %s\n", icon, color, check.Name, check.Message))
	} else {
		result.WriteString(fmt.Sprintf("  %s %s: %s\n", icon, check.Name, check.Message))
	}

	// Add details if verbose
	if h.options.Verbose && len(check.Details) > 0 {
		for key, value := range check.Details {
			result.WriteString(fmt.Sprintf("    %s: %v\n", key, value))
		}
	}

	return result.String()
}

// formatSummary formats the report summary.
func (h *HumanFormatter) formatSummary(summary doctor.ReportSummary) string {
	var output strings.Builder

	output.WriteString("\n" + h.formatHeader("Summary") + "\n")

	// Status line
	if summary.HasErrors {
		if h.options.UseColors {
			output.WriteString("\033[1;31m✗ Environment has issues that need attention.\033[0m\n")
		} else {
			output.WriteString("✗ Environment has issues that need attention.\n")
		}
	} else if summary.HasWarnings {
		if h.options.UseColors {
			output.WriteString("\033[1;33m✓ No critical errors, but there are warnings.\033[0m\n")
		} else {
			output.WriteString("✓ No critical errors, but there are warnings.\n")
		}
	} else {
		if h.options.UseColors {
			output.WriteString("\033[1;32m✓ Environment looks healthy!\033[0m\n")
		} else {
			output.WriteString("✓ Environment looks healthy!\n")
		}
	}

	// Statistics
	output.WriteString(fmt.Sprintf("\nChecks: %d total, ", summary.TotalChecks))
	if h.options.UseColors {
		output.WriteString(fmt.Sprintf("\033[32m%d passed\033[0m, ", summary.Passed))
		output.WriteString(fmt.Sprintf("\033[33m%d warnings\033[0m, ", summary.Warnings))
		output.WriteString(fmt.Sprintf("\033[31m%d errors\033[0m", summary.Errors))
	} else {
		output.WriteString(fmt.Sprintf("%d passed, ", summary.Passed))
		output.WriteString(fmt.Sprintf("%d warnings, ", summary.Warnings))
		output.WriteString(fmt.Sprintf("%d errors", summary.Errors))
	}
	output.WriteString("\n")

	// Exit code hint
	if summary.ExitCode > 0 {
		output.WriteString(fmt.Sprintf("Exit code: %d\n", summary.ExitCode))
	}

	return output.String()
}

// getStatusIcon returns the appropriate icon for a status.
func (h *HumanFormatter) getStatusIcon(status doctor.CheckStatus) string {
	switch status {
	case doctor.StatusPassed:
		return "✓"
	case doctor.StatusWarning:
		return "⚠"
	case doctor.StatusError:
		return "✗"
	case doctor.StatusSkipped:
		return "○"
	default:
		return "?"
	}
}

// getStatusColor returns the ANSI color code for a status.
func (h *HumanFormatter) getStatusColor(status doctor.CheckStatus) string {
	switch status {
	case doctor.StatusPassed:
		return "32" // Green
	case doctor.StatusWarning:
		return "33" // Yellow
	case doctor.StatusError:
		return "31" // Red
	case doctor.StatusSkipped:
		return "90" // Gray
	default:
		return "0" // Default
	}
}

// StatusIcons maps statuses to their icons (exported for testing).
var StatusIcons = map[doctor.CheckStatus]string{
	doctor.StatusPassed:  "✓",
	doctor.StatusWarning: "⚠",
	doctor.StatusError:   "✗",
	doctor.StatusSkipped: "○",
}
