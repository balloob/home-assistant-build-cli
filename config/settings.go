package config

import (
	"github.com/spf13/viper"
)

// Settings holds the application configuration
type Settings struct {
	URL       string `mapstructure:"url"`
	TextMode  bool   `mapstructure:"text"`
	Verbose   bool   `mapstructure:"verbose"`
	ConfigDir string `mapstructure:"config"`
}

// InitDefaults sets the default values for configuration
func InitDefaults() {
	viper.SetDefault("text", true) // Text mode is the default for human-readable output
	viper.SetDefault("verbose", false)
}

// GetSettings returns the current settings from Viper
func GetSettings() *Settings {
	return &Settings{
		URL:       viper.GetString("url"),
		TextMode:  viper.GetBool("text"),
		Verbose:   viper.GetBool("verbose"),
		ConfigDir: viper.GetString("config"),
	}
}
