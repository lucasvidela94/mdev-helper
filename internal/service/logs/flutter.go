// Package logs provides Flutter log streaming functionality.
package logs

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sombi/mobile-dev-helper/internal/logs"
)

// FlutterStreamer streams logs from Flutter devices using flutter logs.
type FlutterStreamer struct {
	// flutterPath is the path to the flutter executable.
	flutterPath string
}

// NewFlutterStreamer creates a new Flutter log streamer.
func NewFlutterStreamer() *FlutterStreamer {
	return &FlutterStreamer{}
}

// Name returns the name of this log source.
func (s *FlutterStreamer) Name() string {
	return "flutter"
}

// CheckPrerequisites verifies that Flutter SDK is available.
func (s *FlutterStreamer) CheckPrerequisites() error {
	flutterPath, err := exec.LookPath("flutter")
	if err != nil {
		return fmt.Errorf("flutter not found in PATH: %w", err)
	}
	s.flutterPath = flutterPath
	return nil
}

// Stream starts streaming Flutter logs and returns a channel of log entries.
func (s *FlutterStreamer) Stream(ctx context.Context, opts logs.StreamOptions) (<-chan logs.LogEntry, error) {
	if err := s.CheckPrerequisites(); err != nil {
		return nil, err
	}

	// Build flutter logs command
	args := []string{"logs"}

	// Add device if specified
	if opts.Package != "" {
		// In Flutter, --device-id specifies the device
		args = append(args, "--device-id", opts.Package)
	}

	cmd := exec.CommandContext(ctx, s.flutterPath, args...)

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start flutter logs: %w", err)
	}

	// Create output channel
	entries := make(chan logs.LogEntry, 100)

	// Start goroutine to parse flutter output
	go func() {
		defer close(entries)
		defer cmd.Process.Kill() // Ensure process is killed when done

		scanner := bufio.NewScanner(stdout)
		filter := logs.BuildFilter(opts)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			entry := s.parseFlutterLine(line)

			// Apply filters
			if !filter.Match(entry) {
				continue
			}

			entries <- entry
		}
	}()

	return entries, nil
}

// flutterLogRegex matches Flutter log format.
// Example: "I/flutter (1234): Message"
// Or: "[ +123 ms] I/flutter (1234): Message"
var flutterLogRegex = regexp.MustCompile(`^(?:\[\s*[\+\-]\d+\s*ms\]\s*)?([VDIWEF])/(\w+)\s*\((\d+)\):\s*(.*)$`)

// parseFlutterLine parses a Flutter log line into a LogEntry.
func (s *FlutterStreamer) parseFlutterLine(line string) logs.LogEntry {
	entry := logs.LogEntry{
		Source: "flutter",
		Raw:    line,
	}

	matches := flutterLogRegex.FindStringSubmatch(line)
	if matches == nil {
		// Not a standard Flutter log line
		entry.Message = line
		entry.Timestamp = time.Now()
		entry.Level = logs.LevelInfo
		return entry
	}

	// Parse level
	entry.Level = s.priorityToLevel(matches[1])

	// Parse tag (usually "flutter" or "DartVM")
	entry.Tag = matches[2]

	// Parse process ID
	if pid, err := strconv.Atoi(matches[3]); err == nil {
		entry.ProcessID = pid
	}

	// Parse message
	entry.Message = matches[4]
	entry.Timestamp = time.Now()

	return entry
}

// priorityToLevel converts a Flutter log priority to LogLevel.
func (s *FlutterStreamer) priorityToLevel(priority string) logs.LogLevel {
	switch priority {
	case "V":
		return logs.LevelVerbose
	case "D":
		return logs.LevelDebug
	case "I":
		return logs.LevelInfo
	case "W":
		return logs.LevelWarning
	case "E":
		return logs.LevelError
	case "F":
		return logs.LevelFatal
	default:
		return logs.LevelInfo
	}
}

// GetDevices returns a list of connected Flutter devices.
func (s *FlutterStreamer) GetDevices() ([]string, error) {
	if err := s.CheckPrerequisites(); err != nil {
		return nil, err
	}

	cmd := exec.Command(s.flutterPath, "devices", "--machine")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Parse JSON output (simplified)
	// In a real implementation, you'd properly parse the JSON
	var devices []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.HasPrefix(line, "[") {
			// This is a simplified check - real implementation would parse JSON
			devices = append(devices, "device")
		}
	}

	return devices, nil
}

// ClearLogs clears the Flutter device logs.
func (s *FlutterStreamer) ClearLogs() error {
	if err := s.CheckPrerequisites(); err != nil {
		return err
	}

	cmd := exec.Command(s.flutterPath, "logs", "--clear")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}

	return nil
}
