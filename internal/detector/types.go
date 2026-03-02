// Package detector provides types for environment detection.
package detector

import (
	"time"
)

// JDKInfo contains information about a detected JDK installation.
type JDKInfo struct {
	Path         string `json:"path"`              // Path to JDK installation
	Version      string `json:"version"`           // Full version (e.g., "17.0.2")
	MajorVersion string `json:"majorVersion"`      // Major version (e.g., "17")
	Vendor       string `json:"vendor"`            // Vendor (e.g., "OpenJDK", "Amazon Corretto")
	JavaHome     string `json:"javaHome"`          // JAVA_HOME value if set
	IsValid      bool   `json:"isValid"`           // Whether the installation is valid
	Warning      string `json:"warning,omitempty"` // Warning message (e.g., JAVA_HOME not set)
	Error        string `json:"error,omitempty"`   // Error message if not valid
}

// SDKInfo contains information about a detected Android SDK.
type SDKInfo struct {
	Path             string   `json:"path"`            // Path to SDK
	Version          string   `json:"version"`         // SDK version
	BuildTools       []string `json:"buildTools"`      // Available build tools versions
	Platforms        []string `json:"platforms"`       // Available platform versions
	NDK              string   `json:"ndk,omitempty"`   // NDK version if available
	CommandLineTools string   `json:"cmdlineTools"`    // Command line tools version
	IsValid          bool     `json:"isValid"`         // Whether the SDK is valid
	Error            string   `json:"error,omitempty"` // Error message if not valid
}

// NodeInfo contains information about a detected Node.js installation.
type NodeInfo struct {
	Path         string `json:"path"`            // Path to Node executable
	Version      string `json:"version"`         // Full version (e.g., "20.10.0")
	MajorVersion string `json:"majorVersion"`    // Major version (e.g., "20")
	IsValid      bool   `json:"isValid"`         // Whether the installation is valid
	Error        string `json:"error,omitempty"` // Error message if not valid
}

// GradleInfo contains information about a detected Gradle wrapper.
type GradleInfo struct {
	Path          string `json:"path"`            // Path to gradlew script
	GradleVersion string `json:"gradleVersion"`   // Gradle version from wrapper
	JavaHome      string `json:"javaHome"`        // JAVA_HOME used by Gradle
	IsValid       bool   `json:"isValid"`         // Whether the wrapper is valid
	HasWrapper    bool   `json:"hasWrapper"`      // Whether wrapper exists
	Error         string `json:"error,omitempty"` // Error message if not valid
}

// ExpoInfo contains information about an Expo project.
type ExpoInfo struct {
	IsManaged   bool   `json:"isManaged"`       // true if Managed workflow
	IsBare      bool   `json:"isBare"`          // true if Bare workflow
	HasExpo     bool   `json:"hasExpo"`         // true if Expo SDK is present
	ExpoVersion string `json:"expoVersion"`     // Expo SDK version
	ProjectPath string `json:"projectPath"`     // Path to project root
	Error       string `json:"error,omitempty"` // Error message if not valid
}

// MiseInfo contains information about a mise tool.
type MiseInfo struct {
	Name        string `json:"name"`        // Tool name (e.g., "node", "java")
	Version     string `json:"version"`     // Installed version
	IsInstalled bool   `json:"isInstalled"` // Whether tool is installed via mise
	Path        string `json:"path"`        // Path to the tool
}

// AVDInfo contains information about an Android Virtual Device.
type AVDInfo struct {
	Name      string `json:"name"`            // AVD name
	TargetAPI string `json:"targetApi"`       // Target API level (e.g., "android-34")
	Device    string `json:"device"`          // Device profile (e.g., "pixel_7")
	Arch      string `json:"arch"`            // Architecture (x86_64, arm64)
	Path      string `json:"path"`            // Path to AVD directory
	IsValid   bool   `json:"isValid"`         // Whether AVD config is valid
	Error     string `json:"error,omitempty"` // Error message if invalid
}

// RunningEmulatorInfo contains information about a running emulator.
type RunningEmulatorInfo struct {
	DeviceID string `json:"deviceId"`          // Device ID (e.g., "emulator-5554")
	Status   string `json:"status"`            // Connection status (device, offline, unauthorized)
	AVDName  string `json:"avdName,omitempty"` // Associated AVD name if detectable
	Product  string `json:"product,omitempty"` // Product name from adb
	Model    string `json:"model,omitempty"`   // Model name from adb
}

// EmulatorInfo contains information about Android emulator installation and AVDs.
type EmulatorInfo struct {
	BinaryPath string                `json:"binaryPath,omitempty"` // Path to emulator binary
	Version    string                `json:"version,omitempty"`    // Emulator version
	AVDs       []AVDInfo             `json:"avds"`                 // Available AVDs
	Running    []RunningEmulatorInfo `json:"running"`              // Currently running emulators
	IsValid    bool                  `json:"isValid"`              // Whether emulator is properly set up
	Error      string                `json:"error,omitempty"`      // Error message if not valid
}

