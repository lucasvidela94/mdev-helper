package checker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

func TestPerformanceCheckerName(t *testing.T) {
	checker := NewPerformanceChecker(".")
	if checker.Name() != "Performance Recommendations" {
		t.Errorf("Expected name 'Performance Recommendations', got %s", checker.Name())
	}
}

func TestPerformanceCheckerCategory(t *testing.T) {
	checker := NewPerformanceChecker(".")
	if checker.Category() != doctor.CategoryPerformance {
		t.Errorf("Expected category 'performance', got %s", checker.Category())
	}
}

func TestPerformanceCheckerCheck(t *testing.T) {
	checker := NewPerformanceChecker(".")
	result := checker.Check()

	// Should have a name
	if result.Name != "Performance Recommendations" {
		t.Errorf("Expected result name 'Performance Recommendations', got %s", result.Name)
	}

	// Should have valid status
	if result.Status != doctor.StatusPassed && result.Status != doctor.StatusWarning {
		t.Errorf("Expected status 'passed' or 'warning', got %s", result.Status)
	}

	// Should have a message
	if result.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Should have details
	if result.Details == nil {
		t.Fatal("Expected details to be present")
	}
}

func TestPerformanceCheckerWithEmptyPath(t *testing.T) {
	checker := NewPerformanceChecker("")
	if checker.projectPath != "." {
		t.Errorf("Expected project path to be '.', got %s", checker.projectPath)
	}
}

func TestParseMemoryValue(t *testing.T) {
	tests := []struct {
		input    string
		expected int // in MB
	}{
		{"4g", 4096},
		{"4G", 4096},
		{"4096m", 4096},
		{"4096M", 4096},
		{"8192k", 8},
		{"4194304", 4},
		{"2g", 2048},
		{"8g", 8192},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseMemoryValue(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %d MB, got %d MB", tt.expected, result)
			}
		})
	}
}

func TestHasAdequateMemory(t *testing.T) {
	tests := []struct {
		content  string
		expected bool
	}{
		{"org.gradle.jvmargs=-Xmx4g", true},
		{"org.gradle.jvmargs=-Xmx4096m", true},
		{"org.gradle.jvmargs=-Xmx2g", false},
		{"org.gradle.jvmargs=-Xmx2048m", false},
		{"org.gradle.jvmargs=-Xmx1g -XX:+HeapDumpOnOutOfMemoryError", false},
		{"# org.gradle.jvmargs=-Xmx4g", false}, // Commented out
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			result := hasAdequateMemory(tt.content)
			if result != tt.expected {
				t.Errorf("Expected %v for '%s'", tt.expected, tt.content)
			}
		})
	}
}

func TestCheckGradleDaemon(t *testing.T) {
	// Create a temp directory with gradle.properties in android/ subdirectory
	tempDir := t.TempDir()
	androidDir := filepath.Join(tempDir, "android")
	os.MkdirAll(androidDir, 0755)
	gradleProps := filepath.Join(androidDir, "gradle.properties")

	checker := NewPerformanceChecker(tempDir)

	// Without gradle.properties, should recommend daemon
	rec := checker.checkGradleDaemon()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "recommended" {
		t.Errorf("Expected status 'recommended', got '%s'", rec.Status)
	}

	// With daemon enabled, should show as applied
	content := "org.gradle.daemon=true\n"
	os.WriteFile(gradleProps, []byte(content), 0644)

	rec = checker.checkGradleDaemon()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "applied" {
		t.Errorf("Expected status 'applied', got '%s'", rec.Status)
	}
}

func TestCheckParallelBuilds(t *testing.T) {
	tempDir := t.TempDir()
	androidDir := filepath.Join(tempDir, "android")
	os.MkdirAll(androidDir, 0755)
	gradleProps := filepath.Join(androidDir, "gradle.properties")

	checker := NewPerformanceChecker(tempDir)

	// Without gradle.properties, should recommend parallel builds
	rec := checker.checkParallelBuilds()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "recommended" {
		t.Errorf("Expected status 'recommended', got '%s'", rec.Status)
	}

	// With parallel builds enabled, should show as applied
	content := "org.gradle.parallel=true\n"
	os.WriteFile(gradleProps, []byte(content), 0644)

	rec = checker.checkParallelBuilds()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "applied" {
		t.Errorf("Expected status 'applied', got '%s'", rec.Status)
	}
}

