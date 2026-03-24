package esphomepreset

import (
	"fmt"
	"strconv"
	"strings"
)

// Preset describes a starter component profile for new ESPHome devices.
type Preset struct {
	ID          string            `json:"id"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Terms       map[string]string `json:"terms,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

// Options holds optional pin values used by presets.
type Options struct {
	RelayPin  string
	ButtonPin string
	PWMPin    string
	I2CSDA    string
	I2CSCL    string
}

var presetList = []Preset{
	{
		ID:      "relay",
		Summary: "Single GPIO relay switch",
		Description: "Adds one `switch` output for a relay board or in-wall relay. " +
			"Use this for smart relays and simple on/off loads.",
		Terms: map[string]string{
			"GPIO": "General-purpose input/output pin on the MCU used to read or drive hardware.",
		},
		Parameters: map[string]string{
			"relay-pin": "Relay output pin (default GPIO16)",
		},
	},
	{
		ID:      "plug",
		Summary: "Relay plug baseline",
		Description: "Adds a relay switch plus Wi-Fi signal and status sensors for a " +
			"typical smart plug baseline configuration.",
		Terms: map[string]string{
			"GPIO": "General-purpose input/output pin on the MCU used to read or drive hardware.",
		},
		Parameters: map[string]string{
			"relay-pin": "Relay output pin (default GPIO16)",
		},
	},
	{
		ID:      "light",
		Summary: "PWM dimmable light",
		Description: "Adds a monochromatic light output using PWM via `ledc` (ESP32). " +
			"Use this for dimmers and single-channel LED drivers.",
		Terms: map[string]string{
			"PWM":  "Pulse-width modulation varies duty cycle to control brightness or power level.",
			"GPIO": "General-purpose input/output pin on the MCU used to read or drive hardware.",
		},
		Parameters: map[string]string{
			"pwm-pin": "PWM output pin (default GPIO4)",
		},
	},
	{
		ID:      "button",
		Summary: "GPIO push-button input",
		Description: "Adds one debounced binary_sensor button input using internal pull-up " +
			"resistor settings.",
		Terms: map[string]string{
			"GPIO": "General-purpose input/output pin on the MCU used to read or drive hardware.",
		},
		Parameters: map[string]string{
			"button-pin": "Button input pin (default GPIO0)",
		},
	},
	{
		ID:      "sensor",
		Summary: "I2C environment sensor",
		Description: "Adds I2C bus wiring and a BME280 temperature/pressure/humidity sensor " +
			"as a beginner-friendly environmental sensor baseline.",
		Terms: map[string]string{
			"I2C":  "Two-wire peripheral bus using SDA (data) and SCL (clock) pins.",
			"GPIO": "General-purpose input/output pin on the MCU used to read or drive hardware.",
		},
		Parameters: map[string]string{
			"i2c-sda": "I2C SDA pin (default GPIO21)",
			"i2c-scl": "I2C SCL pin (default GPIO22)",
		},
	},
	{
		ID:      "display-base",
		Summary: "I2C OLED display baseline",
		Description: "Adds an SSD1306 I2C display with a simple lambda render block and font. " +
			"Use as a starter for display-based projects.",
		Terms: map[string]string{
			"I2C":    "Two-wire peripheral bus using SDA (data) and SCL (clock) pins.",
			"lambda": "Inline C++ expression block used by ESPHome for custom rendering or logic.",
		},
		Parameters: map[string]string{
			"i2c-sda": "I2C SDA pin (default GPIO21)",
			"i2c-scl": "I2C SCL pin (default GPIO22)",
		},
	},
}

// List returns all known presets.
func List() []Preset {
	out := make([]Preset, len(presetList))
	copy(out, presetList)
	return out
}

// Get returns a preset by ID.
func Get(id string) (Preset, bool) {
	needle := strings.ToLower(strings.TrimSpace(id))
	for _, preset := range presetList {
		if preset.ID == needle {
			return preset, true
		}
	}
	return Preset{}, false
}

// BuildOverlay returns the YAML overlay map for a preset and options.
func BuildOverlay(id string, options Options) (map[string]any, Preset, error) {
	preset, ok := Get(id)
	if !ok {
		return nil, Preset{}, fmt.Errorf("unknown preset %q", id)
	}

	switch preset.ID {
	case "relay":
		return relayOverlay(options), preset, nil
	case "plug":
		return plugOverlay(options), preset, nil
	case "light":
		return lightOverlay(options), preset, nil
	case "button":
		return buttonOverlay(options), preset, nil
	case "sensor":
		return sensorOverlay(options), preset, nil
	case "display-base":
		return displayOverlay(options), preset, nil
	default:
		return nil, Preset{}, fmt.Errorf("preset %q is not implemented", id)
	}
}

func relayOverlay(options Options) map[string]any {
	relayPin := normalizePin(options.RelayPin, "GPIO16")
	return map[string]any{
		"switch": []any{
			map[string]any{
				"platform":     "gpio",
				"name":         "${friendly_name} Relay",
				"id":           "relay_1",
				"restore_mode": "RESTORE_DEFAULT_OFF",
				"pin":          relayPin,
			},
		},
	}
}

func plugOverlay(options Options) map[string]any {
	relayPin := normalizePin(options.RelayPin, "GPIO16")
	return map[string]any{
		"switch": []any{
			map[string]any{
				"platform":     "gpio",
				"name":         "${friendly_name} Relay",
				"id":           "relay_1",
				"restore_mode": "RESTORE_DEFAULT_OFF",
				"pin":          relayPin,
			},
		},
		"binary_sensor": []any{
			map[string]any{
				"platform": "status",
				"name":     "${friendly_name} Status",
			},
		},
		"sensor": []any{
			map[string]any{
				"platform":        "wifi_signal",
				"name":            "${friendly_name} WiFi Signal",
				"update_interval": "60s",
			},
		},
	}
}

func lightOverlay(options Options) map[string]any {
	pwmPin := normalizePin(options.PWMPin, "GPIO4")
	return map[string]any{
		"output": []any{
			map[string]any{
				"platform": "ledc",
				"pin":      pwmPin,
				"id":       "light_output",
			},
		},
		"light": []any{
			map[string]any{
				"platform":     "monochromatic",
				"name":         "${friendly_name} Light",
				"output":       "light_output",
				"restore_mode": "RESTORE_DEFAULT_OFF",
			},
		},
	}
}

func buttonOverlay(options Options) map[string]any {
	buttonPin := normalizePin(options.ButtonPin, "GPIO0")
	return map[string]any{
		"binary_sensor": []any{
			map[string]any{
				"platform": "gpio",
				"name":     "${friendly_name} Button",
				"pin": map[string]any{
					"number":   buttonPin,
					"inverted": true,
					"mode": map[string]any{
						"input":  true,
						"pullup": true,
					},
				},
				"filters": []any{
					map[string]any{"delayed_off": "10ms"},
				},
			},
		},
	}
}

func sensorOverlay(options Options) map[string]any {
	sda := normalizePin(options.I2CSDA, "GPIO21")
	scl := normalizePin(options.I2CSCL, "GPIO22")
	return map[string]any{
		"i2c": map[string]any{
			"sda":  sda,
			"scl":  scl,
			"scan": true,
		},
		"sensor": []any{
			map[string]any{
				"platform": "bme280_i2c",
				"temperature": map[string]any{
					"name": "${friendly_name} Temperature",
				},
				"pressure": map[string]any{
					"name": "${friendly_name} Pressure",
				},
				"humidity": map[string]any{
					"name": "${friendly_name} Humidity",
				},
				"address":         "0x76",
				"update_interval": "60s",
			},
		},
	}
}

func displayOverlay(options Options) map[string]any {
	sda := normalizePin(options.I2CSDA, "GPIO21")
	scl := normalizePin(options.I2CSCL, "GPIO22")
	return map[string]any{
		"i2c": map[string]any{
			"sda":  sda,
			"scl":  scl,
			"scan": true,
		},
		"font": []any{
			map[string]any{
				"file": "gfonts://Roboto",
				"id":   "font_small",
				"size": 12,
			},
		},
		"display": []any{
			map[string]any{
				"platform": "ssd1306_i2c",
				"model":    "SSD1306 128x64",
				"address":  "0x3C",
				"lambda":   "it.printf(0, 0, id(font_small), \"${friendly_name}\");",
			},
		},
	}
}

func normalizePin(value, defaultPin string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultPin
	}

	upper := strings.ToUpper(value)
	if strings.HasPrefix(upper, "GPIO") {
		return "GPIO" + strings.TrimPrefix(upper, "GPIO")
	}

	if _, err := strconv.Atoi(value); err == nil {
		return "GPIO" + value
	}

	return value
}
