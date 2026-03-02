package logs

import (
	"testing"
	"time"
)

func TestTextFilter_Match(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		invert   bool
		entry    LogEntry
		expected bool
	}{
		{
			name:     "match in message",
			pattern:  "error",
			entry:    LogEntry{Message: "This is an error message"},
			expected: true,
		},
		{
			name:     "no match in message",
			pattern:  "error",
			entry:    LogEntry{Message: "This is a success message"},
			expected: false,
		},
		{
			name:     "match in tag",
			pattern:  "ReactNative",
			entry:    LogEntry{Message: "Some message", Tag: "ReactNativeJS"},
			expected: true,
		},
		{
			name:     "match in package",
			pattern:  "com.example",
			entry:    LogEntry{Message: "Some message", Package: "com.example.app"},
			expected: true,
		},
		{
			name:     "case insensitive match",
			pattern:  "ERROR",
			entry:    LogEntry{Message: "this is an error"},
			expected: true,
		},
		{
			name:     "empty pattern matches all",
			pattern:  "",
			entry:    LogEntry{Message: "any message"},
			expected: true,
		},
		{
			name:     "invert match",
			pattern:  "error",
			invert:   true,
			entry:    LogEntry{Message: "This is an error"},
			expected: false,
		},
		{
			name:     "invert no match",
			pattern:  "error",
			invert:   true,
			entry:    LogEntry{Message: "This is fine"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &TextFilter{
				Pattern: tt.pattern,
				Invert:  tt.invert,
			}
			got := filter.Match(tt.entry)
			if got != tt.expected {
				t.Errorf("TextFilter.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRegexFilter_Match(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		invert   bool
		entry    LogEntry
		expected bool
	}{
		{
			name:     "match pattern",
			pattern:  `error.*\d+`,
			entry:    LogEntry{Message: "error code 123"},
			expected: true,
		},
		{
			name:     "no match pattern",
			pattern:  `error.*\d+`,
			entry:    LogEntry{Message: "success"},
			expected: false,
		},
		{
			name:     "match in tag",
			pattern:  `React.*`,
			entry:    LogEntry{Message: "msg", Tag: "ReactNative"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewRegexFilter(tt.pattern)
			if err != nil {
				t.Fatalf("NewRegexFilter() error = %v", err)
			}
			filter.Invert = tt.invert
			got := filter.Match(tt.entry)
			if got != tt.expected {
				t.Errorf("RegexFilter.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLevelFilter_Match(t *testing.T) {
	tests := []struct {
		name     string
		minLevel LogLevel
		entry    LogEntry
		expected bool
	}{
		{
			name:     "exact level match",
			minLevel: LevelWarning,
			entry:    LogEntry{Level: LevelWarning},
			expected: true,
		},
		{
			name:     "higher level passes",
			minLevel: LevelWarning,
			entry:    LogEntry{Level: LevelError},
			expected: true,
		},
		{
			name:     "lower level fails",
			minLevel: LevelWarning,
			entry:    LogEntry{Level: LevelDebug},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &LevelFilter{MinLevel: tt.minLevel}
			got := filter.Match(tt.entry)
			if got != tt.expected {
				t.Errorf("LevelFilter.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTagFilter_Match(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		invert   bool
		entry    LogEntry
		expected bool
	}{
		{
			name:     "exact match",
			tag:      "ReactNativeJS",
			entry:    LogEntry{Tag: "ReactNativeJS"},
			expected: true,
		},
		{
			name:     "case insensitive match",
			tag:      "reactnativejs",
			entry:    LogEntry{Tag: "ReactNativeJS"},
			expected: true,
		},
		{
			name:     "no match",
			tag:      "ReactNativeJS",
			entry:    LogEntry{Tag: "AndroidRuntime"},
			expected: false,
		},
		{
			name:     "empty tag matches all",
			tag:      "",
			entry:    LogEntry{Tag: "anything"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &TagFilter{
				Tag:    tt.tag,
				Invert: tt.invert,
			}
			got := filter.Match(tt.entry)
			if got != tt.expected {
				t.Errorf("TagFilter.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPackageFilter_Match(t *testing.T) {
	tests := []struct {
		name     string
		pkg      string
		invert   bool
		entry    LogEntry
		expected bool
	}{
		{
			name:     "exact match",
			pkg:      "com.example.app",
			entry:    LogEntry{Package: "com.example.app"},
			expected: true,
		},
		{
			name:     "case insensitive match",
			pkg:      "com.example.app",
			entry:    LogEntry{Package: "COM.EXAMPLE.APP"},
			expected: true,
		},
		{
			name:     "no match",
			pkg:      "com.example.app",
			entry:    LogEntry{Package: "com.other.app"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &PackageFilter{
				Package: tt.pkg,
				Invert:  tt.invert,
			}
			got := filter.Match(tt.entry)
			if got != tt.expected {
				t.Errorf("PackageFilter.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCompoundFilter_Match(t *testing.T) {
	tests := []struct {
		name     string
		filters  []Filter
		entry    LogEntry
		expected bool
	}{
		{
			name: "all filters match",
			filters: []Filter{
				&TextFilter{Pattern: "error"},
				&LevelFilter{MinLevel: LevelWarning},
			},
			entry:    LogEntry{Message: "an error occurred", Level: LevelError},
			expected: true,
		},
		{
			name: "one filter fails",
			filters: []Filter{
				&TextFilter{Pattern: "error"},
				&LevelFilter{MinLevel: LevelWarning},
			},
			entry:    LogEntry{Message: "an error occurred", Level: LevelDebug},
			expected: false,
		},
		{
			name:     "empty compound filter matches all",
			filters:  []Filter{},
			entry:    LogEntry{Message: "anything"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &CompoundFilter{Filters: tt.filters}
			got := filter.Match(tt.entry)
			if got != tt.expected {
				t.Errorf("CompoundFilter.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAnyFilter_Match(t *testing.T) {
	tests := []struct {
		name     string
		filters  []Filter
		entry    LogEntry
		expected bool
	}{
		{
			name: "one filter matches",
			filters: []Filter{
				&TextFilter{Pattern: "error"},
				&TextFilter{Pattern: "warning"},
			},
			entry:    LogEntry{Message: "a warning occurred"},
			expected: true,
		},
		{
			name: "no filter matches",
			filters: []Filter{
				&TextFilter{Pattern: "error"},
				&TextFilter{Pattern: "warning"},
			},
			entry:    LogEntry{Message: "success"},
			expected: false,
		},
		{
			name:     "empty any filter matches all",
			filters:  []Filter{},
			entry:    LogEntry{Message: "anything"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &AnyFilter{Filters: tt.filters}
			got := filter.Match(tt.entry)
			if got != tt.expected {
				t.Errorf("AnyFilter.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildFilter(t *testing.T) {
	tests := []struct {
		name     string
		opts     StreamOptions
		entry    LogEntry
		expected bool
	}{
		{
			name:     "no filters matches all",
			opts:     StreamOptions{},
			entry:    LogEntry{Message: "anything", Level: LevelDebug},
			expected: true,
		},
		{
			name:     "level filter only",
			opts:     StreamOptions{Level: LevelWarning},
			entry:    LogEntry{Message: "error", Level: LevelError},
			expected: true,
		},
		{
			name:     "text filter only",
			opts:     StreamOptions{Filter: "error"},
			entry:    LogEntry{Message: "an error occurred"},
			expected: true,
		},
		{
			name:     "combined filters",
			opts:     StreamOptions{Filter: "error", Level: LevelWarning, Tag: "Android"},
			entry:    LogEntry{Message: "an error occurred", Level: LevelError, Tag: "Android"},
			expected: true,
		},
		{
			name:     "combined filters one fails",
			opts:     StreamOptions{Filter: "error", Level: LevelWarning},
			entry:    LogEntry{Message: "an error occurred", Level: LevelDebug},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := BuildFilter(tt.opts)
			got := filter.Match(tt.entry)
			if got != tt.expected {
				t.Errorf("BuildFilter() match = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LevelVerbose, "VERBOSE"},
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarning, "WARNING"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
		{LevelSilent, "SILENT"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"V", LevelVerbose},
		{"VERBOSE", LevelVerbose},
		{"D", LevelDebug},
		{"DEBUG", LevelDebug},
		{"I", LevelInfo},
		{"INFO", LevelInfo},
		{"W", LevelWarning},
		{"WARN", LevelWarning},
		{"WARNING", LevelWarning},
		{"E", LevelError},
		{"ERR", LevelError},
		{"ERROR", LevelError},
		{"F", LevelFatal},
		{"FATAL", LevelFatal},
		{"S", LevelSilent},
		{"SILENT", LevelSilent},
		{"unknown", LevelDebug}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLogLevel(tt.input)
			if got != tt.expected {
				t.Errorf("ParseLogLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLogEntry_Timestamp(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Level:     LevelInfo,
		Message:   "Test message",
		Source:    "test",
	}

	if entry.Timestamp.Year() != 2024 {
		t.Errorf("Expected year 2024, got %d", entry.Timestamp.Year())
	}

	if entry.Source != "test" {
		t.Errorf("Expected source 'test', got %s", entry.Source)
	}
}
