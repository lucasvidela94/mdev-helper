package logs

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultFormatter_Format(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC),
		Level:     LevelError,
		Tag:       "AndroidRuntime",
		Message:   "Test error message",
		Source:    "android",
	}

	formatter := NewDefaultFormatter()
	formatter.Colorize = false // Disable colors for testing

	output := formatter.Format(entry)

	// Check that output contains expected parts
	if !strings.Contains(output, "10:30:45.123") {
		t.Errorf("Expected timestamp in output, got: %s", output)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected level in output, got: %s", output)
	}
	if !strings.Contains(output, "AndroidRuntime") {
		t.Errorf("Expected tag in output, got: %s", output)
	}
	if !strings.Contains(output, "Test error message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

func TestDefaultFormatter_Format_NoTimestamp(t *testing.T) {
	entry := LogEntry{
		Level:   LevelInfo,
		Message: "Simple message",
	}

	formatter := NewDefaultFormatter()
	formatter.ShowTimestamp = false
	formatter.Colorize = false

	output := formatter.Format(entry)

	if strings.Contains(output, "10:30") {
		t.Errorf("Did not expect timestamp in output, got: %s", output)
	}
	if !strings.Contains(output, "Simple message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

func TestDefaultFormatter_Format_NoTag(t *testing.T) {
	entry := LogEntry{
		Level:   LevelInfo,
		Message: "Message without tag",
	}

	formatter := NewDefaultFormatter()
	formatter.Colorize = false

	output := formatter.Format(entry)

	if !strings.Contains(output, "Message without tag") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

func TestDefaultFormatter_Format_WithSource(t *testing.T) {
	entry := LogEntry{
		Level:   LevelInfo,
		Message: "Test",
		Source:  "metro",
	}

	formatter := NewDefaultFormatter()
	formatter.ShowSource = true
	formatter.Colorize = false

	output := formatter.Format(entry)

	if !strings.Contains(output, "[metro]") {
		t.Errorf("Expected source in output, got: %s", output)
	}
}

func TestDefaultFormatter_Colorize(t *testing.T) {
	entry := LogEntry{
		Level:   LevelError,
		Message: "Error message",
	}

	formatter := NewDefaultFormatter()
	formatter.Colorize = true

	output := formatter.Format(entry)

	// Check that ANSI codes are present
	if !strings.Contains(output, "\033[") {
		t.Errorf("Expected ANSI color codes in output, got: %s", output)
	}
}

func TestRawFormatter_Format(t *testing.T) {
	tests := []struct {
		name     string
		entry    LogEntry
		expected string
	}{
		{
			name:     "with raw field",
			entry:    LogEntry{Raw: "original raw log line", Message: "parsed"},
			expected: "original raw log line",
		},
		{
			name:     "without raw field",
			entry:    LogEntry{Message: "parsed message"},
			expected: "parsed message",
		},
	}

	formatter := &RawFormatter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.Format(tt.entry)
			if got != tt.expected {
				t.Errorf("RawFormatter.Format() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJSONFormatter_Format(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Level:     LevelInfo,
		Tag:       "TestTag",
		Message:   "Test message",
		Source:    "android",
		ProcessID: 1234,
		ThreadID:  5678,
		Package:   "com.test.app",
	}

	tests := []struct {
		name   string
		pretty bool
	}{
		{"compact", false},
		{"pretty", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &JSONFormatter{Pretty: tt.pretty}
			output := formatter.Format(entry)

			// Check for expected fields (using Contains to handle both compact and pretty formats)
			if !strings.Contains(output, `"timestamp"`) {
				t.Errorf("Expected timestamp field in JSON, got: %s", output)
			}
			if !strings.Contains(output, `"level"`) && !strings.Contains(output, `"INFO"`) {
				t.Errorf("Expected level field in JSON, got: %s", output)
			}
			if !strings.Contains(output, `"message"`) && !strings.Contains(output, `"Test message"`) {
				t.Errorf("Expected message field in JSON, got: %s", output)
			}
			if !strings.Contains(output, `"tag"`) && !strings.Contains(output, `"TestTag"`) {
				t.Errorf("Expected tag field in JSON, got: %s", output)
			}
			if !strings.Contains(output, `"source"`) && !strings.Contains(output, `"android"`) {
				t.Errorf("Expected source field in JSON, got: %s", output)
			}
			if !strings.Contains(output, `"processId"`) {
				t.Errorf("Expected processId field in JSON, got: %s", output)
			}
		})
	}
}

func TestJSONFormatter_Format_Minimal(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Level:     LevelInfo,
		Message:   "Minimal entry",
		Source:    "test",
	}

	formatter := &JSONFormatter{Pretty: false}
	output := formatter.Format(entry)

	// Should not contain empty optional fields
	if strings.Contains(output, `"tag"`) {
		t.Errorf("Did not expect tag field in minimal JSON, got: %s", output)
	}
	if strings.Contains(output, `"package"`) {
		t.Errorf("Did not expect package field in minimal JSON, got: %s", output)
	}
}

func TestSimpleFormatter_Format(t *testing.T) {
	entry := LogEntry{
		Level:   LevelWarning,
		Message: "Warning message",
	}

	formatter := &SimpleFormatter{}
	output := formatter.Format(entry)

	expected := "[WARNING] Warning message"
	if output != expected {
		t.Errorf("SimpleFormatter.Format() = %v, want %v", output, expected)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactlyten", 10, "exactlyten"},
		{"this is a long string", 10, "this is..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.expected {
				t.Errorf("truncate() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEscapeJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`simple`, `simple`},
		{`with"quotes`, `with\"quotes`},
		{`with\backslash`, `with\\backslash`},
		{"with\nnewline", `with\nnewline`},
		{"with\ttab", `with\ttab`},
		{"with\rcarriage", `with\rcarriage`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeJSON(tt.input)
			if got != tt.expected {
				t.Errorf("escapeJSON() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestColorize(t *testing.T) {
	output := colorize("test", colorRed)
	expected := "\033[31mtest\033[0m"
	if output != expected {
		t.Errorf("colorize() = %v, want %v", output, expected)
	}
}
