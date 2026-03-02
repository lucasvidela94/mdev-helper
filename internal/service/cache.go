// Package service provides cache cleaning functionality.
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

// CacheService handles cleaning of development caches.
type CacheService struct {
	projectPath string
}

// NewCacheService creates a new CacheService.
func NewCacheService(projectPath string) *CacheService {
	return &CacheService{
		projectPath: projectPath,
	}
}

// Clean cleans caches based on the target type.
func (s *CacheService) Clean(target string) ([]detector.CleanResult, error) {
	var results []detector.CleanResult

	switch target {
	case "all":
		results = append(results, s.cleanGradle()...)
		results = append(results, s.cleanNPM()...)
		results = append(results, s.cleanMetro()...)
		results = append(results, s.cleanPods()...)
		results = append(results, s.cleanAndroidBuild()...)
		results = append(results, s.cleanFlutter()...)
	case "gradle":
		results = append(results, s.cleanGradle()...)
	case "npm":
		results = append(results, s.cleanNPM()...)
	case "metro":
		results = append(results, s.cleanMetro()...)
	case "pods":
		results = append(results, s.cleanPods()...)
	case "android":
		results = append(results, s.cleanAndroidBuild()...)
	case "flutter":
		results = append(results, s.cleanFlutter()...)
	default:
		return nil, fmt.Errorf("unknown clean target: %s (use: all, gradle, npm, metro, pods, android, flutter)", target)
	}

	return results, nil
}

// cleanGradle cleans Gradle cache.
func (s *CacheService) cleanGradle() []detector.CleanResult {
	var results []detector.CleanResult

	gradlePaths := getGradleCachePaths()

	for _, path := range gradlePaths {
		result := cleanDirectory(path, "gradle")
		results = append(results, result)
	}

	return results
}

// cleanNPM cleans npm cache.
func (s *CacheService) cleanNPM() []detector.CleanResult {
	var results []detector.CleanResult

	// Global npm cache
	npmCachePath := filepath.Join(os.Getenv("HOME"), ".npm")
	result := cleanDirectory(npmCachePath, "npm")
	results = append(results, result)

	// Project node_modules (optional - usually in project)
	if s.projectPath != "" {
		nodeModulesPath := filepath.Join(s.projectPath, "node_modules")
		if _, err := os.Stat(nodeModulesPath); err == nil {
			result := cleanDirectory(nodeModulesPath, "node_modules")
			results = append(results, result)
		}
	}

	return results
}

// cleanMetro cleans Metro bundler cache.
func (s *CacheService) cleanMetro() []detector.CleanResult {
	var results []detector.CleanResult

	metroPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".cache", "metro"),
		filepath.Join(os.Getenv("HOME"), ".metro"),
	}

	if s.projectPath != "" {
		metroPaths = append(metroPaths, filepath.Join(s.projectPath, "node_modules", ".cache"))
	}

	for _, path := range metroPaths {
		result := cleanDirectory(path, "metro")
		results = append(results, result)
	}

	return results
}

// cleanPods cleans iOS Pods cache.
func (s *CacheService) cleanPods() []detector.CleanResult {
	var results []detector.CleanResult

	if runtime.GOOS != "darwin" {
		return results
	}

	podsPaths := []string{
		filepath.Join(os.Getenv("HOME"), "Library", "Caches", "CocoaPods"),
		filepath.Join(os.Getenv("HOME"), "Library", "Developer", "Xcode", "DerivedData"),
	}

	if s.projectPath != "" {
		podsPaths = append(podsPaths, filepath.Join(s.projectPath, "ios", "Pods"))
		podsPaths = append(podsPaths, filepath.Join(s.projectPath, "ios", "Podfile.lock"))
	}

	for _, path := range podsPaths {
		result := cleanDirectory(path, "pods")
		results = append(results, result)
	}

	return results
}

// cleanAndroidBuild cleans Android build cache.
func (s *CacheService) cleanAndroidBuild() []detector.CleanResult {
	var results []detector.CleanResult

	paths := []string{}

	if s.projectPath != "" {
		paths = append(paths, filepath.Join(s.projectPath, "android", ".gradle"))
		paths = append(paths, filepath.Join(s.projectPath, "android", "app", "build"))
		paths = append(paths, filepath.Join(s.projectPath, "build"))
	}

	// Global Android build cache
	home := os.Getenv("HOME")
	if home != "" {
		paths = append(paths, filepath.Join(home, ".gradle", "caches", "build-cache-1"))
	}

	for _, path := range paths {
		result := cleanDirectory(path, "android-build")
		results = append(results, result)
	}

	return results
}

// cleanDirectory removes a directory and returns a CleanResult.
func cleanDirectory(path, cacheType string) detector.CleanResult {
	result := detector.CleanResult{
		Path: path,
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			result.Success = true
			result.SizeFreed = "0 B"
			result.Path = path + " (not found)"
			return result
		}
		result.Success = false
		result.Error = err.Error()
		return result
	}

	// Only calculate size for existing directories
	if info.IsDir() {
		size := getDirSize(path)
		result.SizeFreed = formatSize(size)
	}

	// Remove directory
	if err := os.RemoveAll(path); err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	result.Success = true
	return result
}

// getGradleCachePaths returns all possible Gradle cache locations.
func getGradleCachePaths() []string {
	var paths []string
	home := os.Getenv("HOME")

	switch runtime.GOOS {
	case "linux", "darwin":
		paths = []string{
			filepath.Join(home, ".gradle", "caches"),
			filepath.Join(home, ".gradle", "daemon"),
		}
	case "windows":
		paths = []string{
			filepath.Join(os.Getenv("USERPROFILE"), ".gradle", "caches"),
			filepath.Join(os.Getenv("USERPROFILE"), ".gradle", "daemon"),
		}
	}

	return paths
}

// getDirSize calculates the size of a directory.
func getDirSize(path string) int64 {
	var size int64

	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size
}

// formatSize converts bytes to human-readable format.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// cleanFlutter cleans Flutter cache directories.
func (s *CacheService) cleanFlutter() []detector.CleanResult {
	var results []detector.CleanResult

	// Get Flutter build cache paths
	flutterPaths := getFlutterCachePaths()

	for _, path := range flutterPaths {
		result := cleanDirectory(path, "flutter")
		results = append(results, result)
	}

	// Clean project-specific Flutter build directories
	if s.projectPath != "" {
		projectPaths := []string{
			filepath.Join(s.projectPath, "build"),
			filepath.Join(s.projectPath, ".dart_tool"),
		}

		for _, path := range projectPaths {
			result := cleanDirectory(path, "flutter-build")
			results = append(results, result)
		}
	}

	return results
}

// getFlutterCachePaths returns all possible Flutter cache locations.
func getFlutterCachePaths() []string {
	var paths []string
	home := os.Getenv("HOME")

	switch runtime.GOOS {
	case "linux", "darwin":
		paths = []string{
			filepath.Join(home, ".pub-cache"),
			filepath.Join(home, ".flutter"),
		}
	case "windows":
		paths = []string{
			filepath.Join(os.Getenv("APPDATA"), "Pub", "Cache"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "flutter"),
		}
	}

	return paths
}
