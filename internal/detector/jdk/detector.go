// Package jdk provides JDK detection functionality.
package jdk

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

// Standard paths for JDK installations by OS
var standardPaths = func() []string {
	var paths []string
	home := os.Getenv("HOME")

	switch runtime.GOOS {
	case "linux":
		paths = []string{
			"/usr/lib/jvm",
			"/opt/jdk",
			filepath.Join(home, ".jdks"),
			"/usr/lib/jvm",
		}
	case "darwin":
		paths = []string{
			"/Library/Java/JavaVirtualMachines",
			filepath.Join(home, "Library/Java/JavaVirtualMachines"),
			"/opt/jdk",
		}
	case "windows":
		paths = []string{
			"C:\\Program Files\\Java",
			"C:\\Program Files (x86)\\Java",
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Java"),
		}
	}
	return paths
}()

// Detect searches for JDK installations on the system.
func Detect() *detector.JDKInfo {
	// Priority 1: Check JAVA_HOME
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		info := checkJDKPath(javaHome)
		if info != nil {
			info.JavaHome = javaHome
			return info
		}
	}

	// Priority 2: Check standard directories
	for _, basePath := range standardPaths {
		info := findJDKInDirectory(basePath)
		if info != nil {
			info.JavaHome = javaHome
			if javaHome == "" {
				info.Warning = "JAVA_HOME not set"
			}
			return info
		}
	}

	// Priority 3: Try which java
	return detectFromPath()
}

// checkJDKPath checks if a given path contains a valid JDK.
func checkJDKPath(path string) *detector.JDKInfo {
	javaExec := filepath.Join(path, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExec += ".exe"
	}

	if _, err := os.Stat(javaExec); err != nil {
		return nil
	}

	info := &detector.JDKInfo{
		Path:    path,
		IsValid: true,
	}

	// Get version
	version, err := getJavaVersion(javaExec)
	if err == nil {
		info.Version = version.full
		info.MajorVersion = version.major
		info.Vendor = version.vendor
	}

	return info
}

// findJDKInDirectory searches for a JDK in a given directory.
func findJDKInDirectory(dir string) *detector.JDKInfo {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	// Look for directories that might be JDK installations
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		info := checkJDKPath(path)
		if info != nil {
			return info
		}
	}

	return nil
}

// detectFromPath tries to find java in PATH.
func detectFromPath() *detector.JDKInfo {
	javaPath, err := exec.LookPath("java")
	if err != nil {
		return &detector.JDKInfo{
			IsValid: false,
			Error:   "No JDK found. Install JDK or set JAVA_HOME",
		}
	}

	// Resolve symlinks to get actual path
	realPath, err := filepath.EvalSymlinks(javaPath)
	if err != nil {
		realPath = javaPath
	}

	// Get the JDK home from the java binary path
	jdkHome := filepath.Dir(filepath.Dir(realPath)) // bin/java -> bin/ -> jdk/

	info := checkJDKPath(jdkHome)
	if info != nil {
		info.JavaHome = os.Getenv("JAVA_HOME")
		if info.JavaHome == "" {
			info.Warning = "JAVA_HOME not set"
		}
		return info
	}

	// Fallback: just return path info
	return &detector.JDKInfo{
		Path:    filepath.Dir(realPath),
		Version: "unknown",
		IsValid: true,
	}
}

// versionInfo holds parsed version information
type versionInfo struct {
	full   string
	major  string
	vendor string
}

// getJavaVersion runs `java -version` and parses the output.
func getJavaVersion(javaPath string) (versionInfo, error) {
	cmd := exec.Command(javaPath, "-version")
	output, err := cmd.StderrPipe()
	if err != nil {
		return versionInfo{}, err
	}

	if err := cmd.Start(); err != nil {
		return versionInfo{}, err
	}

	scanner := bufio.NewScanner(output)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return versionInfo{}, err
	}
	cmd.Wait()

	if len(lines) < 1 {
		return versionInfo{}, fmt.Errorf("no version output")
	}

	// Parse first line (e.g., "openjdk version "17.0.2" 2022-01-18" or "java version "11.0.1"")
	firstLine := lines[0]

	var full, major, vendor string

	// Try to extract version
	if strings.Contains(firstLine, "openjdk") {
		vendor = "OpenJDK"
		// Format: openjdk version "17.0.2" 2022-01-18
		parts := strings.Fields(firstLine)
		for i, part := range parts {
			if strings.Contains(part, `"`) {
				full = strings.Trim(part, `"`)
				// Extract major version
				if strings.Contains(full, ".") {
					major = strings.Split(full, ".")[0]
				} else {
					major = full
				}
				// Handle "1.8.0" format -> 8
				if major == "1" && len(parts) > i+1 {
					next := strings.Trim(parts[i+1], `"`)
					if strings.HasPrefix(next, "8") {
						major = "8"
					}
				}
				break
			}
		}
	} else if strings.Contains(firstLine, "java version") {
		parts := strings.Fields(firstLine)
		for _, part := range parts {
			if strings.Contains(part, `"`) {
				full = strings.Trim(part, `"`)
				if strings.HasPrefix(full, "1.") {
					// Java 1.8 -> 8
					parts := strings.Split(full, ".")
					if len(parts) >= 2 {
						major = parts[1]
					}
				} else {
					major = strings.Split(full, ".")[0]
				}
				break
			}
		}
	}

	if full == "" {
		return versionInfo{}, fmt.Errorf("could not parse version from: %s", firstLine)
	}

	return versionInfo{full: full, major: major, vendor: vendor}, nil
}
