package esphomerecovery

import "testing"

func TestNormalizeChip(t *testing.T) {
	chip, err := normalizeChip("ESP32")
	if err != nil {
		t.Fatalf("normalizeChip: %v", err)
	}
	if chip != "esp32" {
		t.Fatalf("chip = %q, want esp32", chip)
	}
}

func TestNormalizeChipInvalid(t *testing.T) {
	if _, err := normalizeChip("avr"); err == nil {
		t.Fatal("expected unsupported chip error")
	}
}

func TestResolveToolExplicit(t *testing.T) {
	tool, args, err := resolveTool("custom-esptool")
	if err != nil {
		t.Fatalf("resolveTool: %v", err)
	}
	if tool != "custom-esptool" {
		t.Fatalf("tool = %q, want custom-esptool", tool)
	}
	if len(args) != 0 {
		t.Fatalf("args = %#v, want empty", args)
	}
}