func TestCheckMemorySettings(t *testing.T) {
	tempDir := t.TempDir()
	androidDir := filepath.Join(tempDir, "android")
	os.MkdirAll(androidDir, 0755)
	gradleProps := filepath.Join(androidDir, "gradle.properties")

	checker := NewPerformanceChecker(tempDir)

	// Without gradle.properties, should recommend memory settings
	rec := checker.checkMemorySettings()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "recommended" {
		t.Errorf("Expected status 'recommended', got '%s'", rec.Status)
	}

	// With adequate memory, should show as applied
	content := "org.gradle.jvmargs=-Xmx4g\n"
	os.WriteFile(gradleProps, []byte(content), 0644)

	rec = checker.checkMemorySettings()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "applied" {
		t.Errorf("Expected status 'applied', got '%s'", rec.Status)
	}

	// With inadequate memory, should recommend increase
	content = "org.gradle.jvmargs=-Xmx2g\n"
	os.WriteFile(gradleProps, []byte(content), 0644)

	rec = checker.checkMemorySettings()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "recommended" {
		t.Errorf("Expected status 'recommended' for low memory, got '%s'", rec.Status)
	}
}

func TestCheckBuildCache(t *testing.T) {
	tempDir := t.TempDir()
	androidDir := filepath.Join(tempDir, "android")
	os.MkdirAll(androidDir, 0755)
	gradleProps := filepath.Join(androidDir, "gradle.properties")

	checker := NewPerformanceChecker(tempDir)

	// Without gradle.properties, should recommend build cache
	rec := checker.checkBuildCache()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "recommended" {
		t.Errorf("Expected status 'recommended', got '%s'", rec.Status)
	}

	// With build cache enabled, should show as applied
	content := "org.gradle.caching=true\n"
	os.WriteFile(gradleProps, []byte(content), 0644)

	rec = checker.checkBuildCache()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "applied" {
		t.Errorf("Expected status 'applied', got '%s'", rec.Status)
	}
}

func TestCheckNodeMemoryWithNodeOptions(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("NODE_OPTIONS")
	defer os.Setenv("NODE_OPTIONS", originalValue)

	// Set NODE_OPTIONS with memory setting
	os.Setenv("NODE_OPTIONS", "--max-old-space-size=4096")

	tempDir := t.TempDir()
	checker := NewPerformanceChecker(tempDir)

	rec := checker.checkNodeMemory()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "applied" {
		t.Errorf("Expected status 'applied', got '%s'", rec.Status)
	}
}

func TestCheckNodeMemoryWithReactNative(t *testing.T) {
	// Clear NODE_OPTIONS
	os.Setenv("NODE_OPTIONS", "")

	tempDir := t.TempDir()

	// Create a package.json with react-native
	packageJSON := filepath.Join(tempDir, "package.json")
	content := `{"dependencies": {"react-native": "0.72.0"}}`
	os.WriteFile(packageJSON, []byte(content), 0644)

	checker := NewPerformanceChecker(tempDir)

	rec := checker.checkNodeMemory()
	if rec == nil {
		t.Fatal("Expected recommendation")
	}
	if rec.Status != "recommended" {
		t.Errorf("Expected status 'recommended', got '%s'", rec.Status)
	}
}

func TestCheckNodeMemoryNotApplicable(t *testing.T) {
	// Clear NODE_OPTIONS
	os.Setenv("NODE_OPTIONS", "")

	tempDir := t.TempDir()
	checker := NewPerformanceChecker(tempDir)

	// Without package.json or .nvmrc, should return nil
	rec := checker.checkNodeMemory()
	if rec != nil {
		t.Error("Expected nil for non-Node project")
	}
}

func TestPerformanceRecommendationStructure(t *testing.T) {
	rec := PerformanceRecommendation{
		Title:       "Test Recommendation",
		Description: "This is a test",
		Status:      "recommended",
		Priority:    "high",
		Action:      "Do something",
	}

	if rec.Title != "Test Recommendation" {
		t.Errorf("Expected title 'Test Recommendation', got '%s'", rec.Title)
	}
	if rec.Status != "recommended" {
		t.Errorf("Expected status 'recommended', got '%s'", rec.Status)
	}
	if rec.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", rec.Priority)
	}
}
