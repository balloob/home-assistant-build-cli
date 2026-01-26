package cmd

import (
	"fmt"
	"os"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/update"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	updateForce bool
	updateCheck bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update hab to the latest version",
	Long: `Check for and install the latest version of hab from GitHub releases.

By default, this command will download and install the latest release if a newer
version is available. Use --check to only check without installing.`,
	RunE:    runUpdate,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Force update even if already on latest version")
	updateCmd.Flags().BoolVarP(&updateCheck, "check", "c", false, "Only check for updates without installing")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	textMode := viper.GetBool("text")
	configDir := viper.GetString("config")

	// Check for latest release
	if textMode {
		fmt.Fprintln(os.Stderr, "Checking for updates...")
	}

	check, err := update.CheckForUpdate(configDir, Version)
	if err != nil {
		if textMode {
			fmt.Println(client.FormatErrorText("Failed to check for updates: "+err.Error(), "Check your internet connection"))
		} else {
			fmt.Println(client.FormatError("UPDATE_CHECK_FAILED", err.Error(), nil))
		}
		return nil
	}

	hasUpdate := update.HasUpdate(check)

	// Check-only mode
	if updateCheck {
		result := map[string]interface{}{
			"current_version": Version,
			"latest_version":  check.LatestVersion,
			"update_available": hasUpdate,
			"release_url":     check.ReleaseURL,
		}

		if textMode {
			if hasUpdate {
				fmt.Printf("Update available: %s -> %s\n", Version, check.LatestVersion)
				fmt.Printf("Release: %s\n", check.ReleaseURL)
			} else {
				fmt.Printf("You are on the latest version (%s)\n", Version)
			}
		} else {
			client.PrintSuccess(result, false, "")
		}
		return nil
	}

	// No update needed (unless forced)
	if !hasUpdate && !updateForce {
		if textMode {
			fmt.Printf("You are already on the latest version (%s)\n", Version)
		} else {
			client.PrintSuccess(map[string]interface{}{
				"current_version": Version,
				"latest_version":  check.LatestVersion,
				"updated":         false,
				"message":         "Already on latest version",
			}, false, "")
		}
		return nil
	}

	// Check if we have a download URL
	if check.DownloadURL == "" {
		errMsg := fmt.Sprintf("No binary available for your platform. Visit %s to download manually.", check.ReleaseURL)
		if textMode {
			fmt.Println(client.FormatErrorText(errMsg, ""))
		} else {
			fmt.Println(client.FormatError("NO_BINARY", errMsg, map[string]interface{}{
				"release_url": check.ReleaseURL,
			}))
		}
		return nil
	}

	// Download update
	if textMode {
		fmt.Fprintf(os.Stderr, "Downloading %s...\n", check.LatestVersion)
	}

	tmpPath, err := update.DownloadUpdate(check.DownloadURL)
	if err != nil {
		if textMode {
			fmt.Println(client.FormatErrorText("Failed to download update: "+err.Error(), ""))
		} else {
			fmt.Println(client.FormatError("DOWNLOAD_FAILED", err.Error(), nil))
		}
		return nil
	}

	// Install update
	if textMode {
		fmt.Fprintln(os.Stderr, "Installing update...")
	}

	if err := update.InstallUpdate(tmpPath); err != nil {
		os.Remove(tmpPath) // Clean up temp file
		if textMode {
			fmt.Println(client.FormatErrorText("Failed to install update: "+err.Error(), "You may need to run with elevated permissions"))
		} else {
			fmt.Println(client.FormatError("INSTALL_FAILED", err.Error(), nil))
		}
		return nil
	}

	// Success
	if textMode {
		fmt.Printf("Successfully updated to %s\n", check.LatestVersion)
	} else {
		client.PrintSuccess(map[string]interface{}{
			"previous_version": Version,
			"new_version":      check.LatestVersion,
			"updated":          true,
		}, false, "Successfully updated")
	}

	return nil
}
