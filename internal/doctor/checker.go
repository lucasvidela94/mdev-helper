package doctor

import "time"

// Runner executes a collection of checkers and aggregates results.
type Runner struct {
	checkers []Checker
}

// NewRunner creates a new Runner with no checkers.
func NewRunner() *Runner {
	return &Runner{
		checkers: make([]Checker, 0),
	}
}

// Register adds a checker to the runner.
func (r *Runner) Register(checker Checker) {
	r.checkers = append(r.checkers, checker)
}

// RegisterMultiple adds multiple checkers at once.
func (r *Runner) RegisterMultiple(checkers ...Checker) {
	r.checkers = append(r.checkers, checkers...)
}

// Run executes all registered checkers and returns a complete report.
func (r *Runner) Run() *DoctorReport {
	report := &DoctorReport{
		Timestamp:  time.Now(),
		Categories: make([]CategoryResult, 0),
		Summary:    ReportSummary{},
	}

	for _, checker := range r.checkers {
		result := checker.Check()
		report.AddResult(checker.Category(), result)
	}

	report.CalculateSummary()
	return report
}

// GetCheckers returns all registered checkers.
func (r *Runner) GetCheckers() []Checker {
	return r.checkers
}

// GetCheckersByCategory returns checkers filtered by category.
func (r *Runner) GetCheckersByCategory(category CheckCategory) []Checker {
	var filtered []Checker
	for _, checker := range r.checkers {
		if checker.Category() == category {
			filtered = append(filtered, checker)
		}
	}
	return filtered
}

// Clear removes all registered checkers.
func (r *Runner) Clear() {
	r.checkers = make([]Checker, 0)
}

// Count returns the number of registered checkers.
func (r *Runner) Count() int {
	return len(r.checkers)
}
