package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/home-assistant/hab/internal/fileutil"
)

const ESPHomeContextFile = "esphome-context.json"

// ESPHomeContext stores the selected ESPHome configuration for follow-up commands.
type ESPHomeContext struct {
	Configuration string `json:"configuration"`
}

// GetESPHomeContextPath returns the path to the ESPHome context file.
func GetESPHomeContextPath(configDir string) string {
	return filepath.Join(GetConfigDir(configDir), ESPHomeContextFile)
}

// LoadESPHomeContext loads the saved ESPHome context, if any.
func LoadESPHomeContext(configDir string) (*ESPHomeContext, error) {
	data, err := os.ReadFile(GetESPHomeContextPath(configDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var ctx ESPHomeContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(ctx.Configuration) == "" {
		return nil, nil
	}
	return &ctx, nil
}

// SaveESPHomeContext persists the selected ESPHome configuration.
func SaveESPHomeContext(configDir string, ctx *ESPHomeContext) error {
	if ctx == nil || strings.TrimSpace(ctx.Configuration) == "" {
		return fmt.Errorf("configuration is required")
	}
	if err := EnsureConfigDir(configDir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}

	return fileutil.WriteFileAtomic(GetESPHomeContextPath(configDir), data, 0600)
}

// ClearESPHomeContext removes any saved ESPHome context.
func ClearESPHomeContext(configDir string) error {
	err := os.Remove(GetESPHomeContextPath(configDir))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
