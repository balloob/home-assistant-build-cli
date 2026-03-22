package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const internalApplyUpdateArg = "__hab_internal_apply_update"

// InstallResult describes how an update installation completed.
type InstallResult struct {
	Deferred bool
}

func resolveExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	resolvedPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable path: %w", err)
	}

	return resolvedPath, nil
}

func replaceWithMoveOrCopy(src, dst string, perm os.FileMode) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	if err := copyFile(src, dst, perm); err != nil {
		return err
	}

	_ = os.Remove(src)
	return nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return err
	}

	return dest.Sync()
}
