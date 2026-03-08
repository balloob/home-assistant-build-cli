package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

// templateCreateOpts groups all flag values for the template create command.
type templateCreateOpts struct {
	// Type selector
	Type string
	// Common
	StateTemplate string
	Icon          string
	// Sensor
	Unit        string
	DeviceClass string
	StateClass  string
	// Switch/Light/Fan
	TurnOn  string
	TurnOff string
	// Button
	Press string
	// Cover
	Open    string
	Close   string
	Stop    string
	Position string
	SetPos  string
	Tilt    string
	SetTilt string
	// Lock
	Lock   string
	Unlock string
	// Image
	URL string
	// Number
	Min      float64
	Max      float64
	Step     float64
	SetValue string
	// Select
	Options      []string
	SelectOption string
	// Weather
	Condition   string
	Temperature string
	Humidity    string
	// Light
	Brightness string
	Color      string
	Effect     string
	Effects    []string
	// Fan
	Percentage string
	SetPct     string
	Preset     string
	SetPreset  string
	Oscillate  string
	OscOn      string
	OscOff     string
	Direction  string
	SetDir     string
	// Vacuum
	Start        string
	Pause        string
	ReturnToBase string
	Clean        string
	Locate       string
	SetFanSpeed  string
	FanSpeed     string
	Battery      string
}

// package-level instance populated by cobra flags.
var tplOpts templateCreateOpts

// setupTemplateCreateFlags registers all flags for the template create command.
func setupTemplateCreateFlags(cmd *cobra.Command) {
	o := &tplOpts
	cmd.Flags().StringVarP(&o.Type, "type", "t", "sensor", "Template type: alarm_control_panel, binary_sensor, button, cover, fan, image, light, lock, number, select, sensor, switch, vacuum, weather")
	cmd.Flags().StringVar(&o.StateTemplate, "state", "", "State template (Jinja2)")
	cmd.Flags().StringVar(&o.Icon, "icon", "", "Icon (e.g., mdi:thermometer)")
	cmd.Flags().StringVar(&o.Unit, "unit", "", "Unit of measurement (sensor)")
	cmd.Flags().StringVar(&o.DeviceClass, "device-class", "", "Device class (sensor, binary_sensor, cover)")
	cmd.Flags().StringVar(&o.StateClass, "state-class", "", "State class: measurement, total, total_increasing (sensor)")
	cmd.Flags().StringVar(&o.TurnOn, "turn-on", "", "Turn on action (switch, light, fan)")
	cmd.Flags().StringVar(&o.TurnOff, "turn-off", "", "Turn off action (switch, light, fan)")
	cmd.Flags().StringVar(&o.Press, "press", "", "Press action (button)")
	cmd.Flags().StringVar(&o.Open, "open", "", "Open action (cover)")
	cmd.Flags().StringVar(&o.Close, "close", "", "Close action (cover)")
	cmd.Flags().StringVar(&o.Stop, "stop", "", "Stop action (cover)")
	cmd.Flags().StringVar(&o.Position, "position", "", "Position template (cover)")
	cmd.Flags().StringVar(&o.SetPos, "set-position", "", "Set position action (cover)")
	cmd.Flags().StringVar(&o.Tilt, "tilt", "", "Tilt template (cover)")
	cmd.Flags().StringVar(&o.SetTilt, "set-tilt", "", "Set tilt action (cover)")
	cmd.Flags().StringVar(&o.Lock, "lock", "", "Lock action (lock)")
	cmd.Flags().StringVar(&o.Unlock, "unlock", "", "Unlock action (lock)")
	cmd.Flags().StringVar(&o.URL, "url", "", "URL template (image)")
	cmd.Flags().Float64Var(&o.Min, "min", 0, "Minimum value (number)")
	cmd.Flags().Float64Var(&o.Max, "max", 100, "Maximum value (number)")
	cmd.Flags().Float64Var(&o.Step, "step", 1, "Step value (number)")
	cmd.Flags().StringVar(&o.SetValue, "set-value", "", "Set value action (number)")
	cmd.Flags().StringSliceVar(&o.Options, "options", nil, "Options (select, comma-separated)")
	cmd.Flags().StringVar(&o.SelectOption, "select-option", "", "Select option action (select)")
	cmd.Flags().StringVar(&o.Condition, "condition", "", "Condition template (weather)")
	cmd.Flags().StringVar(&o.Temperature, "temperature", "", "Temperature template (weather)")
	cmd.Flags().StringVar(&o.Humidity, "humidity", "", "Humidity template (weather)")
	cmd.Flags().StringVar(&o.Brightness, "brightness", "", "Brightness template (light)")
	cmd.Flags().StringVar(&o.Color, "color", "", "Color template (light)")
	cmd.Flags().StringVar(&o.Effect, "effect", "", "Effect template (light)")
	cmd.Flags().StringSliceVar(&o.Effects, "effects", nil, "Effect list (light, comma-separated)")
	cmd.Flags().StringVar(&o.Percentage, "percentage", "", "Percentage template (fan)")
	cmd.Flags().StringVar(&o.SetPct, "set-percentage", "", "Set percentage action (fan)")
	cmd.Flags().StringVar(&o.Preset, "preset", "", "Preset mode template (fan)")
	cmd.Flags().StringVar(&o.SetPreset, "set-preset", "", "Set preset mode action (fan)")
	cmd.Flags().StringVar(&o.Oscillate, "oscillate", "", "Oscillating template (fan)")
	cmd.Flags().StringVar(&o.OscOn, "oscillate-on", "", "Set oscillating on action (fan)")
	cmd.Flags().StringVar(&o.OscOff, "oscillate-off", "", "Set oscillating off action (fan)")
	cmd.Flags().StringVar(&o.Direction, "direction", "", "Direction template (fan)")
	cmd.Flags().StringVar(&o.SetDir, "set-direction", "", "Set direction action (fan)")
	cmd.Flags().StringVar(&o.Start, "start", "", "Start action (vacuum)")
	cmd.Flags().StringVar(&o.Pause, "pause", "", "Pause action (vacuum)")
	cmd.Flags().StringVar(&o.ReturnToBase, "return-to-base", "", "Return to base action (vacuum)")
	cmd.Flags().StringVar(&o.Clean, "clean-spot", "", "Clean spot action (vacuum)")
	cmd.Flags().StringVar(&o.Locate, "locate", "", "Locate action (vacuum)")
	cmd.Flags().StringVar(&o.SetFanSpeed, "set-fan-speed", "", "Set fan speed action (vacuum)")
	cmd.Flags().StringVar(&o.FanSpeed, "fan-speed", "", "Fan speed template (vacuum)")
	cmd.Flags().StringVar(&o.Battery, "battery", "", "Battery level template (vacuum)")
}

