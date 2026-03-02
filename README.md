# mdev

CLI tool for managing mobile development environments (React Native, Flutter, etc.)

[![Release](https://img.shields.io/github/v/release/lucasvidela94/mdev-helper?include_prereleases)](https://github.com/lucasvidela94/mdev-helper/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/lucasvidela94/mdev-helper)](https://goreportcard.com/report/github.com/lucasvidela94/mdev-helper)

## Features

- **Doctor** - Comprehensive environment diagnostics with 5 checkers (disk, ports, versions, duplicates, performance)
- **Clean** - Remove caches and temporary files (Gradle, npm, Metro, Flutter, Android)
- **Emulator** - Manage Android emulators (list, status)
- **Logs** - Stream logs from Android, Metro, and Flutter
- **Config** - Interactive configuration wizard with auto-detection
- **Update** - Self-updating capability
- **Repair** - Fix broken development tools

## Installation

### macOS

```bash
curl -L https://github.com/lucasvidela94/mdev-helper/releases/latest/download/mdev-1.0.0-darwin-amd64.tar.gz | tar xz
sudo mv mdev /usr/local/bin/
```

### Linux

```bash
curl -L https://github.com/lucasvidela94/mdev-helper/releases/latest/download/mdev-1.0.0-linux-amd64.tar.gz | tar xz
sudo mv mdev /usr/local/bin/
```

### Windows

Download from [GitHub Releases](https://github.com/lucasvidela94/mdev-helper/releases) and add to PATH.

### From Source

```bash
go install github.com/lucasvidela94/mdev-helper@latest
```

## Quick Start

```bash
# Initialize configuration (interactive wizard)
mdev config init

# Check your environment
mdev doctor

# Check for updates
mdev update check
```

## Commands

### Doctor

Comprehensive environment diagnostics:

```bash
# Run all checks
mdev doctor

# JSON output for CI/CD
mdev doctor --format=json

# Verbose output
mdev doctor --verbose

# Exit codes: 0=ok, 1=warnings, 2=errors
```

Checks include:
- Disk space (< 10GB warning)
- Port availability (8081 Metro, 5037 ADB)
- Tool versions vs minimum requirements
- Duplicate SDK detection
- Performance recommendations

### Clean

Remove caches and temporary files:

```bash
# Clean everything
mdev clean all

# Clean specific caches
mdev clean gradle      # Gradle cache
mdev clean npm         # npm/node_modules
mdev clean metro       # Metro bundler cache
mdev clean flutter     # Flutter cache
mdev clean android     # Android build cache

# Clean node_modules and reinstall
mdev clean modules --reinstall

# Dry run (see what would be deleted)
mdev clean all --dry-run
```

### Config

Configuration management:

```bash
# Interactive setup wizard
mdev config init

# Get a config value
mdev config get android_home

# Set a config value
mdev config set android_home /path/to/sdk

# List all config
mdev config list
```

### Emulator

Manage Android emulators:

```bash
# List available AVDs
mdev emulator list

# Check emulator status
mdev emulator status
```

### Logs

Stream logs from your app:

```bash
# Android logs
mdev logs android

# Metro bundler logs (React Native)
mdev logs metro

# Flutter logs
mdev logs flutter

# Follow mode (real-time)
mdev logs android --follow

# Filter by package
mdev logs android --package com.example.app

# Export to file
mdev logs android --output logs.txt
```

### Update

Self-update capability:

```bash
# Check for updates
mdev update check

# Update to latest version
mdev update

# Show version info
mdev version
```

### Info

Show environment information:

```bash
mdev info
```

### Repair

Fix broken tools:

```bash
# Validate tools
mdev repair validate

# Fix gradlew permissions
mdev repair gradlew
```

## Configuration

Configuration file: `~/.mdev.yaml`

Example:

```yaml
verbose: false
android_home: /home/user/Android/Sdk
java_home: /usr/lib/jvm/java-17-openjdk
flutter_home: /home/user/flutter
```

### Environment Variables

All configuration options can be set via environment variables with the `MDEV_` prefix:

- `MDEV_VERBOSE` - Enable verbose output
- `MDEV_ANDROID_HOME` - Android SDK path
- `MDEV_JAVA_HOME` - JDK path
- `MDEV_FLUTTER_HOME` - Flutter SDK path

**Note:** Legacy `MOBILE_DEV_*` environment variables are still supported but deprecated.

## Supported Frameworks

- React Native (Expo and Bare workflow)
- Flutter
- Ionic
- Kotlin Multiplatform

## Development

```bash
# Clone the repository
git clone https://github.com/lucasvidela94/mdev-helper.git
cd mobile-dev-helper

# Install dependencies
go mod download

# Run tests
make test

# Build for current platform
make build

# Build for all platforms
make build-all

# Run linter
make lint

# Format code
make fmt
```

## Architecture

Built with **Spec-Driven Development (SDD)** methodology:

- 15 completed SDD changes
- 100+ test cases
- Modular detector pattern
- Cross-platform support

## CI/CD Integration

```bash
# Exit codes for automation
mdev doctor || echo "Environment issues detected"

# JSON output for parsing
mdev doctor --format=json | jq '.summary.status'
```

## Troubleshooting

### Doctor shows warnings

Run `mdev config init` to automatically detect and configure your SDKs.

### Update fails

Ensure you have write permissions to the installation directory:
```bash
sudo mdev update
```

### Logs not showing

Make sure your device/emulator is connected and authorized:
```bash
adb devices
```

## Links

- [React Native Environment Setup](https://reactnative.dev/docs/environment-setup)
- [Expo Documentation](https://docs.expo.dev/)
- [Flutter Documentation](https://docs.flutter.dev/)
- [Android SDK Command Line Tools](https://developer.android.com/studio/command-line)

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

---

Built with Go and the Spec-Driven Development methodology.
