package checker

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

// ToolRequirement defines minimum version requirements for a tool.
type ToolRequirement struct {
	Name           string
	MinVersion     string
	Command        []string
	VersionArgs    []string
	VersionPattern *regexp.Regexp
	VersionParser  func(string) (string, error)
}

// DefaultRequirements returns the default tool requirements.
func DefaultRequirements() []ToolRequirement {
	return []ToolRequirement{
		{
			Name:           "Node.js",
			MinVersion:     "18.0.0",
			Command:        []string{"node"},
			VersionArgs:    []string{"--version"},
			VersionPattern: regexp.MustCompile(`v?(\d+\.\d+\.\d+)`),
		},
		{
			Name:           "Java",
			MinVersion:     "17.0.0",
			Command:        []string{"java"},
			VersionArgs:    []string{"-version"},
			VersionPattern: regexp.MustCompile(`(\d+\.\d+\.\d+)`),
			VersionParser: func(output string) (string, error) {
				// Java outputs version to stderr
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.Contains(line, "version") || strings.Contains(line, "openjdk") {
						return line, nil
					}
				}
				return output, nil
			},
		},
		{
			Name:           "Android SDK",
			MinVersion:     "33.0.0",
			Command:        []string{"sdkmanager", "--version"},
			VersionArgs:    []string{},
			VersionPattern: regexp.MustCompile(`(\d+\.\d+\.\d+)`),
		},
		{
			Name:           "Flutter",
			MinVersion:     "3.0.0",
			Command:        []string{"flutter"},
			VersionArgs:    []string{"--version"},
			VersionPattern: regexp.MustCompile(`(\d+\.\d+\.\d+)`),
		},
		{
			Name:           "Git",
			MinVersion:     "2.30.0",
			Command:        []string{"git"},
			VersionArgs:    []string{"--version"},
			VersionPattern: regexp.MustCompile(`(\d+\.\d+\.\d+)`),
		},
	}
}

// ToolVersionChecker checks tool versions against minimum requirements.
type ToolVersionChecker struct {
	requirements []ToolRequirement
}

// NewToolVersionChecker creates a new tool version checker with default requirements.
func NewToolVersionChecker() *ToolVersionChecker {
	return &ToolVersionChecker{
		requirements: DefaultRequirements(),
	}
}

// NewToolVersionCheckerWithRequirements creates a checker with custom requirements.
func NewToolVersionCheckerWithRequirements(reqs []ToolRequirement) *ToolVersionChecker {
	return &ToolVersionChecker{
		requirements: reqs,
	}
}

// Name returns the checker name.
func (t *ToolVersionChecker) Name() string {
	return "Tool Versions"
}

// Category returns the checker category.
func (t *ToolVersionChecker) Category() doctor.CheckCategory {
	return doctor.CategoryTools
}

// Check performs the tool version checks.
func (t *ToolVersionChecker) Check() doctor.CheckResult {
	results := make([]ToolCheckResult, 0, len(t.requirements))
	var failedTools []string
	var outdatedTools []string
	var missingTools []string

	for _, req := range t.requirements {
		result := checkToolVersion(req)
		results = append(results, result)

		switch result.Status {
		case "missing":
			missingTools = append(missingTools, req.Name)
		case "outdated":
			outdatedTools = append(outdatedTools, req.Name)
		case "error":
			failedTools = append(failedTools, req.Name)
		}
	}

	result := doctor.CheckResult{
		Name: t.Name(),
		Details: map[string]interface{}{
			"toolsChecked": len(t.requirements),
			"results":      results,
		},
	}

	if len(missingTools) > 0 {
		result.Status = doctor.StatusError
		result.Message = fmt.Sprintf("Missing tools: %v", missingTools)
	} else if len(failedTools) > 0 {
		result.Status = doctor.StatusError
		result.Message = fmt.Sprintf("Failed to check tools: %v", failedTools)
	} else if len(outdatedTools) > 0 {
		result.Status = doctor.StatusWarning
		result.Message = fmt.Sprintf("Outdated tools: %v", outdatedTools)
	} else {
		result.Status = doctor.StatusPassed
		result.Message = "All tools meet minimum version requirements"
	}

	return result
}

// ToolCheckResult contains the result of checking a single tool.
type ToolCheckResult struct {
	Name       string `json:"name"`
	Installed  bool   `json:"installed"`
	Version    string `json:"version,omitempty"`
	MinVersion string `json:"minVersion"`
	Status     string `json:"status"` // ok, outdated, missing, error
	Message    string `json:"message,omitempty"`
}

// checkToolVersion checks a single tool's version.
func checkToolVersion(req ToolRequirement) ToolCheckResult {
	result := ToolCheckResult{
		Name:       req.Name,
		MinVersion: req.MinVersion,
	}

	// Check if tool is available
	cmdPath, err := exec.LookPath(req.Command[0])
	if err != nil {
		result.Installed = false
		result.Status = "missing"
		result.Message = fmt.Sprintf("%s not found in PATH", req.Name)
		return result
	}

	result.Installed = true

	// Get version
	args := append(req.Command[1:], req.VersionArgs...)
	cmd := exec.Command(cmdPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("Failed to get version: %v", err)
		return result
	}

	outputStr := string(output)

	// Apply custom parser if provided
	if req.VersionParser != nil {
		parsed, err := req.VersionParser(outputStr)
		if err == nil {
			outputStr = parsed
		}
	}

	// Extract version using regex
	version := ""
	if req.VersionPattern != nil {
		matches := req.VersionPattern.FindStringSubmatch(outputStr)
		if len(matches) > 1 {
			version = matches[1]
		} else if len(matches) == 1 {
			version = matches[0]
		}
	}

	if version == "" {
		result.Status = "error"
		result.Message = "Could not parse version"
		return result
	}

	result.Version = version

	// Compare versions
	comparison, err := compareVersions(version, req.MinVersion)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("Version comparison failed: %v", err)
		return result
	}

	if comparison >= 0 {
		result.Status = "ok"
		result.Message = fmt.Sprintf("Version %s meets minimum %s", version, req.MinVersion)
	} else {
		result.Status = "outdated"
		result.Message = fmt.Sprintf("Version %s is below minimum %s", version, req.MinVersion)
	}

	return result
}

// compareVersions compares two semantic version strings.
// Returns: >0 if v1 > v2, 0 if v1 == v2, <0 if v1 < v2
func compareVersions(v1, v2 string) (int, error) {
	// Parse versions
	parts1, err := parseVersionParts(v1)
	if err != nil {
		return 0, fmt.Errorf("invalid version %s: %v", v1, err)
	}

	parts2, err := parseVersionParts(v2)
	if err != nil {
		return 0, fmt.Errorf("invalid version %s: %v", v2, err)
	}

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}

		if p1 > p2 {
			return 1, nil
		}
		if p1 < p2 {
			return -1, nil
		}
	}

	return 0, nil
}

// parseVersionParts parses a version string into numeric parts.
func parseVersionParts(version string) ([]int, error) {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")

	// Split by dots
	parts := strings.Split(version, ".")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		// Try to parse as integer
		n, err := strconv.Atoi(part)
		if err != nil {
			// Skip non-numeric parts (like pre-release tags)
			break
		}
		result = append(result, n)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no numeric version parts found")
	}

	return result, nil
}

// IsVersionCompatible checks if a version meets the minimum requirement.
func IsVersionCompatible(version, minVersion string) bool {
	comparison, err := compareVersions(version, minVersion)
	if err != nil {
		return false
	}
	return comparison >= 0
}
