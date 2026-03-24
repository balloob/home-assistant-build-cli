package esphomepreset

import "testing"

func TestGetPreset(t *testing.T) {
	preset, ok := Get("relay")
	if !ok {
		t.Fatal("expected relay preset")
	}
	if preset.ID != "relay" {
		t.Fatalf("preset.ID = %q, want relay", preset.ID)
	}
}

func TestBuildOverlay(t *testing.T) {
	overlay, preset, err := BuildOverlay("light", Options{PWMPin: "5"})
	if err != nil {
		t.Fatalf("BuildOverlay: %v", err)
	}
	if preset.ID != "light" {
		t.Fatalf("preset.ID = %q, want light", preset.ID)
	}

	outputs, ok := overlay["output"].([]any)
	if !ok || len(outputs) == 0 {
		t.Fatalf("overlay output = %#v", overlay["output"])
	}
	out := outputs[0].(map[string]any)
	if out["pin"] != "GPIO5" {
		t.Fatalf("pin = %#v, want GPIO5", out["pin"])
	}
}

func TestBuildOverlayUnknownPreset(t *testing.T) {
	if _, _, err := BuildOverlay("missing", Options{}); err == nil {
		t.Fatal("expected error for unknown preset")
	}
}
