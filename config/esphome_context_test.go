package config

import (
	"os"
	"testing"
)

func TestSaveAndLoadESPHomeContext(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := &ESPHomeContext{Configuration: "living-room.yaml"}

	if err := SaveESPHomeContext(tmpDir, ctx); err != nil {
		t.Fatalf("SaveESPHomeContext: %v", err)
	}

	loaded, err := LoadESPHomeContext(tmpDir)
	if err != nil {
		t.Fatalf("LoadESPHomeContext: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadESPHomeContext returned nil")
	}
	if loaded.Configuration != ctx.Configuration {
		t.Fatalf("Configuration = %q, want %q", loaded.Configuration, ctx.Configuration)
	}
}

func TestLoadESPHomeContextNoFile(t *testing.T) {
	loaded, err := LoadESPHomeContext(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded != nil {
		t.Fatalf("LoadESPHomeContext returned %+v, want nil", loaded)
	}
}

func TestClearESPHomeContext(t *testing.T) {
	tmpDir := t.TempDir()
	if err := SaveESPHomeContext(tmpDir, &ESPHomeContext{Configuration: "kitchen.yaml"}); err != nil {
		t.Fatalf("SaveESPHomeContext: %v", err)
	}

	if err := ClearESPHomeContext(tmpDir); err != nil {
		t.Fatalf("ClearESPHomeContext: %v", err)
	}

	if _, err := os.Stat(GetESPHomeContextPath(tmpDir)); !os.IsNotExist(err) {
		t.Fatalf("context file still exists: %v", err)
	}
}

func TestSaveESPHomeContextRequiresConfiguration(t *testing.T) {
	if err := SaveESPHomeContext(t.TempDir(), &ESPHomeContext{}); err == nil {
		t.Fatal("expected error for empty configuration")
	}
}
