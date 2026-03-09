// Package config manages configuration file paths and application settings
// for the hab CLI. Configuration is stored under ~/.config/home-assistant-builder/.
package config

import (
	"os"
	"path/filepath"
)

const (
	// DefaultConfigDir is the default configuration directory name
	DefaultConfigDir = "home-assistant-builder"
	// CredentialsFile is the encrypted credentials file name
	CredentialsFile = "credentials.json"
	// ConfigFile is the configuration file name
	ConfigFile = "config.json"
)

// GetConfigDir returns the configuration directory path.
// If configDir is empty, returns the default XDG config directory.
func GetConfigDir(configDir string) string {
	if configDir != "" {
		return configDir
	}

	// Try XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, DefaultConfigDir)
	}

	// Fall back to ~/.config
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", DefaultConfigDir)
	}
	return filepath.Join(home, ".config", DefaultConfigDir)
}

// GetCredentialsPath returns the path to the credentials file.
func GetCredentialsPath(configDir string) string {
	return filepath.Join(GetConfigDir(configDir), CredentialsFile)
}

// GetConfigPath returns the path to the config file.
func GetConfigPath(configDir string) string {
	return filepath.Join(GetConfigDir(configDir), ConfigFile)
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir(configDir string) error {
	dir := GetConfigDir(configDir)
	return os.MkdirAll(dir, 0700)
}
