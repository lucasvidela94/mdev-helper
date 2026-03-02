// Package logs provides formatting utilities for log entries.
package logs

import (
	"fmt"
	"strings"
	"time"
)

// Formatter formats log entries for display.
type Formatter interface {
	// Format converts a log entry to a string.
	Format(entry LogEntry) string
}

// DefaultFormatter is the standard log formatter.
type DefaultFormatter struct {
	// ShowTimestamp includes the timestamp in output.
	ShowTimestamp bool
	// ShowLevel includes the level in output.
	ShowLevel bool
	// ShowTag includes the tag in output.
	ShowTag bool
	// ShowSource includes the source in output.
	ShowSource bool
	// Colorize enables colored output.
	Colorize bool
}

// NewDefaultFormatter creates a default formatter with common settings.
func NewDefaultFormatter() *DefaultFormatter {
	return &DefaultFormatter{
		ShowTimestamp: true,
		ShowLevel:     true,
		ShowTag:       true,
		ShowSource:    false,
		Colorize:      true,
	}
}

// Format converts a log entry to a formatted string.
func (f *DefaultFormatter) Format(entry LogEntry) string {
	var parts []string

	// Add timestamp
	if f.ShowTimestamp {
		ts := entry.Timestamp.Format("15:04:05.000")
		parts = append(parts, f.colorize(ts, colorGray))
	}

	// Add source
	if f.ShowSource {
		parts = append(parts, f.colorize(fmt.Sprintf("[%s]", entry.Source), colorCyan))
	}

	// Add level
	if f.ShowLevel {
		levelStr := fmt.Sprintf("%-7s", entry.Level.String())
		parts = append(parts, f.colorizeLevel(levelStr, entry.Level))
	}

	// Add tag
	if f.ShowTag && entry.Tag != "" {
		tag := fmt.Sprintf("%-20s", truncate(entry.Tag, 20))
		parts = append(parts, f.colorize(tag, colorYellow))
	}

	// Add message
	parts = append(parts, entry.Message)

	return strings.Join(parts, " ")
}

// colorizeLevel returns the level string with appropriate color.
func (f *DefaultFormatter) colorizeLevel(level string, lvl LogLevel) string {
	if !f.Colorize {
		return level
	}

	switch lvl {
	case LevelVerbose, LevelDebug:
		return colorize(level, colorGray)
	case LevelInfo:
		return colorize(level, colorGreen)
	case LevelWarning:
		return colorize(level, colorYellow)
	case LevelError, LevelFatal:
		return colorize(level, colorRed)
	default:
		return level
	}
}

// colorize wraps text with color codes if colorize is enabled.
func (f *DefaultFormatter) colorize(text, color string) string {
	if !f.Colorize {
		return text
	}
	return colorize(text, color)
}

// ANSI color codes.
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
)

// colorize wraps text with ANSI color codes.
func colorize(text, color string) string {
	return color + text + colorReset
}

// truncate truncates a string to a maximum length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// RawFormatter outputs log entries in their raw format.
type RawFormatter struct{}

// Format returns the raw log entry message.
func (f *RawFormatter) Format(entry LogEntry) string {
	if entry.Raw != "" {
		return entry.Raw
	}
	return entry.Message
}

// JSONFormatter outputs log entries as JSON.
type JSONFormatter struct {
	// Pretty enables pretty-printed JSON output.
	Pretty bool
}

// Format returns the log entry as a JSON string.
func (f *JSONFormatter) Format(entry LogEntry) string {
	if f.Pretty {
		return formatJSONPretty(entry)
	}
	return formatJSON(entry)
}

// formatJSON returns a compact JSON representation.
func formatJSON(entry LogEntry) string {
	// Simple JSON formatting without external dependencies
	parts := []string{
		fmt.Sprintf(`"timestamp":"%s"`, entry.Timestamp.Format(time.RFC3339Nano)),
		fmt.Sprintf(`"level":"%s"`, entry.Level.String()),
		fmt.Sprintf(`"source":"%s"`, entry.Source),
	}

	if entry.Tag != "" {
		parts = append(parts, fmt.Sprintf(`"tag":"%s"`, escapeJSON(entry.Tag)))
	}
	if entry.Package != "" {
		parts = append(parts, fmt.Sprintf(`"package":"%s"`, escapeJSON(entry.Package)))
	}
	if entry.ProcessID > 0 {
		parts = append(parts, fmt.Sprintf(`"processId":%d`, entry.ProcessID))
	}
	if entry.ThreadID > 0 {
		parts = append(parts, fmt.Sprintf(`"threadId":%d`, entry.ThreadID))
	}

	parts = append(parts, fmt.Sprintf(`"message":"%s"`, escapeJSON(entry.Message)))

	return "{" + strings.Join(parts, ",") + "}"
}

// formatJSONPretty returns a pretty-printed JSON representation.
func formatJSONPretty(entry LogEntry) string {
	parts := []string{
		fmt.Sprintf("  \"timestamp\": \"%s\"", entry.Timestamp.Format(time.RFC3339Nano)),
		fmt.Sprintf("  \"level\": \"%s\"", entry.Level.String()),
		fmt.Sprintf("  \"source\": \"%s\"", entry.Source),
	}

	if entry.Tag != "" {
		parts = append(parts, fmt.Sprintf("  \"tag\": \"%s\"", escapeJSON(entry.Tag)))
	}
	if entry.Package != "" {
		parts = append(parts, fmt.Sprintf("  \"package\": \"%s\"", escapeJSON(entry.Package)))
	}
	if entry.ProcessID > 0 {
		parts = append(parts, fmt.Sprintf("  \"processId\": %d", entry.ProcessID))
	}
	if entry.ThreadID > 0 {
		parts = append(parts, fmt.Sprintf("  \"threadId\": %d", entry.ThreadID))
	}

	parts = append(parts, fmt.Sprintf("  \"message\": \"%s\"", escapeJSON(entry.Message)))

	return "{\n" + strings.Join(parts, ",\n") + "\n}"
}

// escapeJSON escapes special characters for JSON output.
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// SimpleFormatter is a minimal formatter for basic output.
type SimpleFormatter struct{}

// Format returns a simple formatted log entry.
func (f *SimpleFormatter) Format(entry LogEntry) string {
	return fmt.Sprintf("[%s] %s", entry.Level.String(), entry.Message)
}
