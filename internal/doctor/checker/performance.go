package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

// PerformanceChecker checks for performance optimization opportunities.
type PerformanceChecker struct {
	projectPath string
}

// NewPerformanceChecker creates a new performance checker.
func NewPerformanceChecker(projectPath string) *PerformanceChecker {
	if projectPath == "" {
		projectPath = "."
	}
	return &PerformanceChecker{projectPath: projectPath}
}

// Name returns the checker name.
func (p *PerformanceChecker) Name() string {
	return "Performance Recommendations"
}

// Category returns the checker category.
func (p *PerformanceChecker) Category() doctor.CheckCategory {
	return doctor.CategoryPerformance
}

// Check performs the performance checks.
func (p *PerformanceChecker) Check() doctor.CheckResult {
	recommendations := make([]PerformanceRecommendation, 0)

	// Check Gradle daemon
	if rec := p.checkGradleDaemon(); rec != nil {
		recommendations = append(recommendations, *rec)
	}

	// Check parallel builds
	if rec := p.checkParallelBuilds(); rec != nil {
		recommendations = append(recommendations, *rec)
	}

	// Check memory settings
	if rec := p.checkMemorySettings(); rec != nil {
		recommendations = append(recommendations, *rec)
	}

	// Check for build cache
	if rec := p.checkBuildCache(); rec != nil {
		recommendations = append(recommendations, *rec)
	}

	// Check for Node.js memory
	if rec := p.checkNodeMemory(); rec != nil {
		recommendations = append(recommendations, *rec)
	}

	// Count issues
	issueCount := 0
	for _, rec := range recommendations {
		if rec.Status == "recommended" {
			issueCount++
		}
	}

	result := doctor.CheckResult{
		Name: p.Name(),
		Details: map[string]interface{}{
			"recommendations": recommendations,
			"issuesFound":     issueCount,
		},
	}

	if issueCount > 0 {
		result.Status = doctor.StatusWarning
		result.Message = fmt.Sprintf("Found %d performance optimization opportunities", issueCount)
	} else {
		result.Status = doctor.StatusPassed
		result.Message = "No performance issues detected"
	}

	return result
}

