package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},
		{"1.10.0", "1.9.0", 1},
		{"1.2.3", "1.2.3", 0},
		{"0.0.1", "0.0.0", 1},
		// Different length versions
		{"1.0", "1.0.0", 0},
		{"1.0.1", "1.0", 1},
		{"1", "1.0.0", 0},
		{"2", "1.9.9", 1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := CompareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestHasUpdate(t *testing.T) {
	tests := []struct {
		name  string
		check *UpdateCheck
		want  bool
	}{
		{"nil check", nil, false},
		{"empty versions", &UpdateCheck{}, false},
		{"same version", &UpdateCheck{LatestVersion: "v1.0.0", CurrentVersion: "v1.0.0"}, false},
		{"newer available", &UpdateCheck{LatestVersion: "v1.1.0", CurrentVersion: "v1.0.0"}, true},
		{"older available", &UpdateCheck{LatestVersion: "v1.0.0", CurrentVersion: "v1.1.0"}, false},
		{"dev version", &UpdateCheck{LatestVersion: "v1.0.0", CurrentVersion: "dev"}, false},
		{"empty current", &UpdateCheck{LatestVersion: "v1.0.0", CurrentVersion: ""}, false},
		{"with v prefix", &UpdateCheck{LatestVersion: "v2.0.0", CurrentVersion: "v1.0.0"}, true},
		{"without v prefix", &UpdateCheck{LatestVersion: "2.0.0", CurrentVersion: "1.0.0"}, true},
		{"mixed prefix", &UpdateCheck{LatestVersion: "v2.0.0", CurrentVersion: "1.0.0"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasUpdate(tt.check)
			if got != tt.want {
				t.Errorf("HasUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAssetForPlatform(t *testing.T) {
	release := &Release{
		Assets: []Asset{
			{Name: "hab-linux-amd64", BrowserDownloadURL: "https://example.com/hab-linux-amd64"},
			{Name: "hab-darwin-arm64", BrowserDownloadURL: "https://example.com/hab-darwin-arm64"},
			{Name: "hab-windows-amd64.exe", BrowserDownloadURL: "https://example.com/hab-windows-amd64.exe"},
		},
	}

	// We can only test for the current platform
	url, err := GetAssetForPlatform(release)
	// If our platform is in the list, it should work
	if err != nil {
		// It's ok if the current platform isn't in our test list
		t.Logf("GetAssetForPlatform returned error (expected if platform not in test data): %v", err)
	} else if url == "" {
		t.Error("expected non-empty URL")
	}

	// Test with empty assets
	emptyRelease := &Release{Assets: []Asset{}}
	_, err = GetAssetForPlatform(emptyRelease)
	if err == nil {
		t.Error("expected error for empty assets")
	}
}

func TestSaveAndLoadUpdateCheck(t *testing.T) {
	tmpDir := t.TempDir()

	check := &UpdateCheck{
		LastCheck:      time.Now().Truncate(time.Second),
		LatestVersion:  "v1.2.3",
		CurrentVersion: "v1.0.0",
		DownloadURL:    "https://example.com/download",
		ReleaseURL:     "https://example.com/release",
	}

	// Save
	err := SaveUpdateCheck(tmpDir, check)
	if err != nil {
		t.Fatalf("SaveUpdateCheck: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, UpdateCheckFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("update check file was not created")
	}

	// Load
	loaded, err := LoadUpdateCheck(tmpDir)
	if err != nil {
		t.Fatalf("LoadUpdateCheck: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadUpdateCheck returned nil")
	}
	if loaded.LatestVersion != check.LatestVersion {
		t.Errorf("LatestVersion = %q, want %q", loaded.LatestVersion, check.LatestVersion)
	}
	if loaded.CurrentVersion != check.CurrentVersion {
		t.Errorf("CurrentVersion = %q, want %q", loaded.CurrentVersion, check.CurrentVersion)
	}
	if loaded.DownloadURL != check.DownloadURL {
		t.Errorf("DownloadURL = %q, want %q", loaded.DownloadURL, check.DownloadURL)
	}
	if loaded.ReleaseURL != check.ReleaseURL {
		t.Errorf("ReleaseURL = %q, want %q", loaded.ReleaseURL, check.ReleaseURL)
	}
}

func TestLoadUpdateCheck_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	check, err := LoadUpdateCheck(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if check != nil {
		t.Errorf("expected nil, got %+v", check)
	}
}

func TestNeedsCheck(t *testing.T) {
	t.Run("no file means needs check", func(t *testing.T) {
		tmpDir := t.TempDir()
		if !NeedsCheck(tmpDir) {
			t.Error("expected NeedsCheck=true when no check file exists")
		}
	})

	t.Run("recent check means no need", func(t *testing.T) {
		tmpDir := t.TempDir()
		check := &UpdateCheck{
			LastCheck:      time.Now(),
			LatestVersion:  "v1.0.0",
			CurrentVersion: "v1.0.0",
		}
		if err := SaveUpdateCheck(tmpDir, check); err != nil {
			t.Fatalf("SaveUpdateCheck: %v", err)
		}
		if NeedsCheck(tmpDir) {
			t.Error("expected NeedsCheck=false for recent check")
		}
	})

	t.Run("old check means needs check", func(t *testing.T) {
		tmpDir := t.TempDir()
		check := &UpdateCheck{
			LastCheck:      time.Now().Add(-48 * time.Hour),
			LatestVersion:  "v1.0.0",
			CurrentVersion: "v1.0.0",
		}
		if err := SaveUpdateCheck(tmpDir, check); err != nil {
			t.Fatalf("SaveUpdateCheck: %v", err)
		}
		if !NeedsCheck(tmpDir) {
			t.Error("expected NeedsCheck=true for old check")
		}
	})
}
