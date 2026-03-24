package esphometasmota

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Template represents a Tasmota template export payload.
type Template struct {
	Name string `json:"NAME"`
	GPIO []int  `json:"GPIO"`
	FLAG int    `json:"FLAG"`
	BASE int    `json:"BASE"`
	CMND string `json:"CMND,omitempty"`
}

// PinMapping describes a single template GPIO entry.
type PinMapping struct {
	Index     int    `json:"index"`
	Pin       string `json:"pin"`
	Code      int    `json:"code"`
	Component string `json:"component"`
	Supported bool   `json:"supported"`
	Notes     string `json:"notes,omitempty"`
}

// Analysis summarizes a parsed Tasmota template.
type Analysis struct {
	Name             string       `json:"name"`
	Base             int          `json:"base"`
	Flag             int          `json:"flag"`
	Command          string       `json:"command,omitempty"`
	SupportedCount   int          `json:"supported_count"`
	UnsupportedCount int          `json:"unsupported_count"`
	Mappings         []PinMapping `json:"mappings"`
	Warnings         []string     `json:"warnings,omitempty"`
}

type decodedComponent struct {
	kind      string
	index     int
	inverted  bool
	pullup    bool
	component string
	supported bool
	notes     string
}

var tasmotaGPIOOrder = []string{
	"GPIO0",
	"GPIO1",
	"GPIO2",
	"GPIO3",
	"GPIO4",
	"GPIO5",
	"GPIO9",
	"GPIO10",
	"GPIO12",
	"GPIO13",
	"GPIO14",
	"GPIO15",
	"GPIO16",
	"A0",
}

// ParseTemplate parses a Tasmota template JSON string.
func ParseTemplate(raw string) (*Template, error) {
	parsed := &Template{}
	if err := json.Unmarshal([]byte(raw), parsed); err != nil {
		return nil, fmt.Errorf("parse tasmota template: %w", err)
	}
	if len(parsed.GPIO) == 0 {
		return nil, fmt.Errorf("template GPIO list is empty")
	}
	return parsed, nil
}

// Analyze returns a detailed conversion analysis for a template.
func Analyze(template *Template) Analysis {
	analysis := Analysis{
		Name:     template.Name,
		Base:     template.BASE,
		Flag:     template.FLAG,
		Command:  template.CMND,
		Mappings: make([]PinMapping, 0),
		Warnings: make([]string, 0),
	}

	if template.BASE != 0 {
		analysis.Warnings = append(analysis.Warnings, "BASE metadata is advisory and may not map one-to-one in ESPHome.")
	}
	if strings.TrimSpace(template.CMND) != "" {
		analysis.Warnings = append(analysis.Warnings, "CMND automation strings are not auto-converted; review manually.")
	}

	for idx, code := range template.GPIO {
		if code == 0 || code == 255 {
			continue
		}

		pin := gpioNameForIndex(idx)
		decoded := decodeCode(code)
		mapping := PinMapping{
			Index:     idx,
			Pin:       pin,
			Code:      code,
			Component: decoded.component,
			Supported: decoded.supported,
			Notes:     decoded.notes,
		}
		analysis.Mappings = append(analysis.Mappings, mapping)

		if decoded.supported {
			analysis.SupportedCount++
		} else {
			analysis.UnsupportedCount++
		}
	}

	if analysis.SupportedCount == 0 {
		analysis.Warnings = append(analysis.Warnings, "No supported GPIO mappings detected; conversion will produce a minimal scaffold only.")
	}

	return analysis
}

