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

func TestSetDeviceNameViaSubstitution(t *testing.T) {
	original := "substitutions:\n  devicename: old-node\nesphome:\n  name: ${devicename}\nwifi:\n  use_address: ${devicename}.local\n"

	updated, err := SetDeviceName(original, "new-node")
	if err != nil {
		t.Fatalf("SetDeviceName: %v", err)
	}

	config, err := ParseConfig(updated)
	if err != nil {
		t.Fatalf("ParseConfig(updated): %v", err)
	}

	substitutions, ok := config["substitutions"].(map[string]any)
	if !ok {
		t.Fatal("expected substitutions map")
	}
	if substitutions["devicename"] != "new-node" {
		t.Fatalf("substitutions.devicename = %#v, want new-node", substitutions["devicename"])
	}

	esphome, ok := config["esphome"].(map[string]any)
	if !ok {
		t.Fatal("expected esphome map")
	}
	if esphome["name"] != "${devicename}" {
		t.Fatalf("esphome.name = %#v, want ${devicename}", esphome["name"])
	}
}

func TestSetDeviceNameLiteral(t *testing.T) {
	original := "esphome:\n  name: old-node\n"

	updated, err := SetDeviceName(original, "new-node")
	if err != nil {
		t.Fatalf("SetDeviceName: %v", err)
	}

	config, err := ParseConfig(updated)
	if err != nil {
		t.Fatalf("ParseConfig(updated): %v", err)
	}

	esphome, ok := config["esphome"].(map[string]any)
	if !ok {
		t.Fatal("expected esphome map")
	}
	if esphome["name"] != "new-node" {
		t.Fatalf("esphome.name = %#v, want new-node", esphome["name"])
	}
}

func TestSetDeviceNameDollarSubstitution(t *testing.T) {
	original := "substitutions:\n  devicename: old-node\nesphome:\n  name: $devicename\n"

	updated, err := SetDeviceName(original, "new-node")
	if err != nil {
		t.Fatalf("SetDeviceName: %v", err)
	}

	config, err := ParseConfig(updated)
	if err != nil {
		t.Fatalf("ParseConfig(updated): %v", err)
	}

	substitutions, ok := config["substitutions"].(map[string]any)
	if !ok {
		t.Fatal("expected substitutions map")
	}
	if substitutions["devicename"] != "new-node" {
		t.Fatalf("substitutions.devicename = %#v, want new-node", substitutions["devicename"])
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
