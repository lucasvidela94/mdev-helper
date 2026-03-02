package formatter

import (
	"io"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

// Formatter is the interface for output formatters.
type Formatter interface {
	// Format formats a DoctorReport and writes it to the provided writer.
	Format(report *doctor.DoctorReport, w io.Writer) error
	// Name returns the formatter name.
	Name() string
	// ContentType returns the MIME type of the output.
	ContentType() string
}

// FormatOptions contains options for formatting.
type FormatOptions struct {
	// UseColors enables ANSI color codes (for human formatters).
	UseColors bool
	// Verbose includes additional details.
	Verbose bool
	// ShowPassed includes passed checks in output.
	ShowPassed bool
}

// DefaultOptions returns the default format options.
func DefaultOptions() FormatOptions {
	return FormatOptions{
		UseColors:  true,
		Verbose:    false,
		ShowPassed: true,
	}
}

// Registry holds available formatters.
type Registry struct {
	formatters map[string]Formatter
}

// NewRegistry creates a new formatter registry.
func NewRegistry() *Registry {
	return &Registry{
		formatters: make(map[string]Formatter),
	}
}

// Register adds a formatter to the registry.
func (r *Registry) Register(formatter Formatter) {
	r.formatters[formatter.Name()] = formatter
}

// Get retrieves a formatter by name.
func (r *Registry) Get(name string) (Formatter, bool) {
	formatter, ok := r.formatters[name]
	return formatter, ok
}

// List returns all registered formatter names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.formatters))
	for name := range r.formatters {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry contains the standard formatters.
var DefaultRegistry = func() *Registry {
	r := NewRegistry()
	r.Register(NewHumanFormatter(DefaultOptions()))
	r.Register(NewJSONFormatter())
	return r
}()
