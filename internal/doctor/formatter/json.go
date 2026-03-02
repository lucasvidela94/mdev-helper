package formatter

import (
	"encoding/json"
	"io"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

// JSONFormatter formats reports as JSON.
type JSONFormatter struct {
	pretty bool
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{pretty: true}
}

// NewCompactJSONFormatter creates a JSON formatter without pretty printing.
func NewCompactJSONFormatter() *JSONFormatter {
	return &JSONFormatter{pretty: false}
}

// Name returns the formatter name.
func (j *JSONFormatter) Name() string {
	return "json"
}

// ContentType returns the MIME type.
func (j *JSONFormatter) ContentType() string {
	return "application/json"
}

// Format formats the report as JSON.
func (j *JSONFormatter) Format(report *doctor.DoctorReport, w io.Writer) error {
	encoder := json.NewEncoder(w)
	if j.pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(report)
}

// SetPretty enables or disables pretty printing.
func (j *JSONFormatter) SetPretty(pretty bool) {
	j.pretty = pretty
}

// JSONOutput represents the JSON structure for testing and documentation.
type JSONOutput struct {
	Timestamp  string           `json:"timestamp"`
	Categories []CategoryOutput `json:"categories"`
	Summary    SummaryOutput    `json:"summary"`
}

// CategoryOutput represents a category in JSON output.
type CategoryOutput struct {
	Category     string        `json:"category"`
	Checks       []CheckOutput `json:"checks"`
	PassedCount  int           `json:"passedCount"`
	WarningCount int           `json:"warningCount"`
	ErrorCount   int           `json:"errorCount"`
}

// CheckOutput represents a check in JSON output.
type CheckOutput struct {
	Name    string                 `json:"name"`
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SummaryOutput represents the summary in JSON output.
type SummaryOutput struct {
	TotalChecks int  `json:"totalChecks"`
	Passed      int  `json:"passed"`
	Warnings    int  `json:"warnings"`
	Errors      int  `json:"errors"`
	ExitCode    int  `json:"exitCode"`
	HasErrors   bool `json:"hasErrors"`
	HasWarnings bool `json:"hasWarnings"`
}
