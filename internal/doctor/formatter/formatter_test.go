package formatter

import (
	"bytes"
	"sort"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("Expected registry to not be nil")
	}

	if registry.formatters == nil {
		t.Fatal("Expected formatters map to be initialized")
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()
	formatter := NewHumanFormatter(DefaultOptions())

	registry.Register(formatter)

	if len(registry.formatters) != 1 {
		t.Errorf("Expected 1 formatter, got %d", len(registry.formatters))
	}

	if _, ok := registry.formatters["human"]; !ok {
		t.Error("Expected 'human' formatter to be registered")
	}
}

func TestRegistryGet(t *testing.T) {
	registry := NewRegistry()
	formatter := NewHumanFormatter(DefaultOptions())
	registry.Register(formatter)

	// Get existing formatter
	f, ok := registry.Get("human")
	if !ok {
		t.Error("Expected to find 'human' formatter")
	}
	if f == nil {
		t.Error("Expected formatter to not be nil")
	}
	if f.Name() != "human" {
		t.Errorf("Expected name 'human', got %s", f.Name())
	}

	// Get non-existent formatter
	f, ok = registry.Get("nonexistent")
	if ok {
		t.Error("Expected not to find 'nonexistent' formatter")
	}
	if f != nil {
		t.Error("Expected nil formatter for non-existent name")
	}
}

func TestRegistryList(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	list := registry.List()
	if len(list) != 0 {
		t.Errorf("Expected 0 formatters, got %d", len(list))
	}

	// Register some formatters
	registry.Register(NewHumanFormatter(DefaultOptions()))
	registry.Register(NewJSONFormatter())

	list = registry.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 formatters, got %d", len(list))
	}

	// Sort for consistent comparison
	sort.Strings(list)

	if list[0] != "human" {
		t.Errorf("Expected first formatter to be 'human', got '%s'", list[0])
	}
	if list[1] != "json" {
		t.Errorf("Expected second formatter to be 'json', got '%s'", list[1])
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if !opts.UseColors {
		t.Error("Expected UseColors to be true by default")
	}

	if opts.Verbose {
		t.Error("Expected Verbose to be false by default")
	}

	if !opts.ShowPassed {
		t.Error("Expected ShowPassed to be true by default")
	}
}

func TestDefaultRegistry(t *testing.T) {
	// DefaultRegistry should have human and json formatters
	if DefaultRegistry == nil {
		t.Fatal("Expected DefaultRegistry to be initialized")
	}

	humanFormatter, ok := DefaultRegistry.Get("human")
	if !ok {
		t.Error("Expected DefaultRegistry to contain 'human' formatter")
	}
	if humanFormatter == nil {
		t.Error("Expected 'human' formatter to not be nil")
	}

	jsonFormatter, ok := DefaultRegistry.Get("json")
	if !ok {
		t.Error("Expected DefaultRegistry to contain 'json' formatter")
	}
	if jsonFormatter == nil {
		t.Error("Expected 'json' formatter to not be nil")
	}

	// Should have exactly 2 formatters
	list := DefaultRegistry.List()
	if len(list) != 2 {
		t.Errorf("Expected DefaultRegistry to have 2 formatters, got %d", len(list))
	}
}

func TestFormatterInterface(t *testing.T) {
	// Test that all formatters implement the interface correctly
	formatters := []Formatter{
		NewHumanFormatter(DefaultOptions()),
		NewJSONFormatter(),
		NewCompactJSONFormatter(),
	}

	report := &doctor.DoctorReport{}
	report.AddResult(doctor.CategoryTools, doctor.CheckResult{
		Name:    "Test",
		Status:  doctor.StatusPassed,
		Message: "Test message",
	})
	report.CalculateSummary()

	for _, formatter := range formatters {
		t.Run(formatter.Name(), func(t *testing.T) {
			// Test Name
			if formatter.Name() == "" {
				t.Error("Expected formatter name to not be empty")
			}

			// Test ContentType
			if formatter.ContentType() == "" {
				t.Error("Expected content type to not be empty")
			}

			// Test Format
			var buf bytes.Buffer
			err := formatter.Format(report, &buf)
			if err != nil {
				t.Errorf("Format failed: %v", err)
			}

			if buf.Len() == 0 {
				t.Error("Expected formatted output to not be empty")
			}
		})
	}
}

func TestRegistryMultipleRegistrations(t *testing.T) {
	registry := NewRegistry()

	// Register same name multiple times (should overwrite)
	registry.Register(NewHumanFormatter(DefaultOptions()))
	registry.Register(NewHumanFormatter(FormatOptions{UseColors: false}))

	list := registry.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 formatter after overwrite, got %d", len(list))
	}
}
