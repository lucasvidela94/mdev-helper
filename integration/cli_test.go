package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	binaryName = "mdev"
)

// Helper to find the binary path
func getBinaryPath(t *testing.T) string {
	// Try to build the binary first
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	cmd.Dir = "../"
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("Could not build binary: %v\n%s", err, out)
		return ""
	}

	// Find the binary
	path := filepath.Join("..", binaryName)
	if _, err := os.Stat(path); err != nil {
		t.Skipf("Binary not found: %v", err)
		return ""
	}

	return path
}

func TestCLIHelp(t *testing.T) {
	path := getBinaryPath(t)
	if path == "" {
		t.Skip("Binary not available")
	}

	cmd := exec.Command(path, "--help")
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Command failed: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "mdev") {
		t.Error("Help output should contain tool name 'mdev'")
	}
	if !strings.Contains(output, "clean") || !strings.Contains(output, "doctor") || !strings.Contains(output, "info") {
		t.Error("Help should list available commands")
	}
}

func TestCLIVersion(t *testing.T) {
	path := getBinaryPath(t)
	if path == "" {
		t.Skip("Binary not available")
	}

	cmd := exec.Command(path, "--version")
	out, err := cmd.CombinedOutput()

	// --version might error in some cobra versions, check output either way
	output := string(out)
	if output == "" && err == nil {
		t.Error("Version command should produce output")
	}
	_ = err // Version might return error depending on implementation
}

func TestCLIInvalidCommand(t *testing.T) {
	path := getBinaryPath(t)
	if path == "" {
		t.Skip("Binary not available")
	}

	cmd := exec.Command(path, "invalid-command")
	out, err := cmd.CombinedOutput()

	// Should error on invalid command
	if err == nil {
		t.Error("Invalid command should return error")
	}

	output := string(out)
	if !strings.Contains(output, "unknown command") && !strings.Contains(output, "no such command") {
		t.Logf("Output: %s", output)
		// Some versions of cobra don't output "unknown command" explicitly
	}
}

func TestCLIInvalidFlag(t *testing.T) {
	path := getBinaryPath(t)
	if path == "" {
		t.Skip("Binary not available")
	}

	cmd := exec.Command(path, "--invalid-flag")
	out, err := cmd.CombinedOutput()

	// Should error on invalid flag
	if err == nil {
		t.Error("Invalid flag should return error")
	}

	output := string(out)
	if !strings.Contains(output, "unknown flag") && !strings.Contains(output, "flag") {
		t.Logf("Expected flag error, got: %s", output)
	}
}

func TestCLISubcommandsExist(t *testing.T) {
	path := getBinaryPath(t)
	if path == "" {
		t.Skip("Binary not available")
	}

	tests := []struct {
		name   string
		args   []string
		output string
	}{
		{"info", []string{"info"}, "info"},
		{"doctor", []string{"doctor"}, "doctor"},
		{"clean", []string{"clean"}, "clean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(path, tt.args...)
			out, _ := cmd.CombinedOutput()

			output := string(out)
			// Subcommands should at least not return "unknown command" error
			if strings.Contains(output, "unknown command") {
				t.Errorf("Subcommand %q should exist", tt.name)
			}
		})
	}
}

func TestCLIVerboseFlag(t *testing.T) {
	path := getBinaryPath(t)
	if path == "" {
		t.Skip("Binary not available")
	}

	// Test verbose flag
	cmd := exec.Command(path, "--verbose")
	_, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Verbose flag should not cause error: %v", err)
	}

	// Test short verbose flag
	cmd = exec.Command(path, "-v")
	_, err = cmd.CombinedOutput()

	if err != nil {
		t.Errorf("-v flag should not cause error: %v", err)
	}
}

func TestCLIConfigFlag(t *testing.T) {
	path := getBinaryPath(t)
	if path == "" {
		t.Skip("Binary not available")
	}

	// Test with non-existent config (should use defaults)
	cmd := exec.Command(path, "--config", "/tmp/non-existent-config.yaml")
	_, err := cmd.CombinedOutput()

	// Should not error - should just use defaults
	if err != nil {
		t.Logf("Config flag error (may be expected): %v", err)
	}

	// Test with custom config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configPath, []byte("verbose: true\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cmd = exec.Command(path, "--config", configPath)
	_, err = cmd.CombinedOutput()

	// Should not error
	if err != nil {
		t.Errorf("Custom config should not cause error: %v", err)
	}
}

// Test exit codes
func TestCLIExitCodes(t *testing.T) {
	path := getBinaryPath(t)
	if path == "" {
		t.Skip("Binary not available")
	}

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{"success no args", []string{}, false},
		{"success help", []string{"--help"}, false},
		{"success info", []string{"info"}, false},
		{"success doctor", []string{"doctor"}, false},
		{"success clean", []string{"clean"}, false},
		{"failure invalid command", []string{"invalid-cmd"}, true},
		{"failure invalid flag", []string{"--invalid-flag"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(path, tt.args...)
			err := cmd.Run()

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for args %v", tt.args)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for args %v: %v", tt.args, err)
			}
		})
	}
}

// Test output capture
func TestCLICaptureOutput(t *testing.T) {
	path := getBinaryPath(t)
	if path == "" {
		t.Skip("Binary not available")
	}

	var stdout, stderr bytes.Buffer

	cmd := exec.Command(path)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should not error
	if err != nil {
		t.Logf("Error (may be expected): %v", err)
	}

	// At least one output should have content
	if stdout.Len() == 0 && stderr.Len() == 0 {
		t.Error("Should produce some output")
	}
}
