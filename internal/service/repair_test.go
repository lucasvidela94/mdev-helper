package service

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindGradlewInProject(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir := t.TempDir()

	// Test 1: gradlew in root
	rootGradlew := filepath.Join(tempDir, "gradlew")
	if runtime.GOOS != "windows" {
		if err := os.WriteFile(rootGradlew, []byte("#!/bin/bash\necho test"), 0755); err != nil {
			t.Fatalf("Failed to create root gradlew: %v", err)
		}
	} else {
		if err := os.WriteFile(rootGradlew, []byte("@echo off\necho test"), 0644); err != nil {
			t.Fatalf("Failed to create root gradlew: %v", err)
		}
	}

	result := findGradlewInProject(tempDir)
	if result != rootGradlew {
		t.Errorf("findGradlewInProject() = %v, want %v", result, rootGradlew)
	}

	// Remove root gradlew for next test
	os.Remove(rootGradlew)

	// Test 2: gradlew in android/ subdirectory
	androidDir := filepath.Join(tempDir, "android")
	if err := os.MkdirAll(androidDir, 0755); err != nil {
		t.Fatalf("Failed to create android dir: %v", err)
	}
	androidGradlew := filepath.Join(androidDir, "gradlew")
	if runtime.GOOS != "windows" {
		if err := os.WriteFile(androidGradlew, []byte("#!/bin/bash\necho test"), 0755); err != nil {
			t.Fatalf("Failed to create android gradlew: %v", err)
		}
	} else {
		if err := os.WriteFile(androidGradlew, []byte("@echo off\necho test"), 0644); err != nil {
			t.Fatalf("Failed to create android gradlew: %v", err)
		}
	}

	result = findGradlewInProject(tempDir)
	if result != androidGradlew {
		t.Errorf("findGradlewInProject() = %v, want %v", result, androidGradlew)
	}

	// Remove android gradlew for next test
	os.Remove(androidGradlew)

	// Test 3: gradlew in android/app/ subdirectory
	androidAppDir := filepath.Join(tempDir, "android", "app")
	if err := os.MkdirAll(androidAppDir, 0755); err != nil {
		t.Fatalf("Failed to create android/app dir: %v", err)
	}
	androidAppGradlew := filepath.Join(androidAppDir, "gradlew")
	if runtime.GOOS != "windows" {
		if err := os.WriteFile(androidAppGradlew, []byte("#!/bin/bash\necho test"), 0755); err != nil {
			t.Fatalf("Failed to create android/app gradlew: %v", err)
		}
	} else {
		if err := os.WriteFile(androidAppGradlew, []byte("@echo off\necho test"), 0644); err != nil {
			t.Fatalf("Failed to create android/app gradlew: %v", err)
		}
	}

	result = findGradlewInProject(tempDir)
	// Note: android/app is checked after android/, so if only android/app has gradlew, it should find it
	// But since android/ is checked first and doesn't have gradlew, it should find android/app/gradlew
	expectedPath := filepath.Join(tempDir, "android", "app", "gradlew")
	if result != expectedPath {
		t.Errorf("findGradlewInProject() = %v, want %v", result, expectedPath)
	}

	// Test 4: no gradlew found
	os.Remove(androidAppGradlew)
	result = findGradlewInProject(tempDir)
	if result != "" {
		t.Errorf("findGradlewInProject() = %v, want empty string", result)
	}
}

func TestFindGradlewInProject_EmptyPath(t *testing.T) {
	// Test with empty path - should use current working directory
	// This test mainly ensures no panic occurs
	result := findGradlewInProject("")
	// Result depends on current working directory, so we just ensure no panic
	_ = result
}
