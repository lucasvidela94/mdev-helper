// Package constants provides centralized application constants.
package constants

// App constants for the mdev CLI tool.
const (
	// AppName is the CLI binary name.
	AppName = "mdev"

	// ConfigFileName is the base name of the config file (without extension).
	ConfigFileName = ".mdev"

	// LegacyConfigFileName is the old config file name for migration.
	LegacyConfigFileName = ".mobile-dev-helper"

	// CacheDirName is the directory name for cache storage.
	CacheDirName = "mdev"

	// LegacyCacheDirName is the old cache directory name.
	LegacyCacheDirName = "mobile-dev-helper"

	// EnvPrefix is the environment variable prefix.
	EnvPrefix = "MDEV"

	// LegacyEnvPrefix is the old environment variable prefix for backward compatibility.
	LegacyEnvPrefix = "MOBILE_DEV"
)
