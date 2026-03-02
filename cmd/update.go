// Package cmd contains the CLI commands.
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sombi/mobile-dev-helper/internal/updater"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command group.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update mdev to the latest version",
	Long: `Check for and install updates to the mdev CLI.

This command group allows you to check for available updates and
install the latest version from GitHub releases.`,
}

// updateCheckCmd checks for available updates.
var updateCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for available updates",
	Long: `Check if a newer version of mdev is available.

Queries the GitHub releases API to find the latest version and
compares it with the currently installed version.`,
	Example: `  # Check for updates
  mdev update check`,
	RunE: runUpdateCheck,
}

// updateInstallCmd installs the latest update.
var updateInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the latest version",
	Long: `Download and install the latest version of mdev.

This command will:
1. Check for the latest release on GitHub
2. Display the changelog
3. Ask for confirmation (unless --yes flag is used)
4. Download the new binary with progress indication
5. Create a backup of the current binary
6. Atomically replace the current binary
7. Verify the update succeeded

If the update fails, the original binary will be restored automatically.`,
	Example: `  # Install latest update (with confirmation)
  mdev update install

  # Install latest update without confirmation
  mdev update install --yes`,
	RunE: runUpdateInstall,
}

var updateYesFlag bool

func init() {
	updateInstallCmd.Flags().BoolVarP(&updateYesFlag, "yes", "y", false, "Skip confirmation prompt")

	updateCmd.AddCommand(updateCheckCmd)
	updateCmd.AddCommand(updateInstallCmd)
	rootCmd.AddCommand(updateCmd)
}

func runUpdateCheck(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	u := updater.NewUpdater()

	fmt.Println("Checking for updates...")
	fmt.Println()

	result, err := u.Check(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if result.UpdateAvailable {
		fmt.Printf("📦 A new version is available!\n\n")
		fmt.Printf("   Current version: %s\n", result.CurrentVersion)
		fmt.Printf("   Latest version:  %s\n", result.LatestVersion)
		fmt.Printf("   Release URL:     %s\n", result.ReleaseURL)
		fmt.Println()
		fmt.Println("   Run 'mdev update install' to update.")
	} else {
		fmt.Printf("✅ You're up to date!\n\n")
		fmt.Printf("   Current version: %s\n", result.CurrentVersion)

		if updater.IsPrerelease(result.CurrentVersion) {
			fmt.Println()
			fmt.Println("   ℹ Note: You have a pre-release version installed.")
		}
	}

	return nil
}

func runUpdateInstall(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	u := updater.NewUpdater()

	fmt.Println("Checking for updates...")
	fmt.Println()

	// Check for updates
	checkResult, err := u.Check(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !checkResult.UpdateAvailable {
		fmt.Printf("✅ You're already on the latest version (%s).\n", checkResult.CurrentVersion)
		return nil
	}

	// Display update information
	fmt.Printf("📦 Update Available\n\n")
	fmt.Printf("   Current version: %s\n", checkResult.CurrentVersion)
	fmt.Printf("   Latest version:  %s\n", checkResult.LatestVersion)
	fmt.Printf("   Published:       %s\n", checkResult.ReleaseInfo.PublishedAt.Format(time.RFC3339))
	fmt.Println()

	// Check if we have an asset for this platform
	if checkResult.ReleaseInfo.Asset == nil {
		fmt.Println("⚠️  No binary available for your platform.")
		fmt.Printf("   Please download manually from: %s\n", checkResult.ReleaseURL)
		return nil
	}

	asset := checkResult.ReleaseInfo.Asset
	fmt.Printf("   Download size:   %s\n", updater.FormatBytes(asset.Size))
	fmt.Println()

	// Display changelog (truncated if too long)
	if checkResult.ReleaseInfo.Changelog != "" {
		fmt.Println("📝 Changelog:")
		fmt.Println(strings.Repeat("─", 50))
		changelog := checkResult.ReleaseInfo.Changelog
		if len(changelog) > 1000 {
			changelog = changelog[:1000] + "\n\n... (truncated)"
		}
		fmt.Println(changelog)
		fmt.Println(strings.Repeat("─", 50))
		fmt.Println()
	}

	// Ask for confirmation unless --yes flag is used
	if !updateYesFlag {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to install this update? [y/N]: ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Update cancelled.")
			return nil
		}
		fmt.Println()
	}

	// Download the update
	fmt.Println("⬇️  Downloading update...")

	downloader := updater.NewDownloader()

	// Progress callback
	progressCallback := func(downloaded, total int64, speed float64) {
		if total > 0 {
			percent := float64(downloaded) * 100 / float64(total)
			fmt.Printf("\r   Progress: %.1f%% (%s / %s) at %s",
				percent,
				updater.FormatBytes(downloaded),
				updater.FormatBytes(total),
				updater.FormatSpeed(speed),
			)
		} else {
			fmt.Printf("\r   Downloaded: %s at %s",
				updater.FormatBytes(downloaded),
				updater.FormatSpeed(speed),
			)
		}
	}

	downloadResult, err := downloader.Download(ctx, updater.DownloadOptions{
		Asset:            asset,
		ProgressCallback: progressCallback,
	})
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer downloader.Cleanup(downloadResult.Path)

	fmt.Println() // New line after progress
	fmt.Printf("   Downloaded %s in %s\n", updater.FormatBytes(downloadResult.Size), downloadResult.Duration.Round(time.Millisecond))
	fmt.Println()

	// Verify the downloaded binary
	fmt.Println("🔍 Verifying download...")
	if err := downloader.VerifyBinary(downloadResult.Path, 1024*1024); err != nil { // Min 1MB
		return fmt.Errorf("download verification failed: %w", err)
	}
	fmt.Println("   ✓ Download verified")
	fmt.Println()

	// Apply the update
	fmt.Println("📦 Installing update...")
	if err := downloader.ApplyUpdate(downloadResult.Path); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}
	fmt.Println("   ✓ Update installed successfully")
	fmt.Println()

	fmt.Printf("🎉 Successfully updated to version %s!\n", checkResult.LatestVersion)
	fmt.Println()
	fmt.Println("   Run 'mdev version' to verify the installation.")

	return nil
}
