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
	"time"
)

var (
	lookPath              = exec.LookPath
	commandContext        = exec.CommandContext
	pythonSupportsESPTool = probePythonESPTool
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
		resolved, args := normalizeResolvedTool(tool)
		return resolved, args, nil
	}

	if tool := strings.TrimSpace(os.Getenv("HAB_ESPTOOL_BIN")); tool != "" {
		resolved, args := normalizeResolvedTool(tool)
		return resolved, args, nil
	}

	for _, candidate := range []string{"esptool", "esptool.py"} {
		if resolved, err := lookPath(candidate); err == nil {
			tool, args := normalizeResolvedTool(resolved)
			return tool, args, nil
		}
	}

	if resolved, err := lookPath("uvx"); err == nil {
		return resolved, []string{"esptool"}, nil
	}

	pythonCandidates := []string{"python3", "python"}
	if runtime.GOOS == "windows" {
		pythonCandidates = append(pythonCandidates, "py")
	}
	for _, candidate := range pythonCandidates {
		if resolved, err := lookPath(candidate); err == nil && pythonSupportsESPTool(resolved) {
			if filepath.Base(resolved) == "py.exe" || filepath.Base(resolved) == "py" {
				return resolved, []string{"-3", "-m", "esptool"}, nil
			}
			return resolved, []string{"-m", "esptool"}, nil
		}
	}

	return "", nil, fmt.Errorf("esptool not found (set HAB_ESPTOOL_BIN, install esptool, or install uvx for `uvx esptool`)")
}

func normalizeResolvedTool(tool string) (string, []string) {
	base := strings.ToLower(filepath.Base(strings.TrimSpace(tool)))
	if base == "uvx" || base == "uvx.exe" {
		return tool, []string{"esptool"}
	}
	return tool, nil
}

func probePythonESPTool(tool string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []string{"-c", "import esptool"}
	if filepath.Base(tool) == "py.exe" || filepath.Base(tool) == "py" {
		args = []string{"-3", "-c", "import esptool"}
	}

	cmd := commandContext(ctx, tool, args...)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
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
	cmd := commandContext(ctx, tool, args...)

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