// BuildOverlay converts supported template mappings into an ESPHome overlay map.
func BuildOverlay(template *Template) (map[string]any, Analysis) {
	analysis := Analyze(template)
	overlay := map[string]any{}

	switches := make([]any, 0)
	binarySensors := make([]any, 0)
	outputs := make([]any, 0)
	warnings := analysis.Warnings

	var statusLEDPin any
	i2c := map[string]any{}
	energyPins := map[string]any{}

	for _, mapping := range analysis.Mappings {
		decoded := decodeCode(mapping.Code)
		if !decoded.supported {
			continue
		}

		if mapping.Pin == "A0" && decoded.kind != "analog" {
			warnings = append(warnings, fmt.Sprintf("Ignoring non-analog mapping %s on A0", mapping.Component))
			continue
		}

		switch decoded.kind {
		case "relay":
			switches = append(switches, map[string]any{
				"platform":     "gpio",
				"name":         fmt.Sprintf("${friendly_name} Relay %d", decoded.index),
				"id":           fmt.Sprintf("relay_%d", decoded.index),
				"restore_mode": "RESTORE_DEFAULT_OFF",
				"pin":          pinValue(mapping.Pin, decoded.inverted),
			})
		case "button", "switch":
			pinConfig := map[string]any{"number": mapping.Pin}
			if decoded.pullup {
				pinConfig["mode"] = map[string]any{"input": true, "pullup": true}
			}
			if decoded.inverted {
				pinConfig["inverted"] = true
			}

			kindLabel := "Switch"
			if decoded.kind == "button" {
				kindLabel = "Button"
			}
			name := fmt.Sprintf("${friendly_name} %s %d", kindLabel, decoded.index)
			binarySensors = append(binarySensors, map[string]any{
				"platform": "gpio",
				"name":     name,
				"pin":      pinConfig,
			})
		case "pwm":
			output := map[string]any{
				"platform": "ledc",
				"id":       fmt.Sprintf("pwm_%d", decoded.index),
				"pin":      mapping.Pin,
			}
			if decoded.inverted {
				output["inverted"] = true
			}
			outputs = append(outputs, output)
		case "led":
			if statusLEDPin == nil {
				statusLEDPin = pinValue(mapping.Pin, decoded.inverted)
			}
		case "i2c_scl":
			i2c["scl"] = mapping.Pin
		case "i2c_sda":
			i2c["sda"] = mapping.Pin
		case "energy_sel":
			energyPins["sel_pin"] = pinValue(mapping.Pin, decoded.inverted)
		case "energy_cf1":
			energyPins["cf1_pin"] = mapping.Pin
		case "energy_cf":
			energyPins["cf_pin"] = mapping.Pin
		}
	}

	if len(switches) > 0 {
		overlay["switch"] = switches
	}
	if len(binarySensors) > 0 {
		overlay["binary_sensor"] = binarySensors
	}
	if len(outputs) > 0 {
		overlay["output"] = outputs
		switch len(outputs) {
		case 1:
			overlay["light"] = []any{
				map[string]any{
					"platform":     "monochromatic",
					"name":         "${friendly_name} Light",
					"output":       "pwm_1",
					"restore_mode": "RESTORE_DEFAULT_OFF",
				},
			}
		case 3:
			overlay["light"] = []any{
				map[string]any{
					"platform":     "rgb",
					"name":         "${friendly_name} Light",
					"red":          "pwm_1",
					"green":        "pwm_2",
					"blue":         "pwm_3",
					"restore_mode": "RESTORE_DEFAULT_OFF",
				},
			}
		default:
			warnings = append(warnings, "Detected PWM outputs that do not map cleanly to a default ESPHome light; review output IDs manually.")
		}
	}

	if statusLEDPin != nil {
		overlay["status_led"] = map[string]any{"pin": statusLEDPin}
	}

	if len(i2c) > 0 {
		if i2c["sda"] != nil && i2c["scl"] != nil {
			i2c["scan"] = true
			overlay["i2c"] = i2c
		} else {
			warnings = append(warnings, "Detected one I2C pin without its pair; set both SDA and SCL manually.")
		}
	}

	if energyPins["cf_pin"] != nil && energyPins["cf1_pin"] != nil && energyPins["sel_pin"] != nil {
		sensors := sensorSlice(overlay["sensor"])
		sensors = append(sensors, map[string]any{
			"platform": "hlw8012",
			"sel_pin":  energyPins["sel_pin"],
			"cf_pin":   energyPins["cf_pin"],
			"cf1_pin":  energyPins["cf1_pin"],
			"current": map[string]any{
				"name": "${friendly_name} Current",
			},
			"voltage": map[string]any{
				"name": "${friendly_name} Voltage",
			},
			"power": map[string]any{
				"name": "${friendly_name} Power",
			},
			"update_interval": "10s",
		})
		overlay["sensor"] = sensors
	} else if len(energyPins) > 0 {
		warnings = append(warnings, "Detected partial HLW8012/BL0937 energy pins; complete cf/cf1/sel mapping manually.")
	}

	analysis.Warnings = warnings
	return overlay, analysis
}

