# mdev

CLI tool for managing mobile development environments (React Native, Flutter, etc.)

## Installation

### From Source

```bash
go install github.com/sombi/mobile-dev-helper@latest
```

### Using Make

```bash
make build
make install
```

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/sombi/mobile-dev-helper/releases).

## Usage

```bash
# Show help
mdev --help

# Diagnose environment issues
mdev doctor

# Clean caches
mdev clean all

# Show environment information
mdev info

# Repair broken tools
mdev repair
```

## Configuration

Configuration file: `~/.mdev.yaml`

Example:

```yaml
verbose: false
log_level: info
cache_dir: ~/.cache/mdev
project_path: ""
```

### Environment Variables

All configuration options can be set via environment variables with the `MDEV_` prefix:

- `MDEV_VERBOSE` - Enable verbose output
- `MDEV_LOG_LEVEL` - Set log level (debug, info, warn, error)
- `MDEV_CACHE_DIR` - Custom cache directory
- `MDEV_PROJECT_PATH` - Default project path

**Note:** Legacy `MOBILE_DEV_*` environment variables are still supported but deprecated.

## Features

- **Clean** - Remove caches and temporary files (Gradle, npm, Metro, CocoaPods)
- **Doctor** - Diagnose environment issues
- **Info** - Show environment information
- **Repair** - Fix broken development tools

## Supported Frameworks

- React Native
- Flutter
- Ionic
- Kotlin

## Development

```bash
# Run tests
make test

# Build for all platforms
make build-all

# Run linter
make lint

# Format code
make fmt
```

## Links

- React Native: https://reactnative.dev/docs/environment-setup
- Expo: https://docs.expo.dev/
- Flutter: https://docs.flutter.dev/
- Android SDK: https://developer.android.com/studio/command-line
