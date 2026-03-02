// Package service provides business logic for environment management.
package service

import (
	"time"

	"github.com/sombi/mobile-dev-helper/internal/detector"
	"github.com/sombi/mobile-dev-helper/internal/detector/dart"
	"github.com/sombi/mobile-dev-helper/internal/detector/flutter"
	"github.com/sombi/mobile-dev-helper/internal/detector/gradlew"
	"github.com/sombi/mobile-dev-helper/internal/detector/jdk"
	"github.com/sombi/mobile-dev-helper/internal/detector/node"
	"github.com/sombi/mobile-dev-helper/internal/detector/pathtools"
	"github.com/sombi/mobile-dev-helper/internal/detector/sdk"
	"github.com/sombi/mobile-dev-helper/internal/detector/shell"
	"github.com/sombi/mobile-dev-helper/internal/expo"
	"github.com/sombi/mobile-dev-helper/internal/mise"
	"github.com/sombi/mobile-dev-helper/internal/suggestions"
)

// EnvironmentService orchestrates all environment detectors.
type EnvironmentService struct {
	projectPath string
	shellInfo   *shell.ShellInfo
}

// NewEnvironmentService creates a new EnvironmentService.
func NewEnvironmentService(projectPath string) *EnvironmentService {
	return &EnvironmentService{
		projectPath: projectPath,
		shellInfo:   shell.Detect(),
	}
}

// Detect runs all environment detectors and returns a comprehensive report.
func (s *EnvironmentService) Detect() *detector.EnvironmentReport {
	report := &detector.EnvironmentReport{
		Timestamp: time.Now(),
		Warnings:  []string{},
		Errors:    []string{},
	}

	// Detect JDK
	jdkInfo := jdk.Detect()
	report.JDK = jdkInfo
	if jdkInfo != nil && !jdkInfo.IsValid {
		report.Errors = append(report.Errors, jdkInfo.Error)
	}
	if jdkInfo != nil && jdkInfo.Warning != "" {
		report.Warnings = append(report.Warnings, jdkInfo.Warning)
	}

	// Detect Android SDK
	sdkInfo := sdk.Detect()
	report.SDK = sdkInfo
	if sdkInfo != nil && !sdkInfo.IsValid {
		report.Errors = append(report.Errors, sdkInfo.Error)
	}

	// Detect Node.js
	nodeInfo := node.Detect()
	report.Node = nodeInfo
	if nodeInfo != nil && !nodeInfo.IsValid {
		report.Errors = append(report.Errors, nodeInfo.Error)
	}

	// Detect Gradle (if project path is set)
	if s.projectPath != "" {
		gradleInfo := gradlew.Detect(s.projectPath)
		report.Gradle = gradleInfo
		if gradleInfo != nil && !gradleInfo.IsValid {
			report.Warnings = append(report.Warnings, gradleInfo.Error)
		}
	}

	// Detect Expo project
	expoInfo := expo.Detect(s.projectPath)
	report.Expo = expoInfo
	if expoInfo != nil && expoInfo.Error != "" {
		report.Warnings = append(report.Warnings, expoInfo.Error)
	}

	// Detect Flutter SDK
	flutterInfo := flutter.Detect()
	report.Flutter = flutterInfo
	if flutterInfo != nil && !flutterInfo.IsValid {
		report.Warnings = append(report.Warnings, flutterInfo.Error)
	}
	if flutterInfo != nil && flutterInfo.Warning != "" {
		report.Warnings = append(report.Warnings, flutterInfo.Warning)
	}

	// Detect Dart SDK (preferably bundled with Flutter)
	var flutterPath string
	if flutterInfo != nil && flutterInfo.IsValid {
		flutterPath = flutterInfo.Path
	}
	dartInfo := dart.Detect(flutterPath)
	report.Dart = dartInfo
	if dartInfo != nil && !dartInfo.IsValid {
		report.Warnings = append(report.Warnings, dartInfo.Error)
	}

	// Check mise integration
	if mise.IsInstalled() {
		// Get info for common tools
		tools := []string{"node", "java", "ruby", "python", "flutter", "dart"}
		for _, tool := range tools {
			if info, err := mise.GetToolInfo(tool); err == nil && info.IsInstalled {
				report.MiseTools = append(report.MiseTools, *info)
			}
		}
	}

	// Check PATH tools
	pathtools.AddToReport(report)

	// Add recommendations
	report.Warnings = s.addRecommendations(report)

	return report
}

// addRecommendations adds helpful recommendations based on detected issues.
func (s *EnvironmentService) addRecommendations(report *detector.EnvironmentReport) []string {
	recommendations := report.Warnings
	gen := suggestions.NewGeneratorWithShell(s.shellInfo)

	// Check for missing JDK
	if report.JDK == nil || !report.JDK.IsValid {
		recommendations = append(recommendations, "Install JDK using: mise install java")
	}

	// Check for missing Android SDK
	if report.SDK == nil || !report.SDK.IsValid {
		suggestion := gen.ForMissingAndroidHome("")
		recommendations = append(recommendations, suggestion.Issue+" - "+suggestion.Solution)
	} else {
		// Check for missing platforms
		if len(report.SDK.Platforms) == 0 {
			suggestion := gen.ForMissingSDKPlatforms()
			recommendations = append(recommendations, suggestion.Issue+" - "+suggestion.Solution)
		}
		// Check for missing build tools
		if len(report.SDK.BuildTools) == 0 {
			suggestion := gen.ForMissingBuildTools()
			recommendations = append(recommendations, suggestion.Issue+" - "+suggestion.Solution)
		}
	}

	// Check for missing Node
	if report.Node == nil || !report.Node.IsValid {
		recommendations = append(recommendations, "Install Node.js using: mise install node")
	}

	// Check for JAVA_HOME not set
	if report.JDK != nil && report.JDK.Warning != "" {
		if report.JDK.Path != "" {
			suggestion := gen.ForMissingJavaHome(report.JDK.Path)
			recommendations = append(recommendations, suggestion.Issue+" - "+suggestion.Solution)
		} else {
			recommendations = append(recommendations, "Set JAVA_HOME environment variable")
		}
	}

	// Check for mise not installed
	if !mise.IsInstalled() {
		suggestion := gen.ForMiseNotInstalled()
		recommendations = append(recommendations, suggestion.Issue+" - "+suggestion.Solution)
	}

	// Check for missing Flutter
	if report.Flutter == nil || !report.Flutter.IsValid {
		recommendations = append(recommendations, "Install Flutter using: mise install flutter")
	}

	// Check for FLUTTER_HOME not set
	if report.Flutter != nil && report.Flutter.Warning != "" {
		if report.Flutter.Path != "" {
			recommendations = append(recommendations, "Set FLUTTER_HOME environment variable to: "+report.Flutter.Path)
		} else {
			recommendations = append(recommendations, "Set FLUTTER_HOME environment variable")
		}
	}

	return recommendations
}
