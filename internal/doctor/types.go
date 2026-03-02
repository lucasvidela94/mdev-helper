// Package doctor provides comprehensive environment diagnostics with modular checkers.
package doctor

// CheckStatus represents the result status of a check.
type CheckStatus string

const (
	// StatusPassed indicates the check passed successfully.
	StatusPassed CheckStatus = "passed"
	// StatusWarning indicates the check passed but with warnings.
	StatusWarning CheckStatus = "warning"
	// StatusError indicates the check failed.
	StatusError CheckStatus = "error"
	// StatusSkipped indicates the check was skipped.
	StatusSkipped CheckStatus = "skipped"
)

// CheckCategory represents the category of a check.
type CheckCategory string

const (
	// CategoryTools checks for development tools.
	CategoryTools CheckCategory = "tools"
	// CategoryEnvironment checks for environment configuration.
	CategoryEnvironment CheckCategory = "environment"
	// CategoryPerformance checks for performance-related issues.
	CategoryPerformance CheckCategory = "performance"
	// CategoryProject checks for project-specific configurations.
	CategoryProject CheckCategory = "project"
)

// Checker is the interface that all diagnostic checkers must implement.
type Checker interface {
	// Name returns the human-readable name of this checker.
	Name() string
	// Category returns the category this checker belongs to.
	Category() CheckCategory
	// Check performs the diagnostic check and returns the result.
	Check() CheckResult
}

// CheckResult contains the result of a diagnostic check.
type CheckResult struct {
	// Name is the identifier for this check.
	Name string `json:"name"`
	// Status is the outcome of the check.
	Status CheckStatus `json:"status"`
	// Message is a human-readable description of the result.
	Message string `json:"message"`
	// Details contains additional structured data about the check.
	Details map[string]interface{} `json:"details,omitempty"`
}

// IsPassed returns true if the check passed (including warnings).
func (r CheckResult) IsPassed() bool {
	return r.Status == StatusPassed || r.Status == StatusWarning
}

// IsCritical returns true if the check has errors.
func (r CheckResult) IsCritical() bool {
	return r.Status == StatusError
}

// HasWarnings returns true if the check has warnings.
func (r CheckResult) HasWarnings() bool {
	return r.Status == StatusWarning
}
