package esphomeconfig

import (
	"strings"
	"testing"
)

func TestNormalizeWizardPlatform(t *testing.T) {
	platform, err := NormalizeWizardPlatform("esp32")
	if err != nil {
		t.Fatalf("NormalizeWizardPlatform: %v", err)
	}
	if platform != "ESP32" {
		t.Fatalf("platform = %q, want ESP32", platform)
	}
}

func TestApplyPatch(t *testing.T) {
	original := "wifi:\n  ssid: old\n"
	updated, err := ApplyPatch(original, map[string]any{"logger": map[string]any{}}, []string{"wifi.ssid=new"})
	if err != nil {
		t.Fatalf("ApplyPatch: %v", err)
	}
	if updated == original {
		t.Fatal("expected patched YAML to differ from original")
	}
	if _, err := ParseConfig(updated); err != nil {
		t.Fatalf("ParseConfig(updated): %v", err)
	}
}

func TestExtractCreateDetails(t *testing.T) {
	content := "api:\n  encryption:\n    key: abc\nota:\n  - platform: esphome\n    password: def\n"
	details := ExtractCreateDetails(content)
	if details["api_encryption_key"] != "abc" {
		t.Fatalf("api_encryption_key = %#v, want abc", details["api_encryption_key"])
	}
	if details["ota_password"] != "def" {
		t.Fatalf("ota_password = %#v, want def", details["ota_password"])
	}
}

func TestApplyPatchPreservesSecretTags(t *testing.T) {
	original := "api:\n  encryption:\n    key: !secret encryption_key\nwifi:\n  ssid: !secret wifi_ssid\n"

	updated, err := ApplyPatch(original, nil, []string{"esphome.name=test-node"})
	if err != nil {
		t.Fatalf("ApplyPatch: %v", err)
	}

	if !strings.Contains(updated, "!secret encryption_key") {
		t.Fatalf("patched YAML lost !secret tag for encryption key:\n%s", updated)
	}
	if !strings.Contains(updated, "!secret wifi_ssid") {
		t.Fatalf("patched YAML lost !secret tag for wifi ssid:\n%s", updated)
	}
}

func TestApplyPatchPreservesLambdaTag(t *testing.T) {
	original := "sensor:\n  - platform: template\n    lambda: !lambda |-\n      return 42;\n"

	updated, err := ApplyPatch(original, nil, []string{"esphome.name=test-node"})
	if err != nil {
		t.Fatalf("ApplyPatch: %v", err)
	}

	if !strings.Contains(updated, "!lambda") {
		t.Fatalf("patched YAML lost !lambda tag:\n%s", updated)
	}
}
