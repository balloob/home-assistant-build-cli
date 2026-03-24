package esphomerecovery

import (
	"errors"
	"testing"
)

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

func TestResolveToolExplicitUVX(t *testing.T) {
	tool, args, err := resolveTool("/usr/bin/uvx")
	if err != nil {
		t.Fatalf("resolveTool: %v", err)
	}
	if tool != "/usr/bin/uvx" {
		t.Fatalf("tool = %q, want /usr/bin/uvx", tool)
	}
	if len(args) != 1 || args[0] != "esptool" {
		t.Fatalf("args = %#v, want [esptool]", args)
	}
}

func TestResolveToolFallsBackToUVX(t *testing.T) {
	originalLookPath := lookPath
	originalPythonSupportsESPTool := pythonSupportsESPTool
	t.Cleanup(func() {
		lookPath = originalLookPath
		pythonSupportsESPTool = originalPythonSupportsESPTool
	})

	lookPath = func(file string) (string, error) {
		switch file {
		case "uvx":
			return "/usr/bin/uvx", nil
		default:
			return "", errors.New("not found")
		}
	}
	pythonSupportsESPTool = func(string) bool { return false }

	tool, args, err := resolveTool("")
	if err != nil {
		t.Fatalf("resolveTool: %v", err)
	}
	if tool != "/usr/bin/uvx" {
		t.Fatalf("tool = %q, want /usr/bin/uvx", tool)
	}
	if len(args) != 1 || args[0] != "esptool" {
		t.Fatalf("args = %#v, want [esptool]", args)
	}
}

func TestResolveToolUsesPythonModuleWhenAvailable(t *testing.T) {
	originalLookPath := lookPath
	originalPythonSupportsESPTool := pythonSupportsESPTool
	t.Cleanup(func() {
		lookPath = originalLookPath
		pythonSupportsESPTool = originalPythonSupportsESPTool
	})

	lookPath = func(file string) (string, error) {
		switch file {
		case "python3":
			return "/usr/bin/python3", nil
		default:
			return "", errors.New("not found")
		}
	}
	pythonSupportsESPTool = func(tool string) bool { return tool == "/usr/bin/python3" }

	tool, args, err := resolveTool("")
	if err != nil {
		t.Fatalf("resolveTool: %v", err)
	}
	if tool != "/usr/bin/python3" {
		t.Fatalf("tool = %q, want /usr/bin/python3", tool)
	}
	if len(args) != 2 || args[0] != "-m" || args[1] != "esptool" {
		t.Fatalf("args = %#v, want [-m esptool]", args)
	}
}