func decodeCode(code int) decodedComponent {
	switch {
	case code >= 224 && code <= 231:
		idx := code - 223
		return decodedComponent{kind: "relay", index: idx, component: fmt.Sprintf("Relay%d", idx), supported: true}
	case code >= 256 && code <= 263:
		idx := code - 255
		return decodedComponent{kind: "relay", index: idx, inverted: true, component: fmt.Sprintf("Relay_i%d", idx), supported: true}
	case code >= 32 && code <= 39:
		idx := code - 31
		return decodedComponent{kind: "button", index: idx, pullup: true, component: fmt.Sprintf("Button%d", idx), supported: true}
	case code >= 64 && code <= 71:
		idx := code - 63
		return decodedComponent{kind: "button", index: idx, component: fmt.Sprintf("Button_n%d", idx), supported: true}
	case code >= 96 && code <= 103:
		idx := code - 95
		return decodedComponent{kind: "button", index: idx, pullup: true, inverted: true, component: fmt.Sprintf("Button_i%d", idx), supported: true}
	case code >= 128 && code <= 135:
		idx := code - 127
		return decodedComponent{kind: "button", index: idx, inverted: true, component: fmt.Sprintf("Button_in%d", idx), supported: true}
	case code >= 160 && code <= 167:
		idx := code - 159
		return decodedComponent{kind: "switch", index: idx, pullup: true, component: fmt.Sprintf("Switch%d", idx), supported: true}
	case code >= 192 && code <= 199:
		idx := code - 191
		return decodedComponent{kind: "switch", index: idx, component: fmt.Sprintf("Switch_n%d", idx), supported: true}
	case code >= 416 && code <= 420:
		idx := code - 415
		return decodedComponent{kind: "pwm", index: idx, component: fmt.Sprintf("PWM%d", idx), supported: true}
	case code >= 448 && code <= 452:
		idx := code - 447
		return decodedComponent{kind: "pwm", index: idx, inverted: true, component: fmt.Sprintf("PWM_i%d", idx), supported: true}
	case code >= 288 && code <= 291:
		idx := code - 287
		return decodedComponent{kind: "led", index: idx, component: fmt.Sprintf("Led%d", idx), supported: true}
	case code >= 320 && code <= 323:
		idx := code - 319
		return decodedComponent{kind: "led", index: idx, inverted: true, component: fmt.Sprintf("Led_i%d", idx), supported: true}
	case code == 608:
		return decodedComponent{kind: "i2c_scl", component: "I2C SCL1", supported: true}
	case code == 640:
		return decodedComponent{kind: "i2c_sda", component: "I2C SDA1", supported: true}
	case code == 2592:
		return decodedComponent{kind: "energy_sel", component: "HLWBL SEL", supported: true}
	case code == 2624:
		return decodedComponent{kind: "energy_sel", inverted: true, component: "HLWBL SEL_i", supported: true}
	case code == 2656:
		return decodedComponent{kind: "energy_cf1", component: "HLWBL CF1", supported: true}
	case code == 2688:
		return decodedComponent{kind: "energy_cf", component: "HLW8012 CF", supported: true}
	case code == 2720:
		return decodedComponent{kind: "energy_cf", component: "BL0937 CF", supported: true}
	default:
		return decodedComponent{component: fmt.Sprintf("GPIO code %d", code), supported: false, notes: "unmapped GPIO function"}
	}
}

func gpioNameForIndex(index int) string {
	if index >= 0 && index < len(tasmotaGPIOOrder) {
		return tasmotaGPIOOrder[index]
	}
	return fmt.Sprintf("GPIO_INDEX_%d", index)
}

func pinValue(pin string, inverted bool) any {
	if !inverted {
		return pin
	}
	return map[string]any{
		"number":   pin,
		"inverted": true,
	}
}

func sensorSlice(value any) []any {
	if value == nil {
		return []any{}
	}
	if sensors, ok := value.([]any); ok {
		return sensors
	}
	return []any{}
}
