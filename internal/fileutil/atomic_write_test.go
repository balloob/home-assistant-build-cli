package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	if err := WriteFileAtomic(path, []byte("first"), 0600); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	if err := WriteFileAtomic(path, []byte("second"), 0600); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if string(data) != "second" {
		t.Fatalf("content = %q, want %q", string(data), "second")
	}
}
