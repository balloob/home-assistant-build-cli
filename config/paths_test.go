package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigDir_ExplicitPath(t *testing.T) {
	dir := GetConfigDir("/custom/path")
	if dir != "/custom/path" {
		t.Errorf("GetConfigDir(/custom/path) = %q, want %q", dir, "/custom/path")
	}
}

func TestGetConfigDir_XDGConfigHome(t *testing.T) {
	// Save and restore env
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", orig)

	os.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")
	dir := GetConfigDir("")
	want := filepath.Join("/tmp/xdg-test", DefaultConfigDir)
	if dir != want {
		t.Errorf("GetConfigDir('') with XDG = %q, want %q", dir, want)
	}
}

func TestGetConfigDir_DefaultFallback(t *testing.T) {
	// Unset XDG_CONFIG_HOME to test fallback
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", orig)
	os.Unsetenv("XDG_CONFIG_HOME")

	dir := GetConfigDir("")
	// Should contain .config/home-assistant-builder
	if !strings.HasSuffix(dir, filepath.Join(".config", DefaultConfigDir)) {
		t.Errorf("GetConfigDir('') = %q, expected suffix %q", dir, filepath.Join(".config", DefaultConfigDir))
	}
}

func TestGetCredentialsPath(t *testing.T) {
	path := GetCredentialsPath("/my/config")
	want := filepath.Join("/my/config", CredentialsFile)
	if path != want {
		t.Errorf("GetCredentialsPath = %q, want %q", path, want)
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath("/my/config")
	want := filepath.Join("/my/config", ConfigFile)
	if path != want {
		t.Errorf("GetConfigPath = %q, want %q", path, want)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub", "nested")

	err := EnsureConfigDir(target)
	if err != nil {
		t.Fatalf("EnsureConfigDir failed: %v", err)
	}

	// The function calls GetConfigDir(target) which returns target as-is since it's non-empty
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory")
	}
	// Check permissions (Unix only)
	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("permissions = %o, want 0700", perm)
	}
}

func TestEnsureConfigDir_Idempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "existing")

	// Create it first
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Should not error when called again
	err := EnsureConfigDir(target)
	if err != nil {
		t.Fatalf("EnsureConfigDir on existing dir failed: %v", err)
	}
}

func TestConstants(t *testing.T) {
	if DefaultConfigDir != "home-assistant-builder" {
		t.Errorf("DefaultConfigDir = %q", DefaultConfigDir)
	}
	if CredentialsFile != "credentials.json" {
		t.Errorf("CredentialsFile = %q", CredentialsFile)
	}
	if ConfigFile != "config.json" {
		t.Errorf("ConfigFile = %q", ConfigFile)
	}
}