// CleanResult contains the result of a cache cleaning operation.
type CleanResult struct {
	Path      string `json:"path"`            // Path that was cleaned
	Success   bool   `json:"success"`         // Whether cleaning succeeded
	SizeFreed string `json:"sizeFreed"`       // Size of freed space
	Error     string `json:"error,omitempty"` // Error message if failed
}

// RepairResult contains the result of a repair operation.
type RepairResult struct {
	Tool    string `json:"tool"`            // Tool that was repaired
	Action  string `json:"action"`          // Action taken (e.g., "chmod", "download")
	Success bool   `json:"success"`         // Whether repair succeeded
	Message string `json:"message"`         // Human-readable message
	Error   string `json:"error,omitempty"` // Error message if failed
}

// FlutterInfo contains information about a detected Flutter SDK installation.
type FlutterInfo struct {
	Path         string         `json:"path"`              // Path to Flutter SDK
	Version      string         `json:"version"`           // Flutter version (e.g., "3.24.0")
	Channel      string         `json:"channel"`           // Flutter channel (stable, beta, dev, master)
	FrameworkRev string         `json:"frameworkRev"`      // Framework revision hash
	EngineRev    string         `json:"engineRev"`         // Engine revision hash
	DartPath     string         `json:"dartPath"`          // Path to bundled Dart SDK
	IsValid      bool           `json:"isValid"`           // Whether the installation is valid
	FlutterHome  string         `json:"flutterHome"`       // FLUTTER_HOME value if set
	Warning      string         `json:"warning,omitempty"` // Warning message (e.g., FLUTTER_HOME not set)
	Error        string         `json:"error,omitempty"`   // Error message if not valid
	Doctor       *FlutterDoctor `json:"doctor,omitempty"`  // Flutter doctor output
}

// DartInfo contains information about a detected Dart SDK installation.
type DartInfo struct {
	Path      string `json:"path"`              // Path to Dart SDK
	Version   string `json:"version"`           // Dart version (e.g., "3.5.0")
	IsValid   bool   `json:"isValid"`           // Whether the installation is valid
	IsBundled bool   `json:"isBundled"`         // Whether this is the Dart bundled with Flutter
	Flutter   string `json:"flutter,omitempty"` // Path to Flutter SDK if bundled
	Error     string `json:"error,omitempty"`   // Error message if not valid
}

// FlutterProjectInfo contains information about a detected Flutter project.
type FlutterProjectInfo struct {
	Path         string   `json:"path"`            // Path to project root
	Name         string   `json:"name"`            // Project name from pubspec.yaml
	Version      string   `json:"version"`         // Project version
	FlutterSDK   string   `json:"flutterSdk"`      // Flutter SDK constraint
	DartSDK      string   `json:"dartSdk"`         // Dart SDK constraint
	Dependencies []string `json:"dependencies"`    // List of dependencies
	IsValid      bool     `json:"isValid"`         // Whether this is a valid Flutter project
	Error        string   `json:"error,omitempty"` // Error message if not valid
}

// DoctorCategory represents a category from flutter doctor output.
type DoctorCategory struct {
	Name    string `json:"name"`    // Category name (e.g., "Flutter", "Android toolchain")
	Status  string `json:"status"`  // Status (ok, partial, missing)
	Message string `json:"message"` // Human-readable message
}

// FlutterDoctor contains the parsed output from flutter doctor --machine.
type FlutterDoctor struct {
	Categories []DoctorCategory `json:"categories"` // Doctor categories
	Issues     []string         `json:"issues"`     // List of issues found
}

// FlutterCacheInfo contains information about Flutter cache locations.
type FlutterCacheInfo struct {
	BuildCache  string `json:"buildCache"`  // Path to build cache
	PubCache    string `json:"pubCache"`    // Path to pub cache
	GradleCache string `json:"gradleCache"` // Path to Gradle cache in Flutter
}

// EnvironmentReport contains a comprehensive report of all detected environment info.
type EnvironmentReport struct {
	Timestamp time.Time    `json:"timestamp"`           // Report generation time
	JDK       *JDKInfo     `json:"jdk,omitempty"`       // JDK info (nil if not found)
	SDK       *SDKInfo     `json:"sdk,omitempty"`       // Android SDK info (nil if not found)
	Node      *NodeInfo    `json:"node,omitempty"`      // Node.js info (nil if not found)
	Gradle    *GradleInfo  `json:"gradle,omitempty"`    // Gradle wrapper info (nil if not found)
	Expo      *ExpoInfo    `json:"expo,omitempty"`      // Expo project info (nil if not detected)
	Flutter   *FlutterInfo `json:"flutter,omitempty"`   // Flutter SDK info (nil if not found)
	Dart      *DartInfo    `json:"dart,omitempty"`      // Dart SDK info (nil if not found)
	MiseTools []MiseInfo   `json:"miseTools,omitempty"` // Tools installed via mise
	Warnings  []string     `json:"warnings"`            // Warnings about environment
	Errors    []string     `json:"errors"`              // Errors encountered during detection
}