// templateFormBuilders maps each template type to a function that populates
// the type-specific fields on the config-flow form data map. Each builder
// reads from the shared tplOpts struct.
var templateFormBuilders = map[string]func(m map[string]interface{}){
	"sensor": func(m map[string]interface{}) {
		setIf(m, "state", tplOpts.StateTemplate)
		setIf(m, "unit_of_measurement", tplOpts.Unit)
		setIf(m, "device_class", tplOpts.DeviceClass)
		setIf(m, "state_class", tplOpts.StateClass)
	},
	"binary_sensor": func(m map[string]interface{}) {
		setIf(m, "state", tplOpts.StateTemplate)
		setIf(m, "device_class", tplOpts.DeviceClass)
	},
	"alarm_control_panel": func(m map[string]interface{}) {
		setIf(m, "value_template", tplOpts.StateTemplate)
	},
	"switch": func(m map[string]interface{}) {
		setIf(m, "value_template", tplOpts.StateTemplate)
		setAction(m, "turn_on", tplOpts.TurnOn)
		setAction(m, "turn_off", tplOpts.TurnOff)
	},
	"button": func(m map[string]interface{}) {
		setAction(m, "press", tplOpts.Press)
	},
	"cover": func(m map[string]interface{}) {
		setIf(m, "value_template", tplOpts.StateTemplate)
		setIf(m, "device_class", tplOpts.DeviceClass)
		setIf(m, "position_template", tplOpts.Position)
		setAction(m, "open_cover", tplOpts.Open)
		setAction(m, "close_cover", tplOpts.Close)
		setAction(m, "stop_cover", tplOpts.Stop)
		setAction(m, "set_cover_position", tplOpts.SetPos)
		setIf(m, "tilt_template", tplOpts.Tilt)
		setAction(m, "set_cover_tilt_position", tplOpts.SetTilt)
	},
	"lock": func(m map[string]interface{}) {
		setIf(m, "value_template", tplOpts.StateTemplate)
		setAction(m, "lock", tplOpts.Lock)
		setAction(m, "unlock", tplOpts.Unlock)
	},
	"light": func(m map[string]interface{}) {
		setIf(m, "value_template", tplOpts.StateTemplate)
		setAction(m, "turn_on", tplOpts.TurnOn)
		setAction(m, "turn_off", tplOpts.TurnOff)
		setIf(m, "level_template", tplOpts.Brightness)
		setIf(m, "color_template", tplOpts.Color)
		setIf(m, "effect_template", tplOpts.Effect)
		if len(tplOpts.Effects) > 0 {
			m["effect_list"] = tplOpts.Effects
		}
	},
	"fan": func(m map[string]interface{}) {
		setIf(m, "value_template", tplOpts.StateTemplate)
		setAction(m, "turn_on", tplOpts.TurnOn)
		setAction(m, "turn_off", tplOpts.TurnOff)
		setIf(m, "percentage_template", tplOpts.Percentage)
		setAction(m, "set_percentage", tplOpts.SetPct)
		setIf(m, "preset_mode_template", tplOpts.Preset)
		setAction(m, "set_preset_mode", tplOpts.SetPreset)
		setIf(m, "oscillating_template", tplOpts.Oscillate)
		setAction(m, "set_oscillating", tplOpts.OscOn)
		setIf(m, "direction_template", tplOpts.Direction)
		setAction(m, "set_direction", tplOpts.SetDir)
	},
	"vacuum": func(m map[string]interface{}) {
		setIf(m, "value_template", tplOpts.StateTemplate)
		setAction(m, "start", tplOpts.Start)
		setAction(m, "pause", tplOpts.Pause)
		setAction(m, "return_to_base", tplOpts.ReturnToBase)
		setAction(m, "clean_spot", tplOpts.Clean)
		setAction(m, "locate", tplOpts.Locate)
		setAction(m, "set_fan_speed", tplOpts.SetFanSpeed)
		setIf(m, "fan_speed_template", tplOpts.FanSpeed)
		setIf(m, "battery_level_template", tplOpts.Battery)
	},
	"image": func(m map[string]interface{}) {
		setIf(m, "url", tplOpts.URL)
	},
	"number": func(m map[string]interface{}) {
		setIf(m, "state", tplOpts.StateTemplate)
		m["min"] = tplOpts.Min
		m["max"] = tplOpts.Max
		m["step"] = tplOpts.Step
		setAction(m, "set_value", tplOpts.SetValue)
	},
	"select": func(m map[string]interface{}) {
		setIf(m, "state", tplOpts.StateTemplate)
		if len(tplOpts.Options) > 0 {
			m["options"] = buildOptionsTemplate(tplOpts.Options)
		}
		setAction(m, "select_option", tplOpts.SelectOption)
	},
	"weather": func(m map[string]interface{}) {
		setIf(m, "condition_template", tplOpts.Condition)
		setIf(m, "temperature_template", tplOpts.Temperature)
		setIf(m, "humidity_template", tplOpts.Humidity)
	},
}

