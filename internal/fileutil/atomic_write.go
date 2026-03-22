package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// WriteFileAtomic writes data to path via a temp file and atomic replace.
// The temp file is created in the same directory to keep replace operations
// on the same filesystem.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".hab-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err = tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err = tmpFile.Chmod(perm); err != nil {
			_ = tmpFile.Close()
			return fmt.Errorf("set temp file mode: %w", err)
		}
	}

	if err = tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}

	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err = replaceFile(tmpPath, path); err != nil {
		return fmt.Errorf("replace destination file: %w", err)
	}

	cleanup = false
	return nil
}
