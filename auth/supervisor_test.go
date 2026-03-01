package auth

import (
	"os"
	"testing"
)

func TestIsSupervisorEnvironment(t *testing.T) {
	// Save and restore original value
	original := os.Getenv(SupervisorTokenEnv)
	defer os.Setenv(SupervisorTokenEnv, original)

	// Test: not set
	os.Unsetenv(SupervisorTokenEnv)
	if IsSupervisorEnvironment() {
		t.Error("expected false when SUPERVISOR_TOKEN is not set")
	}

	// Test: set
	os.Setenv(SupervisorTokenEnv, "test-token")
	if !IsSupervisorEnvironment() {
		t.Error("expected true when SUPERVISOR_TOKEN is set")
	}

	// Test: empty string
	os.Setenv(SupervisorTokenEnv, "")
	if IsSupervisorEnvironment() {
		t.Error("expected false when SUPERVISOR_TOKEN is empty")
	}
}

func TestGetSupervisorToken(t *testing.T) {
	original := os.Getenv(SupervisorTokenEnv)
	defer os.Setenv(SupervisorTokenEnv, original)

	os.Setenv(SupervisorTokenEnv, "my-supervisor-token")
	token := GetSupervisorToken()
	if token != "my-supervisor-token" {
		t.Errorf("expected 'my-supervisor-token', got '%s'", token)
	}

	os.Unsetenv(SupervisorTokenEnv)
	token = GetSupervisorToken()
	if token != "" {
		t.Errorf("expected empty string, got '%s'", token)
	}
}

func TestLoadCredentials_SupervisorToken(t *testing.T) {
	// Save and restore original env vars
	origSupervisor := os.Getenv(SupervisorTokenEnv)
	origURL := os.Getenv("HAB_URL")
	origToken := os.Getenv("HAB_TOKEN")
	defer func() {
		os.Setenv(SupervisorTokenEnv, origSupervisor)
		os.Setenv("HAB_URL", origURL)
		os.Setenv("HAB_TOKEN", origToken)
	}()

	// Clear all env vars
	os.Unsetenv(SupervisorTokenEnv)
	os.Unsetenv("HAB_URL")
	os.Unsetenv("HAB_TOKEN")

	// Test: SUPERVISOR_TOKEN set
	os.Setenv(SupervisorTokenEnv, "sv-token-123")
	creds, err := LoadCredentials("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil credentials")
	}
	if creds.URL != SupervisorURL {
		t.Errorf("expected URL '%s', got '%s'", SupervisorURL, creds.URL)
	}
	if creds.AccessToken != "sv-token-123" {
		t.Errorf("expected token 'sv-token-123', got '%s'", creds.AccessToken)
	}
	if creds.IsOAuth() {
		t.Error("supervisor credentials should not be OAuth")
	}
	if creds.IsExpired() {
		t.Error("supervisor credentials should not be expired")
	}
	if creds.NeedsRefresh() {
		t.Error("supervisor credentials should not need refresh")
	}
}

func TestLoadCredentials_HABEnvTakesPriorityOverSupervisor(t *testing.T) {
	// Save and restore original env vars
	origSupervisor := os.Getenv(SupervisorTokenEnv)
	origURL := os.Getenv("HAB_URL")
	origToken := os.Getenv("HAB_TOKEN")
	defer func() {
		os.Setenv(SupervisorTokenEnv, origSupervisor)
		os.Setenv("HAB_URL", origURL)
		os.Setenv("HAB_TOKEN", origToken)
	}()

	// Set both SUPERVISOR_TOKEN and HAB_URL/HAB_TOKEN
	os.Setenv(SupervisorTokenEnv, "supervisor-token")
	os.Setenv("HAB_URL", "http://custom:8123")
	os.Setenv("HAB_TOKEN", "custom-token")

	creds, err := LoadCredentials("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil credentials")
	}
	// HAB_URL/HAB_TOKEN should take priority over SUPERVISOR_TOKEN
	if creds.URL != "http://custom:8123" {
		t.Errorf("expected URL 'http://custom:8123', got '%s'", creds.URL)
	}
	if creds.AccessToken != "custom-token" {
		t.Errorf("expected 'custom-token', got '%s'", creds.AccessToken)
	}
}
