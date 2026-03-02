package doctor

import (
	"testing"
)

// MockChecker is a test implementation of the Checker interface.
type MockChecker struct {
	name     string
	category CheckCategory
	result   CheckResult
}

func (m *MockChecker) Name() string {
	return m.name
}

func (m *MockChecker) Category() CheckCategory {
	return m.category
}

func (m *MockChecker) Check() CheckResult {
	return m.result
}

func TestNewRunner(t *testing.T) {
	runner := NewRunner()
	if runner == nil {
		t.Fatal("Expected runner to not be nil")
	}

	if runner.checkers == nil {
		t.Fatal("Expected checkers slice to be initialized")
	}

	if runner.Count() != 0 {
		t.Errorf("Expected 0 checkers, got %d", runner.Count())
	}
}

func TestRunnerRegister(t *testing.T) {
	runner := NewRunner()

	checker := &MockChecker{
		name:     "Test",
		category: CategoryTools,
		result:   CheckResult{Name: "Test", Status: StatusPassed},
	}

	runner.Register(checker)

	if runner.Count() != 1 {
		t.Errorf("Expected 1 checker, got %d", runner.Count())
	}
}

func TestRunnerRegisterMultiple(t *testing.T) {
	runner := NewRunner()

	checkers := []Checker{
		&MockChecker{name: "Test1", category: CategoryTools, result: CheckResult{Status: StatusPassed}},
		&MockChecker{name: "Test2", category: CategoryEnvironment, result: CheckResult{Status: StatusPassed}},
		&MockChecker{name: "Test3", category: CategoryTools, result: CheckResult{Status: StatusPassed}},
	}

	runner.RegisterMultiple(checkers...)

	if runner.Count() != 3 {
		t.Errorf("Expected 3 checkers, got %d", runner.Count())
	}
}

func TestRunnerGetCheckers(t *testing.T) {
	runner := NewRunner()

	checkers := []Checker{
		&MockChecker{name: "Test1", category: CategoryTools},
		&MockChecker{name: "Test2", category: CategoryEnvironment},
	}

	runner.RegisterMultiple(checkers...)

	retrieved := runner.GetCheckers()
	if len(retrieved) != 2 {
		t.Errorf("Expected 2 checkers, got %d", len(retrieved))
	}

	// Verify we got the right checkers
	if retrieved[0].Name() != "Test1" {
		t.Errorf("Expected first checker name 'Test1', got '%s'", retrieved[0].Name())
	}
	if retrieved[1].Name() != "Test2" {
		t.Errorf("Expected second checker name 'Test2', got '%s'", retrieved[1].Name())
	}
}

func TestRunnerGetCheckersByCategory(t *testing.T) {
	runner := NewRunner()

	checkers := []Checker{
		&MockChecker{name: "Tool1", category: CategoryTools},
		&MockChecker{name: "Env1", category: CategoryEnvironment},
		&MockChecker{name: "Tool2", category: CategoryTools},
		&MockChecker{name: "Perf1", category: CategoryPerformance},
	}

	runner.RegisterMultiple(checkers...)

	tools := runner.GetCheckersByCategory(CategoryTools)
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools checkers, got %d", len(tools))
	}

	env := runner.GetCheckersByCategory(CategoryEnvironment)
	if len(env) != 1 {
		t.Errorf("Expected 1 environment checker, got %d", len(env))
	}

	perf := runner.GetCheckersByCategory(CategoryPerformance)
	if len(perf) != 1 {
		t.Errorf("Expected 1 performance checker, got %d", len(perf))
	}

	project := runner.GetCheckersByCategory(CategoryProject)
	if len(project) != 0 {
		t.Errorf("Expected 0 project checkers, got %d", len(project))
	}
}

func TestRunnerClear(t *testing.T) {
	runner := NewRunner()

	runner.Register(&MockChecker{name: "Test", category: CategoryTools})

	if runner.Count() != 1 {
		t.Fatal("Expected 1 checker before clear")
	}

	runner.Clear()

	if runner.Count() != 0 {
		t.Errorf("Expected 0 checkers after clear, got %d", runner.Count())
	}
}