// setIf sets m[key] = value when value is non-empty.
func setIf(m map[string]interface{}, key, value string) {
	if value != "" {
		m[key] = value
	}
}

// setAction sets m[key] to an action sequence when value is non-empty.
func setAction(m map[string]interface{}, key, value string) {
	if value != "" {
		m[key] = buildActionSequence(value)
	}
}

// runTemplateCreate handles the template create command.
func runTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	textMode := getTextMode()

	builder, ok := templateFormBuilders[tplOpts.Type]
	if !ok {
		keys := make([]string, 0, len(templateFormBuilders))
		for k := range templateFormBuilders {
			keys = append(keys, k)
		}
		return fmt.Errorf("invalid template type: %s. Valid types: %s", tplOpts.Type, strings.Join(keys, ", "))
	}

	rest, err := getRESTClient()
	if err != nil {
		return err
	}

	flowID, err := configFlowStart(rest, "template")
	if err != nil {
		return err
	}

	// Menu step: select template type
	menuResult, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
		"next_step_id": tplOpts.Type,
	})
	if err != nil {
		return fmt.Errorf("failed to select template type: %w", err)
	}
	if err := configFlowCheckAbort(menuResult); err != nil {
		return err
	}

	formData := map[string]interface{}{"name": name}
	setIf(formData, "icon", tplOpts.Icon)
	builder(formData)

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create template entity: %w", err)
	}

	result, err := configFlowCheckFinal(finalResult)
	if err != nil {
		return err
	}
	result["type"] = tplOpts.Type

	output.PrintSuccess(result, textMode, fmt.Sprintf("Template %s '%s' created successfully.", tplOpts.Type, name))
	return nil
}

func buildOptionsTemplate(options []string) string {
	var items []string
	for _, opt := range options {
		items = append(items, fmt.Sprintf("'%s'", opt))
	}
	return fmt.Sprintf("{{ [%s] }}", strings.Join(items, ", "))
}

func buildActionSequence(action string) []map[string]interface{} {
	return []map[string]interface{}{
		{"action": action},
	}
}
