package cmd

import "testing"

func TestNormalizeESPHomeWizardPlatform(t *testing.T) {
	platform, err := normalizeESPHomeWizardPlatform("esp32")
	if err != nil {
		t.Fatalf("normalizeESPHomeWizardPlatform: %v", err)
	}
	if platform != "ESP32" {
		t.Fatalf("platform = %q, want ESP32", platform)
	}
}

func TestApplyESPHomePatch(t *testing.T) {
	original := "wifi:\n  ssid: old\n"
	updated, err := applyESPHomePatch(original, map[string]any{"logger": map[string]any{}}, []string{"wifi.ssid=new"})
	if err != nil {
		t.Fatalf("applyESPHomePatch: %v", err)
	}
	if updated == original {
		t.Fatal("expected updated YAML to differ from original")
	}
	if _, err := parseESPHomeConfig(updated); err != nil {
		t.Fatalf("parseESPHomeConfig(updated): %v", err)
	}
}

func TestExtractESPHomeCreateDetails(t *testing.T) {
	content := "api:\n  encryption:\n    key: abc\nota:\n  - platform: esphome\n    password: def\n"
	details, err := extractESPHomeCreateDetails(content)
	if err != nil {
		t.Fatalf("extractESPHomeCreateDetails: %v", err)
	}
	if details["api_encryption_key"] != "abc" {
		t.Fatalf("api_encryption_key = %#v, want abc", details["api_encryption_key"])
	}
	if details["ota_password"] != "def" {
		t.Fatalf("ota_password = %#v, want def", details["ota_password"])
	}
}
