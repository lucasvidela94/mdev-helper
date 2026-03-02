package logs

import (
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/logs"
)

func TestAndroidStreamer_parseLogcatLine(t *testing.T) {
	streamer := NewAndroidStreamer()

	tests := []struct {
		name          string
		line          string
		expectMatch   bool
		expectedTag   string
		expectedMsg   string
		expectedLevel logs.LogLevel
	}{
		{
			name:          "standard logcat line",
			line:          "01-15 10:30:45.123  1234  5678 D AndroidRuntime: Test message",
			expectMatch:   true,
			expectedTag:   "AndroidRuntime",
			expectedMsg:   "Test message",
			expectedLevel: logs.LevelDebug,
		},
		{
			name:          "error level",
			line:          "01-15 10:30:45.123  1234  5678 E ReactNativeJS: Error occurred",
			expectMatch:   true,
			expectedTag:   "ReactNativeJS",
			expectedMsg:   "Error occurred",
			expectedLevel: logs.LevelError,
		},
		{
			name:          "warning level",
			line:          "01-15 10:30:45.123  1234  5678 W System: Warning message",
			expectMatch:   true,
			expectedTag:   "System",
			expectedMsg:   "Warning message",
			expectedLevel: logs.LevelWarning,
		},
		{
			name:          "info level",
			line:          "01-15 10:30:45.123  1234  5678 I ActivityManager: Info message",
			expectMatch:   true,
			expectedTag:   "ActivityManager",
			expectedMsg:   "Info message",
			expectedLevel: logs.LevelInfo,
		},
		{
			name:          "verbose level",
			line:          "01-15 10:30:45.123  1234  5678 V SomeTag: Verbose message",
			expectMatch:   true,
			expectedTag:   "SomeTag",
			expectedMsg:   "Verbose message",
			expectedLevel: logs.LevelVerbose,
		},
		{
			name:          "fatal level",
			line:          "01-15 10:30:45.123  1234  5678 F FatalTag: Fatal message",
			expectMatch:   true,
			expectedTag:   "FatalTag",
			expectedMsg:   "Fatal message",
			expectedLevel: logs.LevelFatal,
		},
		{
			name:          "non-standard line",
			line:          "This is not a logcat line",
			expectMatch:   false,
			expectedMsg:   "This is not a logcat line",
			expectedLevel: logs.LevelDebug,
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
			entry := streamer.parseLogcatLine(tt.line)

			if entry.Source != "android" {
				t.Errorf("Expected source 'android', got %s", entry.Source)
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
			} else {
				// For non-matching lines, message should be the whole line
				if entry.Message != tt.expectedMsg {
					t.Errorf("Expected message %q for non-matching line, got %q", tt.expectedMsg, entry.Message)
				}
			}
		})
	}
}

func TestAndroidStreamer_levelToPriority(t *testing.T) {
	streamer := NewAndroidStreamer()

	tests := []struct {
		level    logs.LogLevel
		expected string
	}{
		{logs.LevelVerbose, "V"},
		{logs.LevelDebug, "D"},
		{logs.LevelInfo, "I"},
		{logs.LevelWarning, "W"},
		{logs.LevelError, "E"},
		{logs.LevelFatal, "F"},
		{logs.LevelSilent, "D"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			got := streamer.levelToPriority(tt.level)
			if got != tt.expected {
				t.Errorf("levelToPriority(%v) = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestAndroidStreamer_priorityToLevel(t *testing.T) {
	streamer := NewAndroidStreamer()

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
		{"X", logs.LevelDebug}, // unknown defaults to debug
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

func TestAndroidStreamer_buildLogcatArgs(t *testing.T) {
	streamer := NewAndroidStreamer()

	tests := []struct {
		name     string
		opts     logs.StreamOptions
		expected []string
	}{
		{
			name:     "basic args",
			opts:     logs.StreamOptions{},
			expected: []string{"logcat", "-v", "threadtime", "-d"},
		},
		{
			name:     "with follow",
			opts:     logs.StreamOptions{Follow: true},
			expected: []string{"logcat", "-v", "threadtime"},
		},
		{
			name:     "with level filter",
			opts:     logs.StreamOptions{Level: logs.LevelWarning},
			expected: []string{"logcat", "-v", "threadtime", "*:W", "-d"},
		},
		{
			name:     "with tag filter",
			opts:     logs.StreamOptions{Tag: "ReactNativeJS"},
			expected: []string{"logcat", "-v", "threadtime", "ReactNativeJS:D", "-d"},
		},
		{
			name:     "with tag and level",
			opts:     logs.StreamOptions{Tag: "AndroidRuntime", Level: logs.LevelError},
			expected: []string{"logcat", "-v", "threadtime", "AndroidRuntime:E", "-d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := streamer.buildLogcatArgs(tt.opts)
			if len(got) != len(tt.expected) {
				t.Errorf("buildLogcatArgs() = %v, want %v", got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("buildLogcatArgs()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestAndroidStreamer_Name(t *testing.T) {
	streamer := NewAndroidStreamer()
	if streamer.Name() != "android" {
		t.Errorf("Name() = %v, want android", streamer.Name())
	}
}
