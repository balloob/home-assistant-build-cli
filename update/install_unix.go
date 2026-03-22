//go:build !windows

package update

import (
	"fmt"
	"os"
)

// InstallUpdate replaces the current binary with the new one.
func InstallUpdate(newBinaryPath string) (*InstallResult, error) {
	execPath, err := resolveExecutablePath()
	if err != nil {
		return nil, err
	}

	backupPath := execPath + ".old"
	_ = os.Remove(backupPath)

	if err := os.Rename(execPath, backupPath); err != nil {
		return nil, fmt.Errorf("failed to prepare current binary backup: %w", err)
	}

	if err := replaceWithMoveOrCopy(newBinaryPath, execPath, 0755); err != nil {
		if restoreErr := os.Rename(backupPath, execPath); restoreErr != nil {
			return nil, fmt.Errorf("failed to install update (%v) and failed to restore previous binary (%v)", err, restoreErr)
		}
		return nil, fmt.Errorf("failed to install new binary: %w", err)
	}

	_ = os.Remove(backupPath)
	return &InstallResult{Deferred: false}, nil
}

// MaybeRunWindowsUpdateHelper is a no-op on non-Windows platforms.
func MaybeRunWindowsUpdateHelper(args []string) (bool, int) {
	return false, 0
}
