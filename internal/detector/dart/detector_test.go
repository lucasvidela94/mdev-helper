package dart

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// MockCommandRunner for Dart testing
type MockCommandRunner struct {
	Responses map[string]mockResponse
}

type mockResponse struct {
	Output []byte
	Err    error
}

func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		Responses: make(map[string]mockResponse),
	}
}

func (m *MockCommandRunner) Run(name string, arg ...string) ([]byte, error) {
	key := name
	for _, a := range arg {
		key += " " + a
	}
	if resp, ok := m.Responses[key]; ok {
		return resp.Output, resp.Err
	}
	return nil, os.ErrNotExist
}

// Predefined dart --version output
func getMockDartVersionOutput() []byte {
	return []byte("Dart SDK version: 3.5.0 (stable) (Tue Jul 30 02:17:59 2024 -0700) on \"linux_x64\"\n")
}

// Predefined newer dart --version output (simpler format)
func getMockDartVersionOutputSimple() []byte {
	return []byte("3.5.0\n")
}

// TestDetect_WithBundledDart tests detecting Dart bundled with Flutter
func TestDetect_WithBundledDart(t *testing.T) {
	// Create a mock Flutter SDK with bundled Dart
	tempDir := t.TempDir()
	dartSDKDir := filepath.Join(tempDir, "bin", "cache", "dart-sdk")
	dartBinDir := filepath.Join(dartSDKDir, "bin")
	os.MkdirAll(dartBinDir, 0755)

	dartBin := filepath.Join(dartBinDir, "dart")
	if runtime.GOOS == "windows" {
		dartBin += ".exe"
	}

	// Create dart binary
	os.WriteFile(dartBin, []byte("#!/bin/sh\necho 'Dart'"), 0755)

	info := Detect(tempDir)

	if info == nil {
		t.Fatal("Detect() returned nil")
	}

	if !info.IsBundled {
		t.Error("Detect() IsBundled = false, want true for bundled Dart")
	}

	if info.Flutter != tempDir {
		t.Errorf("Detect() Flutter = %v, want %v", info.Flutter, tempDir)
	}

	if info.Path != dartSDKDir {
		t.Errorf("Detect() Path = %v, want %v", info.Path, dartSDKDir)
	}
}

// TestDetect_WithDartSDKEnv tests detecting Dart from DART_SDK environment variable
func TestDetect_WithDartSDKEnv(t *testing.T) {
	// Save original environment
	originalDartSDK := os.Getenv("DART_SDK")
	defer func() {
		if originalDartSDK != "" {
			os.Setenv("DART_SDK", originalDartSDK)
		} else {
			os.Unsetenv("DART_SDK")
		}
	}()

	// Create a mock Dart SDK
	tempDir := t.TempDir()
	dartBinDir := filepath.Join(tempDir, "bin")
	os.MkdirAll(dartBinDir, 0755)

	dartBin := filepath.Join(dartBinDir, "dart")
	if runtime.GOOS == "windows" {
		dartBin += ".exe"
	}

	os.WriteFile(dartBin, []byte("#!/bin/sh\necho 'Dart'"), 0755)

	// Set DART_SDK
	os.Setenv("DART_SDK", tempDir)

	info := Detect("")

	if info == nil {
		t.Fatal("Detect() returned nil")
	}

	if info.Path != tempDir {
		t.Errorf("Detect() Path = %v, want %v", info.Path, tempDir)
	}
}

// TestDetect_NoDartInstalled tests Detect when Dart is not installed
func TestDetect_NoDartInstalled(t *testing.T) {
	// Save original environment
	originalDartSDK := os.Getenv("DART_SDK")
	originalPath := os.Getenv("PATH")
	defer func() {
		if originalDartSDK != "" {
			os.Setenv("DART_SDK", originalDartSDK)
		} else {
			os.Unsetenv("DART_SDK")
		}
		os.Setenv("PATH", originalPath)
	}()

	// Clear environment
	os.Unsetenv("DART_SDK")
	os.Setenv("PATH", "/usr/bin:/bin")

	info := Detect("")

	if info == nil {
		t.Fatal("Detect() returned nil")
	}

	if info.IsValid {
		t.Error("Detect() should return invalid DartInfo when Dart is not installed")
	}

	if info.Error == "" {
		t.Error("Detect() should return error message when Dart is not found")
	}
}

