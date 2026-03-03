package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsOAuth(t *testing.T) {
	tests := []struct {
		name string
		cred Credentials
		want bool
	}{
		{"with refresh token", Credentials{RefreshToken: "abc"}, true},
		{"without refresh token", Credentials{AccessToken: "tok"}, false},
		{"empty", Credentials{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cred.IsOAuth(); got != tt.want {
				t.Errorf("IsOAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasValidToken(t *testing.T) {
	tests := []struct {
		name string
		cred Credentials
		want bool
	}{
		{"has token", Credentials{AccessToken: "tok"}, true},
		{"no token", Credentials{}, false},
		{"empty token", Credentials{AccessToken: ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cred.HasValidToken(); got != tt.want {
				t.Errorf("HasValidToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name string
		cred Credentials
		want bool
	}{
		{
			"long-lived token never expires",
			Credentials{AccessToken: "tok"},
			false,
		},
		{
			"oauth with no expiry info",
			Credentials{RefreshToken: "rt", TokenExpiry: 0},
			true,
		},
		{
			"oauth expired",
			Credentials{RefreshToken: "rt", TokenExpiry: float64(time.Now().Unix() - 3600)},
			true,
		},
		{
			"oauth not expired",
			Credentials{RefreshToken: "rt", TokenExpiry: float64(time.Now().Unix() + 3600)},
			false,
		},
		{
			"oauth exactly at expiry",
			Credentials{RefreshToken: "rt", TokenExpiry: float64(time.Now().Unix())},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cred.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNeedsRefresh(t *testing.T) {
	tests := []struct {
		name string
		cred Credentials
		want bool
	}{
		{
			"long-lived token never needs refresh",
			Credentials{AccessToken: "tok"},
			false,
		},
		{
			"oauth with no expiry info",
			Credentials{RefreshToken: "rt", TokenExpiry: 0},
			true,
		},
		{
			"oauth within 5 min of expiry",
			Credentials{RefreshToken: "rt", TokenExpiry: float64(time.Now().Unix() + 200)},
			true,
		},
		{
			"oauth well before expiry",
			Credentials{RefreshToken: "rt", TokenExpiry: float64(time.Now().Unix() + 3600)},
			false,
		},
		{
			"oauth already expired",
			Credentials{RefreshToken: "rt", TokenExpiry: float64(time.Now().Unix() - 100)},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cred.NeedsRefresh(); got != tt.want {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadCredentials_EnvVars(t *testing.T) {
	// Save and restore original env vars
	origURL := os.Getenv("HAB_URL")
	origToken := os.Getenv("HAB_TOKEN")
	origRefresh := os.Getenv("HAB_REFRESH_TOKEN")
	origSupervisor := os.Getenv(SupervisorTokenEnv)
	defer func() {
		os.Setenv("HAB_URL", origURL)
		os.Setenv("HAB_TOKEN", origToken)
		os.Setenv("HAB_REFRESH_TOKEN", origRefresh)
		os.Setenv(SupervisorTokenEnv, origSupervisor)
	}()

	// Clear all env vars
	os.Unsetenv("HAB_URL")
	os.Unsetenv("HAB_TOKEN")
	os.Unsetenv("HAB_REFRESH_TOKEN")
	os.Unsetenv(SupervisorTokenEnv)

	t.Run("HAB_URL and HAB_TOKEN", func(t *testing.T) {
		os.Setenv("HAB_URL", "http://ha:8123")
		os.Setenv("HAB_TOKEN", "my-token")
		defer os.Unsetenv("HAB_URL")
		defer os.Unsetenv("HAB_TOKEN")

		creds, err := LoadCredentials("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if creds == nil {
			t.Fatal("expected non-nil credentials")
		}
		if creds.URL != "http://ha:8123" {
			t.Errorf("URL = %q, want %q", creds.URL, "http://ha:8123")
		}
		if creds.AccessToken != "my-token" {
			t.Errorf("AccessToken = %q, want %q", creds.AccessToken, "my-token")
		}
		if creds.IsOAuth() {
			t.Error("should not be OAuth")
		}
	})

	t.Run("HAB_URL and HAB_REFRESH_TOKEN", func(t *testing.T) {
		os.Setenv("HAB_URL", "http://ha:8123")
		os.Setenv("HAB_REFRESH_TOKEN", "my-refresh")
		os.Unsetenv("HAB_TOKEN")
		defer os.Unsetenv("HAB_URL")
		defer os.Unsetenv("HAB_REFRESH_TOKEN")

		creds, err := LoadCredentials("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if creds == nil {
			t.Fatal("expected non-nil credentials")
		}
		if creds.URL != "http://ha:8123" {
			t.Errorf("URL = %q, want %q", creds.URL, "http://ha:8123")
		}
		if creds.RefreshToken != "my-refresh" {
			t.Errorf("RefreshToken = %q, want %q", creds.RefreshToken, "my-refresh")
		}
		if !creds.IsOAuth() {
			t.Error("should be OAuth")
		}
	})

	t.Run("no env vars and no file returns nil", func(t *testing.T) {
		os.Unsetenv("HAB_URL")
		os.Unsetenv("HAB_TOKEN")
		os.Unsetenv("HAB_REFRESH_TOKEN")
		os.Unsetenv(SupervisorTokenEnv)

		// Use a temp dir that won't have credentials
		tmpDir := t.TempDir()
		creds, err := LoadCredentials(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if creds != nil {
			t.Errorf("expected nil credentials, got %+v", creds)
		}
	})
}

func TestSaveAndDeleteCredentials(t *testing.T) {
	tmpDir := t.TempDir()

	creds := &Credentials{
		URL:         "http://localhost:8123",
		AccessToken: "test-token",
	}

	// Save
	err := SaveCredentials(creds, tmpDir)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	// Verify file exists (credentials.json inside the config dir)
	credsPath := filepath.Join(tmpDir, "credentials.json")
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		t.Errorf("credentials file was not created at %s", credsPath)
	}

	// Delete
	deleted := DeleteCredentials(tmpDir)
	if !deleted {
		t.Error("DeleteCredentials returned false")
	}

	// Verify file removed
	if _, err := os.Stat(credsPath); !os.IsNotExist(err) {
		t.Error("credentials file was not deleted")
	}

	// Delete again (should return false since file is gone)
	deleted2 := DeleteCredentials(tmpDir)
	if deleted2 {
		t.Error("DeleteCredentials should return false when file doesn't exist")
	}
}

func TestSaveAndLoadCredentials_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	// Clear env vars that would shortcut LoadCredentials
	origURL := os.Getenv("HAB_URL")
	origToken := os.Getenv("HAB_TOKEN")
	origRefresh := os.Getenv("HAB_REFRESH_TOKEN")
	origSupervisor := os.Getenv(SupervisorTokenEnv)
	defer func() {
		os.Setenv("HAB_URL", origURL)
		os.Setenv("HAB_TOKEN", origToken)
		os.Setenv("HAB_REFRESH_TOKEN", origRefresh)
		os.Setenv(SupervisorTokenEnv, origSupervisor)
	}()
	os.Unsetenv("HAB_URL")
	os.Unsetenv("HAB_TOKEN")
	os.Unsetenv("HAB_REFRESH_TOKEN")
	os.Unsetenv(SupervisorTokenEnv)

	original := &Credentials{
		URL:          "http://my-ha.local:8123",
		ClientID:     "http://my-ha.local:8123/",
		AccessToken:  "access-tok-123",
		RefreshToken: "refresh-tok-456",
		TokenExpiry:  float64(time.Now().Unix() + 3600),
	}

	// Save
	if err := SaveCredentials(original, tmpDir); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	// Load back
	loaded, err := LoadCredentials(tmpDir)
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadCredentials returned nil")
	}

	if loaded.URL != original.URL {
		t.Errorf("URL = %q, want %q", loaded.URL, original.URL)
	}
	if loaded.ClientID != original.ClientID {
		t.Errorf("ClientID = %q, want %q", loaded.ClientID, original.ClientID)
	}
	if loaded.AccessToken != original.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, original.AccessToken)
	}
	if loaded.RefreshToken != original.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, original.RefreshToken)
	}
	if loaded.TokenExpiry != original.TokenExpiry {
		t.Errorf("TokenExpiry = %v, want %v", loaded.TokenExpiry, original.TokenExpiry)
	}
}