// PerformanceRecommendation represents a performance recommendation.
type PerformanceRecommendation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`   // applied, recommended, not_applicable
	Priority    string `json:"priority"` // high, medium, low
	Action      string `json:"action,omitempty"`
}

// checkGradleDaemon checks if Gradle daemon is enabled.
func (p *PerformanceChecker) checkGradleDaemon() *PerformanceRecommendation {
	// Check gradle.properties in project
	gradleProps := filepath.Join(p.projectPath, "android", "gradle.properties")

	if _, err := os.Stat(gradleProps); err != nil {
		// No Android project, check global gradle.properties
		home := os.Getenv("HOME")
		gradleProps = filepath.Join(home, ".gradle", "gradle.properties")
	}

	content, err := os.ReadFile(gradleProps)
	if err != nil {
		// Can't check, assume not configured
		return &PerformanceRecommendation{
			Title:       "Gradle Daemon",
			Description: "Gradle daemon speeds up builds by keeping the JVM running",
			Status:      "recommended",
			Priority:    "high",
			Action:      "Add 'org.gradle.daemon=true' to ~/.gradle/gradle.properties",
		}
	}

	contentStr := string(content)
	if isPropertyEnabled(contentStr, "org.gradle.daemon") {
		return &PerformanceRecommendation{
			Title:       "Gradle Daemon",
			Description: "Gradle daemon is enabled",
			Status:      "applied",
			Priority:    "high",
		}
	}

	return &PerformanceRecommendation{
		Title:       "Gradle Daemon",
		Description: "Gradle daemon speeds up builds by keeping the JVM running",
		Status:      "recommended",
		Priority:    "high",
		Action:      "Add 'org.gradle.daemon=true' to gradle.properties",
	}
}

// checkParallelBuilds checks if parallel builds are enabled.
func (p *PerformanceChecker) checkParallelBuilds() *PerformanceRecommendation {
	gradleProps := filepath.Join(p.projectPath, "android", "gradle.properties")

	if _, err := os.Stat(gradleProps); err != nil {
		home := os.Getenv("HOME")
		gradleProps = filepath.Join(home, ".gradle", "gradle.properties")
	}

	content, err := os.ReadFile(gradleProps)
	if err != nil {
		return &PerformanceRecommendation{
			Title:       "Parallel Builds",
			Description: "Parallel builds can significantly reduce build time",
			Status:      "recommended",
			Priority:    "high",
			Action:      "Add 'org.gradle.parallel=true' to gradle.properties",
		}
	}

	contentStr := string(content)
	if isPropertyEnabled(contentStr, "org.gradle.parallel") {
		return &PerformanceRecommendation{
			Title:       "Parallel Builds",
			Description: "Parallel builds are enabled",
			Status:      "applied",
			Priority:    "high",
		}
	}

	return &PerformanceRecommendation{
		Title:       "Parallel Builds",
		Description: "Parallel builds can significantly reduce build time",
		Status:      "recommended",
		Priority:    "high",
		Action:      "Add 'org.gradle.parallel=true' to gradle.properties",
	}
}

// checkMemorySettings checks if appropriate memory settings are configured.
func (p *PerformanceChecker) checkMemorySettings() *PerformanceRecommendation {
	gradleProps := filepath.Join(p.projectPath, "android", "gradle.properties")

	if _, err := os.Stat(gradleProps); err != nil {
		home := os.Getenv("HOME")
		gradleProps = filepath.Join(home, ".gradle", "gradle.properties")
	}

	content, err := os.ReadFile(gradleProps)
	if err != nil {
		return &PerformanceRecommendation{
			Title:       "JVM Memory Settings",
			Description: "Adequate JVM memory is required for large builds",
			Status:      "recommended",
			Priority:    "high",
			Action:      "Add 'org.gradle.jvmargs=-Xmx4g' to gradle.properties",
		}
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "org.gradle.jvmargs") {
		// Extract memory setting
		if hasAdequateMemory(contentStr) {
			return &PerformanceRecommendation{
				Title:       "JVM Memory Settings",
				Description: "Adequate JVM memory is configured",
				Status:      "applied",
				Priority:    "high",
			}
		}
		return &PerformanceRecommendation{
			Title:       "JVM Memory Settings",
			Description: "JVM memory is configured but may be insufficient for large builds",
			Status:      "recommended",
			Priority:    "medium",
			Action:      "Increase heap size in org.gradle.jvmargs to at least -Xmx4g",
		}
	}

	return &PerformanceRecommendation{
		Title:       "JVM Memory Settings",
		Description: "Adequate JVM memory is required for large builds",
		Status:      "recommended",
		Priority:    "high",
		Action:      "Add 'org.gradle.jvmargs=-Xmx4g' to gradle.properties",
	}
}

// hasAdequateMemory checks if the memory setting is adequate (>= 4GB).
func hasAdequateMemory(content string) bool {
	// Look for -Xmx setting
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip commented lines
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}
		if strings.Contains(line, "org.gradle.jvmargs") {
			// Extract memory value
			if idx := strings.Index(line, "-Xmx"); idx != -1 {
				memoryPart := line[idx+4:]
				// Find the end of the memory value
				endIdx := strings.IndexAny(memoryPart, " \"")
				if endIdx == -1 {
					endIdx = len(memoryPart)
				}
				memoryValue := memoryPart[:endIdx]

				// Parse the memory value
				return parseMemoryValue(memoryValue) >= 4*1024 // 4GB in MB
			}
		}
	}
	return false
}

// parseMemoryValue parses a memory value like "4g" or "4096m" into MB.
func parseMemoryValue(value string) int {
	value = strings.ToLower(strings.TrimSpace(value))

	if strings.HasSuffix(value, "g") {
		gb, _ := strconv.Atoi(strings.TrimSuffix(value, "g"))
		return gb * 1024
	}
	if strings.HasSuffix(value, "m") {
		mb, _ := strconv.Atoi(strings.TrimSuffix(value, "m"))
		return mb
	}
	if strings.HasSuffix(value, "k") {
		kb, _ := strconv.Atoi(strings.TrimSuffix(value, "k"))
		return kb / 1024
	}

	// Assume bytes
	bytes, _ := strconv.Atoi(value)
	return bytes / (1024 * 1024)
}

// checkBuildCache checks if build cache is enabled.
func (p *PerformanceChecker) checkBuildCache() *PerformanceRecommendation {
	// Check for build cache in gradle.properties
	gradleProps := filepath.Join(p.projectPath, "android", "gradle.properties")

	if _, err := os.Stat(gradleProps); err != nil {
		home := os.Getenv("HOME")
		gradleProps = filepath.Join(home, ".gradle", "gradle.properties")
	}

	content, err := os.ReadFile(gradleProps)
	if err != nil {
		return &PerformanceRecommendation{
			Title:       "Build Cache",
			Description: "Build cache speeds up builds by reusing outputs",
			Status:      "recommended",
			Priority:    "medium",
			Action:      "Add 'org.gradle.caching=true' to gradle.properties",
		}
	}

	contentStr := string(content)
	if isPropertyEnabled(contentStr, "org.gradle.caching") {
		return &PerformanceRecommendation{
			Title:       "Build Cache",
			Description: "Build cache is enabled",
			Status:      "applied",
			Priority:    "medium",
		}
	}

	return &PerformanceRecommendation{
		Title:       "Build Cache",
		Description: "Build cache speeds up builds by reusing outputs",
		Status:      "recommended",
		Priority:    "medium",
		Action:      "Add 'org.gradle.caching=true' to gradle.properties",
	}
}

// checkNodeMemory checks if Node.js memory is configured appropriately.
func (p *PerformanceChecker) checkNodeMemory() *PerformanceRecommendation {
	// Check for NODE_OPTIONS in environment
	nodeOptions := os.Getenv("NODE_OPTIONS")

	if nodeOptions != "" && strings.Contains(nodeOptions, "--max-old-space-size") {
		return &PerformanceRecommendation{
			Title:       "Node.js Memory",
			Description: "Node.js memory is configured via NODE_OPTIONS",
			Status:      "applied",
			Priority:    "medium",
		}
	}

	// Check for .nvmrc or .node-version with memory hints (not standard, but useful)
	nvmrcPath := filepath.Join(p.projectPath, ".nvmrc")
	if _, err := os.Stat(nvmrcPath); err == nil {
		// Project uses nvm, recommend NODE_OPTIONS
		return &PerformanceRecommendation{
			Title:       "Node.js Memory",
			Description: "Large JavaScript bundles may require more Node.js memory",
			Status:      "recommended",
			Priority:    "medium",
			Action:      "Set NODE_OPTIONS='--max-old-space-size=4096' for large builds",
		}
	}

	// Check if this is a React Native / Expo project that might need more memory
	packageJSON := filepath.Join(p.projectPath, "package.json")
	if content, err := os.ReadFile(packageJSON); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "react-native") || strings.Contains(contentStr, "expo") {
			return &PerformanceRecommendation{
				Title:       "Node.js Memory",
				Description: "React Native builds may require more Node.js memory",
				Status:      "recommended",
				Priority:    "medium",
				Action:      "Set NODE_OPTIONS='--max-old-space-size=4096'",
			}
		}
	}

	return nil // Not applicable for non-Node projects
}

// isPropertyEnabled checks if a property is enabled (not commented out and set to true).
func isPropertyEnabled(content, property string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip commented lines
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}
		if strings.Contains(line, property+"=true") {
			return true
		}
	}
	return false
}

// GetSystemMemory returns the total system memory in GB (approximate).
func GetSystemMemory() int {
	// This is a simplified check - on Linux we could read /proc/meminfo
	// For now, return a default based on the OS
	switch runtime.GOOS {
	case "darwin":
		// Try to get memory using sysctl
		// This would require exec, so we'll use a default
		return 16
	case "linux":
		// Try to read /proc/meminfo
		if content, err := os.ReadFile("/proc/meminfo"); err == nil {
			contentStr := string(content)
			if idx := strings.Index(contentStr, "MemTotal:"); idx != -1 {
				line := contentStr[idx:]
				if endIdx := strings.Index(line, "\n"); endIdx != -1 {
					line = line[:endIdx]
					// Parse the value (in kB)
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						if kb, err := strconv.Atoi(fields[1]); err == nil {
							return kb / (1024 * 1024) // Convert to GB
						}
					}
				}
			}
		}
		return 8 // Default assumption
	case "windows":
		return 16 // Default assumption
	default:
		return 8
	}
}
