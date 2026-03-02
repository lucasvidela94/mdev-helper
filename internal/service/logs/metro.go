// Package logs provides Metro bundler log streaming functionality.
package logs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sombi/mobile-dev-helper/internal/logs"
)

// MetroStreamer streams logs from Metro bundler.
type MetroStreamer struct {
	// projectPath is the path to the React Native project.
	projectPath string
}

// NewMetroStreamer creates a new Metro log streamer.
func NewMetroStreamer(projectPath string) *MetroStreamer {
	return &MetroStreamer{
		projectPath: projectPath,
	}
}

// Name returns the name of this log source.
func (s *MetroStreamer) Name() string {
	return "metro"
}

// CheckPrerequisites verifies that this is a React Native project and Metro is available.
func (s *MetroStreamer) CheckPrerequisites() error {
	if s.projectPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine current directory: %w", err)
		}
		s.projectPath = wd
	}

	// Check for package.json
	packageJSONPath := filepath.Join(s.projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); os.IsNotExist(err) {
		return fmt.Errorf("no package.json found in %s", s.projectPath)
	}

	// Check for React Native in package.json
	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	if !strings.Contains(string(content), "react-native") {
		return fmt.Errorf("no react-native dependency found in package.json")
	}

	return nil
}

// IsReactNativeProject checks if the current directory is a React Native project.
func (s *MetroStreamer) IsReactNativeProject() bool {
	if s.projectPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return false
		}
		s.projectPath = wd
	}

	// Check for React Native specific files
	rnFiles := []string{
		"metro.config.js",
		"metro.config.ts",
		"react-native.config.js",
	}

	for _, file := range rnFiles {
		if _, err := os.Stat(filepath.Join(s.projectPath, file)); err == nil {
			return true
		}
	}

	// Check package.json for react-native
	packageJSONPath := filepath.Join(s.projectPath, "package.json")
	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "react-native")
}

// Stream starts streaming Metro logs.
// Note: This reads from Metro's log file or stdout if Metro is running.
func (s *MetroStreamer) Stream(ctx context.Context, opts logs.StreamOptions) (<-chan logs.LogEntry, error) {
	if err := s.CheckPrerequisites(); err != nil {
		return nil, err
	}

	entries := make(chan logs.LogEntry, 100)

	// Metro logs are typically read from the terminal where Metro is running.
	// For now, we'll create a mock streamer that can be extended to read from
	// Metro's log file or connect to the Metro server.
	go func() {
		defer close(entries)

		// This is a placeholder implementation.
		// In a real implementation, you would:
		// 1. Connect to Metro's websocket or log file
		// 2. Parse Metro's JSON log format
		// 3. Stream entries to the channel

		// For now, just send a notification that Metro streaming isn't fully implemented
		select {
		case <-ctx.Done():
			return
		case entries <- logs.LogEntry{
			Timestamp: time.Now(),
			Level:     logs.LevelWarning,
			Message:   "Metro log streaming requires Metro to be running. Start Metro with 'npx react-native start' in another terminal.",
			Source:    "metro",
			Tag:       "Metro",
		}:
		}
	}()

	return entries, nil
}

// metroLogRegex matches Metro log format (simplified)
// Metro typically outputs JSON or formatted text
var metroLogRegex = regexp.MustCompile(`^\[(\w+)\]\s*(.*)$`)

// parseMetroLine parses a Metro log line into a LogEntry.
func (s *MetroStreamer) parseMetroLine(line string) logs.LogEntry {
	entry := logs.LogEntry{
		Source:    "metro",
		Raw:       line,
		Timestamp: time.Now(),
		Tag:       "Metro",
	}

	// Try to parse Metro's log format
	matches := metroLogRegex.FindStringSubmatch(line)
	if matches != nil {
		levelStr := matches[1]
		entry.Level = s.parseLevel(levelStr)
		entry.Message = matches[2]
	} else {
		entry.Level = logs.LevelInfo
		entry.Message = line
	}

	return entry
}

// parseLevel converts a Metro log level string to LogLevel.
func (s *MetroStreamer) parseLevel(level string) logs.LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG", "LOG":
		return logs.LevelDebug
	case "INFO":
		return logs.LevelInfo
	case "WARN", "WARNING":
		return logs.LevelWarning
	case "ERROR":
		return logs.LevelError
	default:
		return logs.LevelInfo
	}
}

// GetMetroStatus returns the status of Metro bundler.
func (s *MetroStreamer) GetMetroStatus() (bool, string) {
	// Check if Metro is running by looking for the process
	// This is a simplified check - in production you'd check if the Metro server is responding
	return false, "Metro status check not implemented"
}
