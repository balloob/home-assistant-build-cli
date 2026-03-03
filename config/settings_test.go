package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestInitDefaults(t *testing.T) {
	// Reset viper state
	v := viper.New()
	viper.Reset()

	InitDefaults()

	if !viper.GetBool("text") {
		t.Error("expected text default to be true")
	}
	if viper.GetBool("verbose") {
		t.Error("expected verbose default to be false")
	}

	// Cleanup: use the new viper to avoid contamination
	_ = v
}

func TestGetSettings(t *testing.T) {
	viper.Reset()
	InitDefaults()

	viper.Set("url", "http://localhost:8123")
	viper.Set("config", "/tmp/test-config")

	settings := GetSettings()

	if settings.URL != "http://localhost:8123" {
		t.Errorf("URL = %q, want %q", settings.URL, "http://localhost:8123")
	}
	if !settings.TextMode {
		t.Error("expected TextMode=true (default)")
	}
	if settings.Verbose {
		t.Error("expected Verbose=false (default)")
	}
	if settings.ConfigDir != "/tmp/test-config" {
		t.Errorf("ConfigDir = %q, want %q", settings.ConfigDir, "/tmp/test-config")
	}

	// Test with text mode explicitly disabled
	viper.Set("text", false)
	settings2 := GetSettings()
	if settings2.TextMode {
		t.Error("expected TextMode=false after explicit set")
	}

	// Cleanup
	viper.Reset()
}
