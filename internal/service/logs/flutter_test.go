package logs

import (
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/logs"
)

func TestFlutterStreamer_Name(t *testing.T) {
	streamer := NewFlutterStreamer()
	if streamer.Name() != "flutter" {
		t.Errorf("Name() = %v, want flutter", streamer.Name())
	}
}

func TestFlutterStreamer_parseFlutterLine(t *testing.T) {
	streamer := NewFlutterStreamer()

	tests := []struct {
		name          string
		line          string
		expectMatch   bool
		expectedTag   string
		expectedMsg   string
		expectedLevel logs.LogLevel
		expectedPID   int
	}{
		{
			name:          "standard flutter log",
			line:          "I/flutter (1234): Test message",
			expectMatch:   true,
			expectedTag:   "flutter",
			expectedMsg:   "Test message",
			expectedLevel: logs.LevelInfo,
			expectedPID:   1234,
		},
		{
			name:          "flutter log with timing",
			line:          "[  +123 ms] I/flutter (5678): Message with timing",
			expectMatch:   true,
			expectedTag:   "flutter",
			expectedMsg:   "Message with timing",
			expectedLevel: logs.LevelInfo,
			expectedPID:   5678,
		},
		{
			name:          "error log",
			line:          "E/flutter (9999): Error occurred",
			expectMatch:   true,
			expectedTag:   "flutter",
			expectedMsg:   "Error occurred",
			expectedLevel: logs.LevelError,
			expectedPID:   9999,
		},
		{
			name:          "warning log",
			line:          "W/DartVM  (1111): Warning message",
			expectMatch:   true,
			expectedTag:   "DartVM",
			expectedMsg:   "Warning message",
			expectedLevel: logs.LevelWarning,
			expectedPID:   1111,
		},
		{
			name:          "debug log",
			line:          "D/flutter (2222): Debug info",
			expectMatch:   true,
			expectedTag:   "flutter",
			expectedMsg:   "Debug info",
			expectedLevel: logs.LevelDebug,
			expectedPID:   2222,
		},
		{
			name:          "verbose log",
			line:          "V/flutter (3333): Verbose output",
			expectMatch:   true,
			expectedTag:   "flutter",
			expectedMsg:   "Verbose output",
			expectedLevel: logs.LevelVerbose,
			expectedPID:   3333,
		},
		{
			name:          "fatal log",
			line:          "F/flutter (4444): Fatal error",
			expectMatch:   true,
			expectedTag:   "flutter",
			expectedMsg:   "Fatal error",
			expectedLevel: logs.LevelFatal,
			expectedPID:   4444,
		},
		{
			name:          "non-standard line",
			line:          "This is not a flutter log",
			expectMatch:   false,
			expectedMsg:   "This is not a flutter log",
			expectedLevel: logs.LevelInfo,
		},
		{
			name:        "empty line",
			line:        "",
			expectMatch: false,
			expectedMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := streamer.parseFlutterLine(tt.line)

			if entry.Source != "flutter" {
				t.Errorf("Expected source 'flutter', got %s", entry.Source)
			}

			if entry.Raw != tt.line {
				t.Errorf("Expected raw line to match input")
			}

			if tt.expectMatch {
				if entry.Tag != tt.expectedTag {
					t.Errorf("Expected tag %q, got %q", tt.expectedTag, entry.Tag)
				}
				if entry.Message != tt.expectedMsg {
					t.Errorf("Expected message %q, got %q", tt.expectedMsg, entry.Message)
				}
				if entry.Level != tt.expectedLevel {
					t.Errorf("Expected level %v, got %v", tt.expectedLevel, entry.Level)
				}
				if entry.ProcessID != tt.expectedPID {
					t.Errorf("Expected PID %d, got %d", tt.expectedPID, entry.ProcessID)
				}
			} else {
				if entry.Message != tt.expectedMsg {
					t.Errorf("Expected message %q for non-matching line, got %q", tt.expectedMsg, entry.Message)
				}
			}
		})
	}
}

func TestFlutterStreamer_priorityToLevel(t *testing.T) {
	streamer := NewFlutterStreamer()

	tests := []struct {
		priority string
		expected logs.LogLevel
	}{
		{"V", logs.LevelVerbose},
		{"D", logs.LevelDebug},
		{"I", logs.LevelInfo},
		{"W", logs.LevelWarning},
		{"E", logs.LevelError},
		{"F", logs.LevelFatal},
		{"X", logs.LevelInfo}, // unknown defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			got := streamer.priorityToLevel(tt.priority)
			if got != tt.expected {
				t.Errorf("priorityToLevel(%q) = %v, want %v", tt.priority, got, tt.expected)
			}
		})
	}
}
