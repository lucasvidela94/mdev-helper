// Package logs provides export functionality for log entries.
package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportFormat represents the format for log export.
type ExportFormat string

const (
	// FormatRaw exports logs as raw text.
	FormatRaw ExportFormat = "raw"
	// FormatJSON exports logs as JSON.
	FormatJSON ExportFormat = "json"
)

// Exporter handles exporting log entries to files.
type Exporter struct {
	// Format is the export format.
	Format ExportFormat
	// OutputPath is the path to write the exported logs.
	OutputPath string
	// Pretty enables pretty-printing for JSON output.
	Pretty bool
}

// NewExporter creates a new exporter with the specified format.
func NewExporter(format ExportFormat, outputPath string) *Exporter {
	return &Exporter{
		Format:     format,
		OutputPath: outputPath,
		Pretty:     false,
	}
}

// Export exports a slice of log entries to the configured output.
func (e *Exporter) Export(entries []LogEntry) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(e.OutputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	switch e.Format {
	case FormatJSON:
		return e.exportJSON(entries)
	case FormatRaw:
		return e.exportRaw(entries)
	default:
		return fmt.Errorf("unsupported export format: %s", e.Format)
	}
}

// exportJSON exports entries as JSON.
func (e *Exporter) exportJSON(entries []LogEntry) error {
	file, err := os.Create(e.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if e.Pretty {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		data := map[string]interface{}{
			"exportedAt": time.Now().Format(time.RFC3339),
			"count":      len(entries),
			"entries":    entries,
		}
		return encoder.Encode(data)
	}

	// Compact JSON format
	data := map[string]interface{}{
		"exportedAt": time.Now().Format(time.RFC3339),
		"count":      len(entries),
		"entries":    entries,
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

// exportRaw exports entries as raw text.
func (e *Exporter) exportRaw(entries []LogEntry) error {
	file, err := os.Create(e.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	formatter := NewDefaultFormatter()
	formatter.Colorize = false // Disable colors for file output

	// Write header
	header := fmt.Sprintf("# Exported: %s\n# Count: %d\n\n",
		time.Now().Format(time.RFC3339),
		len(entries))
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write entries
	for _, entry := range entries {
		line := formatter.Format(entry) + "\n"
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("failed to write entry: %w", err)
		}
	}

	// Write footer
	footer := fmt.Sprintf("\n# End of export\n# Total entries: %d\n", len(entries))
	if _, err := file.WriteString(footer); err != nil {
		return fmt.Errorf("failed to write footer: %w", err)
	}

	return nil
}

// StreamingExporter handles real-time export during log streaming.
type StreamingExporter struct {
	// exporter is the underlying exporter configuration.
	exporter *Exporter
	// file is the open file handle.
	file *os.File
	// formatter formats entries for output.
	formatter Formatter
	// count tracks the number of exported entries.
	count int
	// firstEntry tracks if this is the first entry (for JSON array formatting).
	firstEntry bool
}

// NewStreamingExporter creates a new streaming exporter.
func NewStreamingExporter(format ExportFormat, outputPath string) (*StreamingExporter, error) {
	// Create output directory if needed
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	exporter := &StreamingExporter{
		exporter: &Exporter{
			Format:     format,
			OutputPath: outputPath,
		},
		file:       file,
		firstEntry: true,
	}

	// Initialize based on format
	switch format {
	case FormatJSON:
		exporter.formatter = nil // JSON uses custom encoding
		// Write JSON header
		header := fmt.Sprintf(`{"exportedAt":"%s","entries":[`,
			time.Now().Format(time.RFC3339))
		if _, err := file.WriteString(header); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write JSON header: %w", err)
		}
	case FormatRaw:
		formatter := NewDefaultFormatter()
		formatter.Colorize = false
		exporter.formatter = formatter
		// Write raw header
		header := fmt.Sprintf("# Exported: %s\n\n", time.Now().Format(time.RFC3339))
		if _, err := file.WriteString(header); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write header: %w", err)
		}
	default:
		file.Close()
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	return exporter, nil
}

// Write exports a single log entry during streaming.
func (s *StreamingExporter) Write(entry LogEntry) error {
	s.count++

	switch s.exporter.Format {
	case FormatJSON:
		return s.writeJSON(entry)
	case FormatRaw:
		return s.writeRaw(entry)
	default:
		return fmt.Errorf("unsupported export format: %s", s.exporter.Format)
	}
}

// writeJSON writes a single entry in JSON format.
func (s *StreamingExporter) writeJSON(entry LogEntry) error {
	// Add comma separator if not the first entry
	if !s.firstEntry {
		if _, err := s.file.WriteString(","); err != nil {
			return err
		}
	}
	s.firstEntry = false

	// Marshal entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	if _, err := s.file.Write(data); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	return nil
}

// writeRaw writes a single entry in raw format.
func (s *StreamingExporter) writeRaw(entry LogEntry) error {
	line := s.formatter.Format(entry) + "\n"
	if _, err := s.file.WriteString(line); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}
	return nil
}

// Close finalizes the export and closes the file.
func (s *StreamingExporter) Close() error {
	// Write footer based on format
	switch s.exporter.Format {
	case FormatJSON:
		footer := fmt.Sprintf(`],"count":%d}`, s.count)
		if _, err := s.file.WriteString(footer); err != nil {
			s.file.Close()
			return fmt.Errorf("failed to write JSON footer: %w", err)
		}
	case FormatRaw:
		footer := fmt.Sprintf("\n# End of export\n# Total entries: %d\n", s.count)
		if _, err := s.file.WriteString(footer); err != nil {
			s.file.Close()
			return fmt.Errorf("failed to write footer: %w", err)
		}
	}

	return s.file.Close()
}

// Count returns the number of entries exported.
func (s *StreamingExporter) Count() int {
	return s.count
}

// ParseExportFormat parses an export format string.
func ParseExportFormat(s string) (ExportFormat, error) {
	switch strings.ToLower(s) {
	case "raw", "text", "txt":
		return FormatRaw, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown export format: %s", s)
	}
}
