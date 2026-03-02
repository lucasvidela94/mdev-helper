// Package logs provides core types and utilities for log streaming.
package logs

import (
	"context"
	"time"
)

// LogLevel represents the severity level of a log entry.
type LogLevel int

const (
	// LevelVerbose is the most detailed log level.
	LevelVerbose LogLevel = iota
	// LevelDebug is for debugging information.
	LevelDebug
	// LevelInfo is for general information.
	LevelInfo
	// LevelWarning is for warning messages.
	LevelWarning
	// LevelError is for error messages.
	LevelError
	// LevelFatal is for fatal errors.
	LevelFatal
	// LevelSilent suppresses all logs.
	LevelSilent
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LevelVerbose:
		return "VERBOSE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	case LevelSilent:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a log level from a string.
func ParseLogLevel(s string) LogLevel {
	switch s {
	case "V", "VERBOSE":
		return LevelVerbose
	case "D", "DEBUG":
		return LevelDebug
	case "I", "INFO":
		return LevelInfo
	case "W", "WARN", "WARNING":
		return LevelWarning
	case "E", "ERR", "ERROR":
		return LevelError
	case "F", "FATAL":
		return LevelFatal
	case "S", "SILENT":
		return LevelSilent
	default:
		return LevelDebug
	}
}

// LogEntry represents a single log entry from any source.
type LogEntry struct {
	// Timestamp is when the log entry was created.
	Timestamp time.Time `json:"timestamp"`
	// Level is the severity level of the log entry.
	Level LogLevel `json:"level"`
	// Tag is an optional identifier for the log source (e.g., Android tag).
	Tag string `json:"tag,omitempty"`
	// Message is the actual log message.
	Message string `json:"message"`
	// Source identifies where this log came from (e.g., "android", "metro", "flutter").
	Source string `json:"source"`
	// ProcessID is the process ID if available.
	ProcessID int `json:"processId,omitempty"`
	// ThreadID is the thread ID if available.
	ThreadID int `json:"threadId,omitempty"`
	// Package is the package name if available (Android-specific).
	Package string `json:"package,omitempty"`
	// Raw contains the original raw log line.
	Raw string `json:"raw,omitempty"`
}

// StreamOptions contains options for streaming logs.
type StreamOptions struct {
	// Follow continuously streams new log entries.
	Follow bool
	// Filter is a text filter to apply to log messages.
	Filter string
	// Level filters logs by minimum severity level.
	Level LogLevel
	// Tag filters logs by tag (Android-specific).
	Tag string
	// Package filters logs by package name (Android-specific).
	Package string
	// Lines limits the number of lines to output (0 = unlimited).
	Lines int
	// Since filters logs after a specific time.
	Since time.Time
	// Until filters logs before a specific time.
	Until time.Time
}

// LogStreamer is the interface that all log sources must implement.
type LogStreamer interface {
	// Stream starts streaming logs and returns a channel of log entries.
	// The channel will be closed when streaming ends or the context is cancelled.
	Stream(ctx context.Context, opts StreamOptions) (<-chan LogEntry, error)
	// Name returns the name of this log source.
	Name() string
	// CheckPrerequisites verifies that the necessary tools are available.
	CheckPrerequisites() error
}

// StreamResult contains the result of a streaming operation.
type StreamResult struct {
	// Entries is the total number of log entries processed.
	Entries int
	// Filtered is the number of entries that passed the filter.
	Filtered int
	// Errors contains any errors encountered during streaming.
	Errors []error
	// Duration is how long the streaming operation took.
	Duration time.Duration
}
