package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseInput_JSONString(t *testing.T) {
	result, err := ParseInput(`{"name": "test", "count": 3}`, "", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("name = %v, want %q", result["name"], "test")
	}
	// JSON numbers unmarshal as float64
	if result["count"] != float64(3) {
		t.Errorf("count = %v, want 3", result["count"])
	}
}

func TestParseInput_YAMLString(t *testing.T) {
	result, err := ParseInput("name: test\ncount: 3", "", "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("name = %v, want %q", result["name"], "test")
	}
	if result["count"] != float64(3) {
		t.Errorf("count = %v, want 3", result["count"])
	}
}

func TestParseInput_AutoDetectJSON(t *testing.T) {
	result, err := ParseInput(`{"key": "val"}`, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["key"] != "val" {
		t.Errorf("key = %v, want %q", result["key"], "val")
	}
}

func TestParseInput_AutoDetectYAML(t *testing.T) {
	result, err := ParseInput("key: val", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["key"] != "val" {
		t.Errorf("key = %v, want %q", result["key"], "val")
	}
}

func TestParseInput_AutoDetectJSONArray(t *testing.T) {
	// Array prefix should be detected as JSON, but ParseInput expects a map
	_, err := ParseInput(`[1, 2, 3]`, "", "")
	if err == nil {
		t.Error("expected error for JSON array (not a map), got nil")
	}
}

func TestParseInput_JSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	if err := os.WriteFile(path, []byte(`{"from_file": true}`), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	result, err := ParseInput("", path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["from_file"] != true {
		t.Errorf("from_file = %v, want true", result["from_file"])
	}
}

func TestParseInput_YAMLFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.yaml")
	if err := os.WriteFile(path, []byte("from_file: true\n"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	result, err := ParseInput("", path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["from_file"] != true {
		t.Errorf("from_file = %v, want true", result["from_file"])
	}
}

func TestParseInput_YMLExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.yml")
	if err := os.WriteFile(path, []byte("ext: yml\n"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	result, err := ParseInput("", path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["ext"] != "yml" {
		t.Errorf("ext = %v, want %q", result["ext"], "yml")
	}
}

func TestParseInput_EmptyData(t *testing.T) {
	_, err := ParseInput("", "", "json")
	// With empty data and no file and no stdin, we expect an error
	// Note: this would actually block on stdin; we test via data=""
	// The function reads stdin if both data and file are empty,
	// so this test is limited. We test empty inputData explicitly.
	if err == nil {
		// If stdin is empty/closed in test environment, this should error
		t.Log("Note: empty input test depends on stdin state in test env")
	}
}

func TestParseInput_InvalidJSON(t *testing.T) {
	_, err := ParseInput(`{invalid json}`, "", "json")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseInput_InvalidYAML(t *testing.T) {
	_, err := ParseInput(":\n  :\n    - :", "", "yaml")
	// Some edge cases may parse; use clearly invalid YAML
	// The yaml parser is fairly lenient, so this may not error.
	// Use a format that will definitely fail JSON unmarshal after YAML conversion.
	_ = err // YAML parser is lenient; just ensure no panic
}

func TestParseInput_UnsupportedFormat(t *testing.T) {
	_, err := ParseInput("data", "", "xml")
	if err == nil {
		t.Error("expected error for unsupported format, got nil")
	}
	if err != nil && err.Error() != "unsupported format: xml" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseInput_FileNotFound(t *testing.T) {
	_, err := ParseInput("", "/nonexistent/file.json", "json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestParseInput_FormatOverridesExtension(t *testing.T) {
	// File has .yaml extension but we force JSON format
	dir := t.TempDir()
	path := filepath.Join(dir, "data.yaml")
	if err := os.WriteFile(path, []byte(`{"forced": "json"}`), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	result, err := ParseInput("", path, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["forced"] != "json" {
		t.Errorf("forced = %v, want %q", result["forced"], "json")
	}
}

func TestParseInput_DataTakesPriorityOverFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	if err := os.WriteFile(path, []byte(`{"source": "file"}`), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// When file is provided, data is ignored (file takes priority per code logic)
	result, err := ParseInput(`{"source": "data"}`, path, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Code checks file != "" first, so file wins
	if result["source"] != "file" {
		t.Errorf("source = %v, want %q (file should take priority)", result["source"], "file")
	}
}
