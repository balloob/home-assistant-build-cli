package esphomerecovery

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Options configures esptool-based operations.
type Options struct {
	Port string
	Chip string
	Tool string
}

// CommandResult captures execution details for diagnostics.
type CommandResult struct {
	Tool   string   `json:"tool"`
	Args   []string `json:"args"`
	Stdout string   `json:"stdout,omitempty"`
	Stderr string   `json:"stderr,omitempty"`
}

var supportedChips = map[string]struct{}{
	"auto":    {},
	"esp8266": {},
	"esp32":   {},
	"esp32s2": {},
	"esp32s3": {},
	"esp32c3": {},
	"esp32c6": {},
	"esp32h2": {},
}

// Probe runs esptool flash_id against the target serial port.
func Probe(ctx context.Context, options Options) (*CommandResult, error) {
	tool, prefixArgs, err := resolveTool(options.Tool)
	if err != nil {
		return nil, err
	}

	chip, err := normalizeChip(options.Chip)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(options.Port) == "" {
		return nil, fmt.Errorf("serial port is required")
	}

	args := append(prefixArgs,
		"--port", options.Port,
		"--chip", chip,
		"flash_id",
	)

	return run(ctx, tool, args)
}

// EraseFlash runs esptool erase_flash against the target serial port.
func EraseFlash(ctx context.Context, options Options) (*CommandResult, error) {
	tool, prefixArgs, err := resolveTool(options.Tool)
	if err != nil {
		return nil, err
	}

	chip, err := normalizeChip(options.Chip)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(options.Port) == "" {
		return nil, fmt.Errorf("serial port is required")
	}

	args := append(prefixArgs,
		"--port", options.Port,
		"--chip", chip,
		"erase_flash",
	)

	return run(ctx, tool, args)
}

func resolveTool(explicitTool string) (string, []string, error) {
	if tool := strings.TrimSpace(explicitTool); tool != "" {
		return tool, nil, nil
	}

	if tool := strings.TrimSpace(os.Getenv("HAB_ESPTOOL_BIN")); tool != "" {
		return tool, nil, nil
	}

	for _, candidate := range []string{"esptool", "esptool.py"} {
		if resolved, err := exec.LookPath(candidate); err == nil {
			return resolved, nil, nil
		}
	}

	pythonCandidates := []string{"python3", "python"}
	if runtime.GOOS == "windows" {
		pythonCandidates = append(pythonCandidates, "py")
	}
	for _, candidate := range pythonCandidates {
		if resolved, err := exec.LookPath(candidate); err == nil {
			if filepath.Base(resolved) == "py.exe" || filepath.Base(resolved) == "py" {
				return resolved, []string{"-3", "-m", "esptool"}, nil
			}
			return resolved, []string{"-m", "esptool"}, nil
		}
	}

	return "", nil, fmt.Errorf("esptool not found (set HAB_ESPTOOL_BIN or install esptool)")
}

func normalizeChip(value string) (string, error) {
	chip := strings.ToLower(strings.TrimSpace(value))
	if chip == "" {
		chip = "auto"
	}
	if _, ok := supportedChips[chip]; !ok {
		return "", fmt.Errorf("unsupported chip %q", value)
	}
	return chip, nil
}

func run(ctx context.Context, tool string, args []string) (*CommandResult, error) {
	cmd := exec.CommandContext(ctx, tool, args...)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &CommandResult{
		Tool:   tool,
		Args:   args,
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
	}
	if err != nil {
		if result.Stderr != "" {
			return result, fmt.Errorf("esptool command failed: %s", result.Stderr)
		}
		return result, fmt.Errorf("esptool command failed: %w", err)
	}

	return result, nil
}
