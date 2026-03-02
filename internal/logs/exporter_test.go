package logs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseExportFormat(t *testing.T) {
	tests := []struct {
		input       string
		expected    ExportFormat
		shouldError bool
	}{
		{"raw", FormatRaw, false},
		{"text", FormatRaw, false},
		{"txt", FormatRaw, false},
		{"json", FormatJSON, false},
		{"JSON", FormatJSON, false},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseExportFormat(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Errorf("ParseExportFormat() expected error for input %q", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseExportFormat() unexpected error: %v", err)
				}
				if got != tt.expected {
					t.Errorf("ParseExportFormat() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestExporter_Export_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.json")

	entries := []LogEntry{
		{
			Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Level:     LevelInfo,
			Message:   "Test message",
			Source:    "android",
		},
	}

	exporter := NewExporter(FormatJSON, outputPath)
	err := exporter.Export(entries)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Export file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["count"] != float64(1) {
		t.Errorf("Expected count 1, got %v", result["count"])
	}

	if _, ok := result["exportedAt"]; !ok {
		t.Error("Expected exportedAt field in JSON")
	}

	if _, ok := result["entries"]; !ok {
		t.Error("Expected entries field in JSON")
	}
}

func TestExporter_Export_Raw(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.log")

	entries := []LogEntry{
		{
			Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Level:     LevelInfo,
			Message:   "Test message",
			Source:    "android",
		},
	}

	exporter := NewExporter(FormatRaw, outputPath)
	err := exporter.Export(entries)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# Exported:") {
		t.Error("Expected header in raw export")
	}
	if !strings.Contains(contentStr, "Test message") {
		t.Error("Expected message in raw export")
	}
	if !strings.Contains(contentStr, "# End of export") {
		t.Error("Expected footer in raw export")
	}
}

func TestExporter_Export_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.xml")

	entries := []LogEntry{
		{Message: "Test"},
	}

	exporter := NewExporter(ExportFormat("xml"), outputPath)
	err := exporter.Export(entries)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestExporter_Export_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "dir")
	outputPath := filepath.Join(nestedDir, "test.json")

	entries := []LogEntry{
		{Message: "Test"},
	}

	exporter := NewExporter(FormatJSON, outputPath)
	err := exporter.Export(entries)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("Nested directory was not created")
	}
}

func TestStreamingExporter_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "stream.json")

	exporter, err := NewStreamingExporter(FormatJSON, outputPath)
	if err != nil {
		t.Fatalf("NewStreamingExporter() error = %v", err)
	}

	entries := []LogEntry{
		{Timestamp: time.Now(), Level: LevelInfo, Message: "First", Source: "test"},
		{Timestamp: time.Now(), Level: LevelWarning, Message: "Second", Source: "test"},
	}

	for _, entry := range entries {
		if err := exporter.Write(entry); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}

	if err := exporter.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Verify count
	if exporter.Count() != 2 {
		t.Errorf("Expected count 2, got %d", exporter.Count())
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	contentStr := string(content)
	if !strings.HasPrefix(contentStr, `{"exportedAt":`) {
		t.Error("Expected JSON object start")
	}
	if !strings.Contains(contentStr, `"entries":[`) {
		t.Error("Expected entries array")
	}
	if !strings.HasSuffix(strings.TrimSpace(contentStr), `],"count":2}`) {
		t.Errorf("Expected proper JSON ending, got: %s", contentStr[len(contentStr)-50:])
	}
}

func TestStreamingExporter_Raw(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "stream.log")

	exporter, err := NewStreamingExporter(FormatRaw, outputPath)
	if err != nil {
		t.Fatalf("NewStreamingExporter() error = %v", err)
	}

	entry := LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Level:     LevelInfo,
		Message:   "Test message",
		Source:    "android",
	}

	if err := exporter.Write(entry); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := exporter.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# Exported:") {
		t.Error("Expected header in raw export")
	}
	if !strings.Contains(contentStr, "Test message") {
		t.Error("Expected message in raw export")
	}
	if !strings.Contains(contentStr, "# Total entries: 1") {
		t.Error("Expected footer with count")
	}
}

func TestStreamingExporter_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.xml")

	_, err := NewStreamingExporter(ExportFormat("xml"), outputPath)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestExporter_JSON_Pretty(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "pretty.json")

	entries := []LogEntry{
		{
			Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Level:     LevelInfo,
			Message:   "Test message",
			Source:    "android",
		},
	}

	exporter := NewExporter(FormatJSON, outputPath)
	exporter.Pretty = true
	err := exporter.Export(entries)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Read and verify content has indentation
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "  \"exportedAt\"") {
		t.Error("Expected indented JSON output")
	}
}
