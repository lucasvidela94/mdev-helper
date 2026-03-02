// Package logs provides Android log streaming functionality.
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

// AndroidStreamer streams logs from Android devices using adb logcat.
type AndroidStreamer struct {
	// adbPath is the path to the adb executable.
	adbPath string
}

// NewAndroidStreamer creates a new Android log streamer.
func NewAndroidStreamer() *AndroidStreamer {
	return &AndroidStreamer{}
}

// Name returns the name of this log source.
func (s *AndroidStreamer) Name() string {
	return "android"
}

// CheckPrerequisites verifies that adb is available.
func (s *AndroidStreamer) CheckPrerequisites() error {
	adbPath, err := exec.LookPath("adb")
	if err != nil {
		return fmt.Errorf("adb not found in PATH: %w", err)
	}
	s.adbPath = adbPath
	return nil
}

// Stream starts streaming Android logs and returns a channel of log entries.
func (s *AndroidStreamer) Stream(ctx context.Context, opts logs.StreamOptions) (<-chan logs.LogEntry, error) {
	if err := s.CheckPrerequisites(); err != nil {
		return nil, err
	}

	// Build adb logcat command
	args := s.buildLogcatArgs(opts)
	cmd := exec.CommandContext(ctx, s.adbPath, args...)

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start adb logcat: %w", err)
	}

	// Create output channel
	entries := make(chan logs.LogEntry, 100)

	// Start goroutine to parse logcat output
	go func() {
		defer close(entries)
		defer cmd.Process.Kill() // Ensure process is killed when done

		scanner := bufio.NewScanner(stdout)
		filter := logs.BuildFilter(opts)
		lineCount := 0

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			entry := s.parseLogcatLine(line)

			// Apply filters
			if !filter.Match(entry) {
				continue
			}

			// Apply package filter if specified
			if opts.Package != "" && !strings.EqualFold(entry.Package, opts.Package) {
				continue
			}

			// Apply line limit
			if opts.Lines > 0 && lineCount >= opts.Lines {
				if !opts.Follow {
					return
				}
			}

			entries <- entry
			lineCount++
		}
	}()

	return entries, nil
}

// buildLogcatArgs builds the adb logcat command arguments based on options.
func (s *AndroidStreamer) buildLogcatArgs(opts logs.StreamOptions) []string {
	args := []string{"logcat"}

	// Format: use threadtime for detailed output
	args = append(args, "-v", "threadtime")

	// Apply priority filter
	if opts.Level > logs.LevelVerbose {
		priority := s.levelToPriority(opts.Level)
		if opts.Tag != "" {
			args = append(args, fmt.Sprintf("%s:%s", opts.Tag, priority))
		} else {
			args = append(args, fmt.Sprintf("*:%s", priority))
		}
	} else if opts.Tag != "" {
		args = append(args, fmt.Sprintf("%s:D", opts.Tag))
	}

	// Clear logs if not following (for fresh output)
	if !opts.Follow {
		args = append(args, "-d") // Dump and exit
	}

	return args
}

// levelToPriority converts a LogLevel to Android logcat priority.
func (s *AndroidStreamer) levelToPriority(level logs.LogLevel) string {
	switch level {
	case logs.LevelVerbose:
		return "V"
	case logs.LevelDebug:
		return "D"
	case logs.LevelInfo:
		return "I"
	case logs.LevelWarning:
		return "W"
	case logs.LevelError:
		return "E"
	case logs.LevelFatal:
		return "F"
	default:
		return "D"
	}
}

// logcatLineRegex matches the threadtime format: "MM-DD HH:MM:SS.mmm PID TID L TAG: message"
// Example: "01-15 10:30:45.123 1234 5678 D AndroidRuntime: Message"
var logcatLineRegex = regexp.MustCompile(`^(\d{2}-\d{2})\s+(\d{2}:\d{2}:\d{2}\.\d{3})\s+(\d+)\s+(\d+)\s+([VDIWEF])\s+(.*?):\s*(.*)$`)

// parseLogcatLine parses a logcat line into a LogEntry.
func (s *AndroidStreamer) parseLogcatLine(line string) logs.LogEntry {
	entry := logs.LogEntry{
		Source: "android",
		Raw:    line,
	}

	matches := logcatLineRegex.FindStringSubmatch(line)
	if matches == nil {
		// Not a standard logcat line, treat entire line as message
		entry.Message = line
		entry.Timestamp = time.Now()
		return entry
	}

	// Parse timestamp (use current year since logcat doesn't include it)
	dateStr := fmt.Sprintf("%d-%s", time.Now().Year(), matches[1])
	timeStr := matches[2]
	timestamp, err := time.Parse("2006-01-02 15:04:05.000", dateStr+" "+timeStr)
	if err != nil {
		entry.Timestamp = time.Now()
	} else {
		entry.Timestamp = timestamp
	}

	// Parse process and thread IDs
	if pid, err := strconv.Atoi(matches[3]); err == nil {
		entry.ProcessID = pid
	}
	if tid, err := strconv.Atoi(matches[4]); err == nil {
		entry.ThreadID = tid
	}

	// Parse level
	entry.Level = s.priorityToLevel(matches[5])

	// Parse tag
	entry.Tag = strings.TrimSpace(matches[6])

	// Parse message
	entry.Message = matches[7]

	return entry
}

// priorityToLevel converts an Android logcat priority to LogLevel.
func (s *AndroidStreamer) priorityToLevel(priority string) logs.LogLevel {
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
		return logs.LevelDebug
	}
}

// GetDevices returns a list of connected Android devices.
func (s *AndroidStreamer) GetDevices() ([]string, error) {
	if err := s.CheckPrerequisites(); err != nil {
		return nil, err
	}

	cmd := exec.Command(s.adbPath, "devices", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	var devices []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // Skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			devices = append(devices, parts[0])
		}
	}

	return devices, nil
}

// ClearLogs clears the Android device logs.
func (s *AndroidStreamer) ClearLogs() error {
	if err := s.CheckPrerequisites(); err != nil {
		return err
	}

	cmd := exec.Command(s.adbPath, "logcat", "-c")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}

	return nil
}
