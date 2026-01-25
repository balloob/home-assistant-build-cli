package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/home-assistant/hab/config"
)

const (
	// GitHubOwner is the GitHub repository owner
	GitHubOwner = "home-assistant"
	// GitHubRepo is the GitHub repository name
	GitHubRepo = "home-assistant-build-cli"
	// UpdateCheckFile stores the last update check info
	UpdateCheckFile = "update_check.json"
	// CheckInterval is how often to check for updates (24 hours)
	CheckInterval = 24 * time.Hour
)

// Release represents a GitHub release
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	HTMLURL     string  `json:"html_url"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateCheck stores the last update check info
type UpdateCheck struct {
	LastCheck      time.Time `json:"last_check"`
	LatestVersion  string    `json:"latest_version"`
	CurrentVersion string    `json:"current_version"`
	DownloadURL    string    `json:"download_url"`
	ReleaseURL     string    `json:"release_url"`
}

// getUpdateCheckPath returns the path to the update check file
func getUpdateCheckPath(configDir string) string {
	return filepath.Join(config.GetConfigDir(configDir), UpdateCheckFile)
}

// LoadUpdateCheck loads the last update check info from disk
func LoadUpdateCheck(configDir string) (*UpdateCheck, error) {
	path := getUpdateCheckPath(configDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var check UpdateCheck
	if err := json.Unmarshal(data, &check); err != nil {
		return nil, err
	}
	return &check, nil
}

// SaveUpdateCheck saves the update check info to disk
func SaveUpdateCheck(configDir string, check *UpdateCheck) error {
	if err := config.EnsureConfigDir(configDir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(check, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getUpdateCheckPath(configDir), data, 0600)
}

// NeedsCheck returns true if we should check for updates
func NeedsCheck(configDir string) bool {
	check, err := LoadUpdateCheck(configDir)
	if err != nil || check == nil {
		return true
	}
	return time.Since(check.LastCheck) >= CheckInterval
}

// GetLatestRelease fetches the latest release from GitHub
func GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GitHubOwner, GitHubRepo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "hab-cli")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases found")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// GetAssetForPlatform returns the download URL for the current platform
func GetAssetForPlatform(release *Release) (string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	suffix := ""
	if goos == "windows" {
		suffix = ".exe"
	}

	expectedName := fmt.Sprintf("hab-%s-%s%s", goos, goarch, suffix)

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("no binary found for %s/%s", goos, goarch)
}

// CheckForUpdate checks GitHub for a newer version
func CheckForUpdate(configDir string, currentVersion string) (*UpdateCheck, error) {
	release, err := GetLatestRelease()
	if err != nil {
		return nil, err
	}

	downloadURL, _ := GetAssetForPlatform(release) // Ignore error, URL might not exist

	check := &UpdateCheck{
		LastCheck:      time.Now(),
		LatestVersion:  release.TagName,
		CurrentVersion: currentVersion,
		DownloadURL:    downloadURL,
		ReleaseURL:     release.HTMLURL,
	}

	if err := SaveUpdateCheck(configDir, check); err != nil {
		// Don't fail if we can't save, just log it
		fmt.Fprintf(os.Stderr, "Warning: could not save update check: %v\n", err)
	}

	return check, nil
}

// HasUpdate returns true if there's a newer version available
func HasUpdate(check *UpdateCheck) bool {
	if check == nil || check.LatestVersion == "" || check.CurrentVersion == "" {
		return false
	}
	// Simple string comparison - versions should be semver tags like v1.2.3
	// Remove 'v' prefix for comparison if present
	latest := strings.TrimPrefix(check.LatestVersion, "v")
	current := strings.TrimPrefix(check.CurrentVersion, "v")

	// Skip if current version is "dev" or empty
	if current == "" || current == "dev" {
		return false
	}

	return latest != current && CompareVersions(latest, current) > 0
}

// CompareVersions compares two semver versions
// Returns: 1 if a > b, -1 if a < b, 0 if equal
func CompareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		var numA, numB int
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &numA)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &numB)
		}

		if numA > numB {
			return 1
		}
		if numA < numB {
			return -1
		}
	}
	return 0
}

// DownloadUpdate downloads the latest release binary
func DownloadUpdate(downloadURL string) (string, error) {
	if downloadURL == "" {
		return "", fmt.Errorf("no download URL available")
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "hab-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	tmpFile.Close()

	// Make executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			os.Remove(tmpPath)
			return "", fmt.Errorf("failed to make executable: %w", err)
		}
	}

	return tmpPath, nil
}

// InstallUpdate replaces the current binary with the new one
func InstallUpdate(newBinaryPath string) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks to get actual binary path
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// On Windows, we need to rename the old binary first
	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		os.Remove(oldPath) // Remove any previous .old file
		if err := os.Rename(execPath, oldPath); err != nil {
			return fmt.Errorf("failed to rename old binary: %w", err)
		}
	}

	// Move new binary to executable path
	if err := os.Rename(newBinaryPath, execPath); err != nil {
		// If rename fails (cross-device), try copy
		if err := copyFile(newBinaryPath, execPath); err != nil {
			return fmt.Errorf("failed to install new binary: %w", err)
		}
		os.Remove(newBinaryPath)
	}

	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}

// PrintUpdateNotice prints a notice about available updates to stderr
func PrintUpdateNotice(check *UpdateCheck) {
	fmt.Fprintf(os.Stderr, "\nA new version of hab is available: %s (current: %s)\n", check.LatestVersion, check.CurrentVersion)
	fmt.Fprintf(os.Stderr, "Run 'hab update' to update, or visit: %s\n\n", check.ReleaseURL)
}
