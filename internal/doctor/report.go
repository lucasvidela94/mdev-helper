package doctor

import "time"

// DoctorReport contains the complete results of all diagnostic checks.
type DoctorReport struct {
	// Timestamp when the report was generated.
	Timestamp time.Time `json:"timestamp"`
	// Categories contains results grouped by category.
	Categories []CategoryResult `json:"categories"`
	// Summary provides a high-level overview.
	Summary ReportSummary `json:"summary"`
}

// CategoryResult contains all check results for a specific category.
type CategoryResult struct {
	// Category is the type of checks in this group.
	Category CheckCategory `json:"category"`
	// Checks contains all check results for this category.
	Checks []CheckResult `json:"checks"`
	// PassedCount is the number of passed checks.
	PassedCount int `json:"passedCount"`
	// WarningCount is the number of checks with warnings.
	WarningCount int `json:"warningCount"`
	// ErrorCount is the number of failed checks.
	ErrorCount int `json:"errorCount"`
}

// ReportSummary provides a high-level summary of the entire report.
type ReportSummary struct {
	// TotalChecks is the total number of checks run.
	TotalChecks int `json:"totalChecks"`
	// Passed is the number of checks that passed.
	Passed int `json:"passed"`
	// Warnings is the number of checks with warnings.
	Warnings int `json:"warnings"`
	// Errors is the number of failed checks.
	Errors int `json:"errors"`
	// ExitCode is the recommended exit code (0=all good, 1=warnings, 2=errors).
	ExitCode int `json:"exitCode"`
	// HasErrors indicates if any check failed.
	HasErrors bool `json:"hasErrors"`
	// HasWarnings indicates if any check has warnings.
	HasWarnings bool `json:"hasWarnings"`
}

// AddResult adds a check result to the appropriate category.
func (r *DoctorReport) AddResult(category CheckCategory, result CheckResult) {
	// Find or create category
	var catIdx int
	found := false
	for i, cat := range r.Categories {
		if cat.Category == category {
			catIdx = i
			found = true
			break
		}
	}

	if !found {
		r.Categories = append(r.Categories, CategoryResult{
			Category: category,
			Checks:   []CheckResult{},
		})
		catIdx = len(r.Categories) - 1
	}

	// Add the result
	r.Categories[catIdx].Checks = append(r.Categories[catIdx].Checks, result)

	// Update counts
	switch result.Status {
	case StatusPassed:
		r.Categories[catIdx].PassedCount++
		r.Summary.Passed++
	case StatusWarning:
		r.Categories[catIdx].WarningCount++
		r.Summary.Warnings++
		r.Summary.HasWarnings = true
	case StatusError:
		r.Categories[catIdx].ErrorCount++
		r.Summary.Errors++
		r.Summary.HasErrors = true
	}

	r.Summary.TotalChecks++
}

// CalculateSummary computes the final summary values including exit code.
func (r *DoctorReport) CalculateSummary() {
	if r.Summary.HasErrors {
		r.Summary.ExitCode = 2
	} else if r.Summary.HasWarnings {
		r.Summary.ExitCode = 1
	} else {
		r.Summary.ExitCode = 0
	}
}

// GetCategory returns the results for a specific category.
func (r *DoctorReport) GetCategory(category CheckCategory) *CategoryResult {
	for i := range r.Categories {
		if r.Categories[i].Category == category {
			return &r.Categories[i]
		}
	}
	return nil
}

// GetFailedChecks returns all checks that failed or have warnings.
func (r *DoctorReport) GetFailedChecks() []CheckResult {
	var failed []CheckResult
	for _, cat := range r.Categories {
		for _, check := range cat.Checks {
			if check.Status == StatusError || check.Status == StatusWarning {
				failed = append(failed, check)
			}
		}
	}
	return failed
}
