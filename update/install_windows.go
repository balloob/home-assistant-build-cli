//go:build windows

package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/sys/windows"
)

const helperWaitTimeout = 2 * time.Minute

// InstallUpdate starts a helper process that applies the update after this
// process exits, then returns with Deferred=true.
func InstallUpdate(newBinaryPath string) (*InstallResult, error) {
	execPath, err := resolveExecutablePath()
	if err != nil {
		return nil, err
	}

	helperPath := filepath.Join(os.TempDir(), fmt.Sprintf("hab-update-helper-%d.exe", time.Now().UnixNano()))
	if err := copyFile(execPath, helperPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create update helper: %w", err)
	}

	backupPath := execPath + ".old"
	helperCmd := exec.Command(
		helperPath,
		internalApplyUpdateArg,
		strconv.Itoa(os.Getpid()),
		execPath,
		newBinaryPath,
		backupPath,
	)
	helperCmd.Stdout = os.Stdout
	helperCmd.Stderr = os.Stderr

	if err := helperCmd.Start(); err != nil {
		_ = os.Remove(helperPath)
		return nil, fmt.Errorf("failed to start update helper: %w", err)
	}

	return &InstallResult{Deferred: true}, nil
}

// MaybeRunWindowsUpdateHelper executes the internal update helper flow.
func MaybeRunWindowsUpdateHelper(args []string) (bool, int) {
	if len(args) < 2 || args[1] != internalApplyUpdateArg {
		return false, 0
	}

	if len(args) != 6 {
		fmt.Fprintln(os.Stderr, "Update helper failed: invalid arguments")
		return true, 1
	}

	parentPID, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update helper failed: invalid parent pid: %v\n", err)
		return true, 1
	}

	targetPath := args[3]
	sourcePath := args[4]
	backupPath := args[5]

	if err := waitForProcessExit(uint32(parentPID), helperWaitTimeout); err != nil {
		fmt.Fprintf(os.Stderr, "Update helper failed: %v\n", err)
		return true, 1
	}

	if err := applyWindowsUpdate(targetPath, sourcePath, backupPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install update: %v\n", err)
		return true, 1
	}

	fmt.Fprintln(os.Stderr, "Successfully updated hab.")
	return true, 0
}

func applyWindowsUpdate(targetPath, sourcePath, backupPath string) error {
	_ = os.Remove(backupPath)

	if err := os.Rename(targetPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := replaceWithMoveOrCopy(sourcePath, targetPath, 0755); err != nil {
		if restoreErr := os.Rename(backupPath, targetPath); restoreErr != nil {
			return fmt.Errorf("failed to install update (%v) and failed to restore previous binary (%v)", err, restoreErr)
		}
		return fmt.Errorf("failed to place new binary: %w", err)
	}

	_ = os.Remove(backupPath)
	return nil
}

func waitForProcessExit(pid uint32, timeout time.Duration) error {
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, pid)
	if err != nil {
		if err == windows.ERROR_INVALID_PARAMETER {
			return nil
		}
		return fmt.Errorf("failed to open parent process: %w", err)
	}
	defer windows.CloseHandle(handle)

	status, err := windows.WaitForSingleObject(handle, uint32(timeout/time.Millisecond))
	if err != nil {
		return fmt.Errorf("failed while waiting for parent process: %w", err)
	}

	if status == windows.WAIT_OBJECT_0 {
		return nil
	}
	if status == uint32(windows.WAIT_TIMEOUT) {
		return fmt.Errorf("timeout waiting for parent process to exit")
	}

	return fmt.Errorf("unexpected wait status: %d", status)
}