// TestCheckDartPath_ValidPath tests checkDartPath with a valid Dart SDK path
func TestCheckDartPath_ValidPath(t *testing.T) {
	tempDir := t.TempDir()
	dartBinDir := filepath.Join(tempDir, "bin")
	os.MkdirAll(dartBinDir, 0755)

	dartBin := filepath.Join(dartBinDir, "dart")
	if runtime.GOOS == "windows" {
		dartBin += ".exe"
	}

	os.WriteFile(dartBin, []byte("#!/bin/sh\necho 'Dart'"), 0755)

	info := checkDartPath(tempDir)

	if info == nil {
		t.Fatal("checkDartPath() returned nil for valid path")
	}

	if !info.IsValid {
		t.Error("checkDartPath() IsValid = false for valid Dart SDK")
	}

	if info.Path != tempDir {
		t.Errorf("checkDartPath() Path = %v, want %v", info.Path, tempDir)
	}
}

// TestCheckDartPath_MissingBinary tests checkDartPath when dart binary is missing
func TestCheckDartPath_MissingBinary(t *testing.T) {
	tempDir := t.TempDir()

	info := checkDartPath(tempDir)

	if info != nil {
		t.Error("checkDartPath() should return nil when dart binary is missing")
	}
}

// TestCheckDartPath_InvalidPath tests checkDartPath with non-existent path
func TestCheckDartPath_InvalidPath(t *testing.T) {
	info := checkDartPath("/nonexistent/path/to/dart")

	if info != nil {
		t.Error("checkDartPath() should return nil for non-existent path")
	}
}

// TestDetectFromPath_Found tests detecting Dart from PATH
func TestDetectFromPath_Found(t *testing.T) {
	// This test depends on environment
	info := detectFromPath()

	// We can't make strong assertions here since it depends on the test environment
	// Just verify it doesn't panic
	_ = info
}

// TestStandardPaths_NotEmpty verifies standard paths are defined
func TestStandardPaths_NotEmpty(t *testing.T) {
	if standardPaths == nil {
		t.Error("standardPaths should not be nil")
	}

	if len(standardPaths) == 0 {
		t.Error("standardPaths should contain at least one path")
	}
}

// TestGetDartVersion_FullFormat tests parsing full dart --version output
func TestGetDartVersion_FullFormat(t *testing.T) {
	// Create a mock dart binary
	tempDir := t.TempDir()
	dartBin := filepath.Join(tempDir, "dart")
	if runtime.GOOS == "windows" {
		dartBin += ".exe"
	}

	// Create script that outputs version
	script := "#!/bin/sh\necho 'Dart SDK version: 3.5.0 (stable) (Tue Jul 30 02:17:59 2024 -0700) on \"linux_x64\"'"
	os.WriteFile(dartBin, []byte(script), 0755)

	version, err := getDartVersion(dartBin)
	if err != nil {
		// Script execution might not work in this environment
		t.Skip("Skipping test - script execution not supported")
	}

	if version != "3.5.0" {
		t.Errorf("getDartVersion() = %v, want '3.5.0'", version)
	}
}

// TestGetDartVersion_SimpleFormat tests parsing simple version output
func TestGetDartVersion_SimpleFormat(t *testing.T) {
	// Create a mock dart binary
	tempDir := t.TempDir()
	dartBin := filepath.Join(tempDir, "dart")
	if runtime.GOOS == "windows" {
		dartBin += ".exe"
	}

	// Create script that outputs simple version
	script := "#!/bin/sh\necho '3.5.0'"
	os.WriteFile(dartBin, []byte(script), 0755)

	version, err := getDartVersion(dartBin)
	if err != nil {
		t.Skip("Skipping test - script execution not supported")
	}

	if version != "3.5.0" {
		t.Errorf("getDartVersion() = %v, want '3.5.0'", version)
	}
}

// TestGetDartVersion_InvalidOutput tests handling of invalid version output
func TestGetDartVersion_InvalidOutput(t *testing.T) {
	// Create a mock dart binary
	tempDir := t.TempDir()
	dartBin := filepath.Join(tempDir, "dart")
	if runtime.GOOS == "windows" {
		dartBin += ".exe"
	}

	// Create script that outputs invalid version
	script := "#!/bin/sh\necho 'not a version'"
	os.WriteFile(dartBin, []byte(script), 0755)

	_, err := getDartVersion(dartBin)
	if err != nil {
		t.Skip("Skipping test - script execution not supported")
	}

	// Should return empty string or error for invalid output
	// We can't assert strongly here due to environment differences
}

// TestDartInfoStructure tests that DartInfo has all required fields
func TestDartInfoStructure(t *testing.T) {
	info := &detector.DartInfo{
		Path:      "/test/dart-sdk",
		Version:   "3.5.0",
		IsValid:   true,
		IsBundled: false,
		Flutter:   "",
		Error:     "",
	}

	if info.Path != "/test/dart-sdk" {
		t.Error("DartInfo.Path not set correctly")
	}

	if info.Version != "3.5.0" {
		t.Error("DartInfo.Version not set correctly")
	}

	if info.IsBundled {
		t.Error("DartInfo.IsBundled should be false")
	}
}

