// Package emulator provides Android emulator detection functionality.
package emulator

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// Detector provides Android emulator detection capabilities.
type Detector struct {
	sdkPath string
}

// NewDetector creates a new emulator detector.
// The sdkPath parameter is the path to the Android SDK.
func NewDetector(sdkPath string) *Detector {
	return &Detector{
		sdkPath: sdkPath,
	}
}

// Detect performs full emulator detection including binary, AVDs, and running emulators.
func (d *Detector) Detect() *detector.EmulatorInfo {
	info := &detector.EmulatorInfo{
		IsValid: true,
	}

	// Detect emulator binary
	binaryPath, version, err := d.detectBinary()
	if err != nil {
		info.IsValid = false
		info.Error = err.Error()
		return info
	}
	info.BinaryPath = binaryPath
	info.Version = version

	// List AVDs
	avds, err := d.ListAVDs()
	if err != nil {
		// AVD listing failure is not fatal
		info.AVDs = []detector.AVDInfo{}
	} else {
		info.AVDs = avds
	}

	// List running emulators
	running, err := d.ListRunning()
	if err != nil {
		// Running detection failure is not fatal
		info.Running = []detector.RunningEmulatorInfo{}
	} else {
		info.Running = running
	}

	return info
}

// detectBinary locates the emulator binary and gets its version.
func (d *Detector) detectBinary() (string, string, error) {
	if d.sdkPath == "" {
		return "", "", fmt.Errorf("SDK path not provided")
	}

	emulatorPath := filepath.Join(d.sdkPath, "emulator", "emulator")
	if runtime.GOOS == "windows" {
		emulatorPath += ".exe"
	}

	// Check if binary exists and is executable
	if _, err := os.Stat(emulatorPath); err != nil {
		return "", "", fmt.Errorf("emulator binary not found at %s", emulatorPath)
	}

	// Get version
	version, err := d.getEmulatorVersion(emulatorPath)
	if err != nil {
		// Version detection failure is not fatal
		version = "unknown"
	}

	return emulatorPath, version, nil
}

// getEmulatorVersion runs `emulator -version` and parses the output.
func (d *Detector) getEmulatorVersion(emulatorPath string) (string, error) {
	cmd := exec.Command(emulatorPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from first line
	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no version output")
	}

	firstLine := strings.TrimSpace(lines[0])
	// Format is typically: "Android emulator version 33.1.24.0"
	if strings.Contains(firstLine, "version") {
		parts := strings.Fields(firstLine)
		for i, part := range parts {
			if part == "version" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
	}

	return firstLine, nil
}

// ListAVDs discovers all AVDs in the SDK avd directory.
func (d *Detector) ListAVDs() ([]detector.AVDInfo, error) {
	if d.sdkPath == "" {
		return nil, fmt.Errorf("SDK path not provided")
	}

	avdDir := filepath.Join(d.sdkPath, ".android", "avd")

	// Also check in home directory as fallback
	home := os.Getenv("HOME")
	if home != "" {
		altAvdDir := filepath.Join(home, ".android", "avd")
		if _, err := os.Stat(altAvdDir); err == nil {
			avdDir = altAvdDir
		}
	}

	entries, err := os.ReadDir(avdDir)
	if err != nil {
		return []detector.AVDInfo{}, nil // No AVDs is not an error
	}

	var avds []detector.AVDInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// AVD directories end with .avd
		if !strings.HasSuffix(name, ".avd") {
			continue
		}

		// Remove .avd suffix to get AVD name
		avdName := strings.TrimSuffix(name, ".avd")
		avdPath := filepath.Join(avdDir, name)

		// Parse config.ini for details
		configPath := filepath.Join(avdPath, "config.ini")
		avdInfo := parseAVDConfig(avdName, avdPath, configPath)
		avds = append(avds, avdInfo)
	}

	return avds, nil
}

// parseAVDConfig reads AVD metadata from config.ini.
func parseAVDConfig(name, avdPath, configPath string) detector.AVDInfo {
	info := detector.AVDInfo{
		Name:    name,
		Path:    avdPath,
		IsValid: true,
	}

	file, err := os.Open(configPath)
	if err != nil {
		info.IsValid = false
		info.Error = fmt.Sprintf("cannot read config.ini: %v", err)
		return info
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "target":
			info.TargetAPI = value
		case "hw.device.name":
			info.Device = value
		case "abi.type":
			info.Arch = value
		}
	}

	return info
}

// ListRunning detects running emulators via adb.
func (d *Detector) ListRunning() ([]detector.RunningEmulatorInfo, error) {
	// Find adb binary
	adbPath := d.findAdb()
	if adbPath == "" {
		return nil, fmt.Errorf("adb not found")
	}

	// Run adb devices -l
	cmd := exec.Command(adbPath, "devices", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run adb: %v", err)
	}

	return parseAdbDevices(string(output)), nil
}

// findAdb locates the adb binary.
func (d *Detector) findAdb() string {
	// Priority 1: Check in SDK platform-tools
	if d.sdkPath != "" {
		adbPath := filepath.Join(d.sdkPath, "platform-tools", "adb")
		if runtime.GOOS == "windows" {
			adbPath += ".exe"
		}
		if _, err := os.Stat(adbPath); err == nil {
			return adbPath
		}
	}

	// Priority 2: Check PATH
	if adbPath, err := exec.LookPath("adb"); err == nil {
		return adbPath
	}

	return ""
}

// parseAdbDevices parses the output of `adb devices -l`.
func parseAdbDevices(output string) []detector.RunningEmulatorInfo {
	var devices []detector.RunningEmulatorInfo

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip header and empty lines
		if line == "" || strings.HasPrefix(line, "List of") {
			continue
		}

		device := parseAdbDeviceLine(line)
		if device.DeviceID != "" {
			devices = append(devices, device)
		}
	}

	return devices
}

// parseAdbDeviceLine parses a single line from adb devices output.
func parseAdbDeviceLine(line string) detector.RunningEmulatorInfo {
	device := detector.RunningEmulatorInfo{}

	// Format: emulator-5554 device product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:emulator64_x86_64_arm64
	// Or: emulator-5554 offline
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return device
	}

	device.DeviceID = fields[0]
	device.Status = fields[1]

	// Parse additional properties
	for i := 2; i < len(fields); i++ {
		parts := strings.SplitN(fields[i], ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "product":
			device.Product = value
		case "model":
			device.Model = value
		case "device":
			// This might help identify the AVD
			if device.AVDName == "" {
				device.AVDName = value
			}
		}
	}

	return device
}