func TestRunnerRun(t *testing.T) {
	runner := NewRunner()

	checkers := []Checker{
		&MockChecker{
			name:     "Passing",
			category: CategoryTools,
			result:   CheckResult{Name: "Passing", Status: StatusPassed, Message: "OK"},
		},
		&MockChecker{
			name:     "Warning",
			category: CategoryEnvironment,
			result:   CheckResult{Name: "Warning", Status: StatusWarning, Message: "Careful"},
		},
		&MockChecker{
			name:     "Error",
			category: CategoryTools,
			result:   CheckResult{Name: "Error", Status: StatusError, Message: "Failed"},
		},
	}

	runner.RegisterMultiple(checkers...)

	report := runner.Run()

	// Check summary
	if report.Summary.TotalChecks != 3 {
		t.Errorf("Expected 3 total checks, got %d", report.Summary.TotalChecks)
	}
	if report.Summary.Passed != 1 {
		t.Errorf("Expected 1 passed, got %d", report.Summary.Passed)
	}
	if report.Summary.Warnings != 1 {
		t.Errorf("Expected 1 warning, got %d", report.Summary.Warnings)
	}
	if report.Summary.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", report.Summary.Errors)
	}
	if report.Summary.ExitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", report.Summary.ExitCode)
	}

	// Check categories
	if len(report.Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(report.Categories))
	}

	// Tools category should have 2 checks
	toolsCategory := report.GetCategory(CategoryTools)
	if toolsCategory == nil {
		t.Fatal("Expected tools category")
	}
	if len(toolsCategory.Checks) != 2 {
		t.Errorf("Expected 2 checks in tools category, got %d", len(toolsCategory.Checks))
	}

	// Environment category should have 1 check
	envCategory := report.GetCategory(CategoryEnvironment)
	if envCategory == nil {
		t.Fatal("Expected environment category")
	}
	if len(envCategory.Checks) != 1 {
		t.Errorf("Expected 1 check in environment category, got %d", len(envCategory.Checks))
	}
}

func TestRunnerRunEmpty(t *testing.T) {
	runner := NewRunner()

	report := runner.Run()

	if report.Summary.TotalChecks != 0 {
		t.Errorf("Expected 0 total checks, got %d", report.Summary.TotalChecks)
	}

	if report.Summary.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", report.Summary.ExitCode)
	}

	if len(report.Categories) != 0 {
		t.Errorf("Expected 0 categories, got %d", len(report.Categories))
	}
}

func TestRunnerRunAllPassed(t *testing.T) {
	runner := NewRunner()

	runner.RegisterMultiple(
		&MockChecker{
			name:     "Check1",
			category: CategoryTools,
			result:   CheckResult{Status: StatusPassed},
		},
		&MockChecker{
			name:     "Check2",
			category: CategoryTools,
			result:   CheckResult{Status: StatusPassed},
		},
	)

	report := runner.Run()

	if report.Summary.ExitCode != 0 {
		t.Errorf("Expected exit code 0 for all passed, got %d", report.Summary.ExitCode)
	}

	if report.Summary.HasErrors {
		t.Error("Expected HasErrors to be false")
	}

	if report.Summary.HasWarnings {
		t.Error("Expected HasWarnings to be false")
	}
}

func TestRunnerRunAllWarnings(t *testing.T) {
	runner := NewRunner()

	runner.RegisterMultiple(
		&MockChecker{
			name:     "Check1",
			category: CategoryTools,
			result:   CheckResult{Status: StatusWarning},
		},
		&MockChecker{
			name:     "Check2",
			category: CategoryTools,
			result:   CheckResult{Status: StatusWarning},
		},
	)

	report := runner.Run()

	if report.Summary.ExitCode != 1 {
		t.Errorf("Expected exit code 1 for warnings only, got %d", report.Summary.ExitCode)
	}

	if report.Summary.HasErrors {
		t.Error("Expected HasErrors to be false")
	}

	if !report.Summary.HasWarnings {
		t.Error("Expected HasWarnings to be true")
	}
}