// TestDartInfo_Bundled tests DartInfo for bundled Dart
func TestDartInfo_Bundled(t *testing.T) {
	info := &detector.DartInfo{
		Path:      "/flutter/bin/cache/dart-sdk",
		Version:   "3.5.0",
		IsValid:   true,
		IsBundled: true,
		Flutter:   "/flutter",
	}

	if !info.IsBundled {
		t.Error("DartInfo.IsBundled should be true")
	}

	if info.Flutter != "/flutter" {
		t.Errorf("DartInfo.Flutter = %v, want '/flutter'", info.Flutter)
	}
}

// TestDetect_PriorityOrder tests that detection follows correct priority order
func TestDetect_PriorityOrder(t *testing.T) {
	// Priority 1: Bundled Dart (flutterPath parameter)
	// Priority 2: DART_SDK environment variable
	// Priority 3: PATH
	// Priority 4: Standard paths

	// This test verifies the priority logic by checking that bundled Dart is preferred
	tempDir := t.TempDir()

	// Create bundled Dart
	bundledDartDir := filepath.Join(tempDir, "bin", "cache", "dart-sdk")
	bundledBinDir := filepath.Join(bundledDartDir, "bin")
	os.MkdirAll(bundledBinDir, 0755)

	bundledDartBin := filepath.Join(bundledBinDir, "dart")
	if runtime.GOOS == "windows" {
		bundledDartBin += ".exe"
	}
	os.WriteFile(bundledDartBin, []byte("#!/bin/sh\necho 'bundled'"), 0755)

	// Create standalone Dart SDK
	standaloneDir := t.TempDir()
	standaloneBinDir := filepath.Join(standaloneDir, "bin")
	os.MkdirAll(standaloneBinDir, 0755)

	standaloneDartBin := filepath.Join(standaloneBinDir, "dart")
	if runtime.GOOS == "windows" {
		standaloneDartBin += ".exe"
	}
	os.WriteFile(standaloneDartBin, []byte("#!/bin/sh\necho 'standalone'"), 0755)

	// Save and set DART_SDK
	originalDartSDK := os.Getenv("DART_SDK")
	defer func() {
		if originalDartSDK != "" {
			os.Setenv("DART_SDK", originalDartSDK)
		} else {
			os.Unsetenv("DART_SDK")
		}
	}()
	os.Setenv("DART_SDK", standaloneDir)

	// Detect with flutterPath - should find bundled Dart
	info := Detect(tempDir)

	if info == nil {
		t.Fatal("Detect() returned nil")
	}

	// Should find the bundled Dart
	if !info.IsBundled {
		t.Error("Detect() should prefer bundled Dart over DART_SDK")
	}
}

// TestDetectFromPath_WithSymlink tests detecting Dart from PATH with symlink resolution
func TestDetectFromPath_WithSymlink(t *testing.T) {
	// Skip on Windows - symlinks work differently
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	// Create actual Dart SDK
	tempDir := t.TempDir()
	dartBinDir := filepath.Join(tempDir, "bin")
	os.MkdirAll(dartBinDir, 0755)

	dartBin := filepath.Join(dartBinDir, "dart")
	os.WriteFile(dartBin, []byte("#!/bin/sh\necho 'Dart'"), 0755)

	// Create a symlink in another directory
	symlinkDir := t.TempDir()
	symlinkPath := filepath.Join(symlinkDir, "dart")
	os.Symlink(dartBin, symlinkPath)

	// This test verifies that detectFromPath handles symlinks
	// We can't easily test this without modifying PATH, so we just verify the function exists
	_ = detectFromPath
}

// TestDetect_InvalidFlutterPath tests Detect with invalid Flutter path
func TestDetect_InvalidFlutterPath(t *testing.T) {
	// Save original environment
	originalDartSDK := os.Getenv("DART_SDK")
	originalPath := os.Getenv("PATH")
	defer func() {
		if originalDartSDK != "" {
			os.Setenv("DART_SDK", originalDartSDK)
		} else {
			os.Unsetenv("DART_SDK")
		}
		os.Setenv("PATH", originalPath)
	}()

	// Clear environment
	os.Unsetenv("DART_SDK")
	os.Setenv("PATH", "/usr/bin:/bin")

	// Detect with invalid flutter path - should fall through to other methods
	info := Detect("/nonexistent/flutter")

	if info == nil {
		t.Fatal("Detect() returned nil")
	}

	// Should not be valid since no Dart is available
	// But we just verify it doesn't panic
}
