package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

// Template create flag variables — package-level because they're shared between
// setupTemplateCreateFlags and the various build functions.
var (
	helperTemplateCreateType string
	// Common options
	helperTemplateCreateStateTemplate string
	helperTemplateCreateIcon          string
	// Sensor options
	helperTemplateCreateUnit        string
	helperTemplateCreateDeviceClass string
	helperTemplateCreateStateClass  string
	// Switch/Light/Fan options
	helperTemplateCreateTurnOn  string
	helperTemplateCreateTurnOff string
	// Button options
	helperTemplateCreatePress string
	// Cover options
	helperTemplateCreateOpen     string
	helperTemplateCreateClose    string
	helperTemplateCreateStop     string
	helperTemplateCreatePosition string
	helperTemplateCreateSetPos   string
	helperTemplateCreateTilt     string
	helperTemplateCreateSetTilt  string
	// Lock options
	helperTemplateCreateLock   string
	helperTemplateCreateUnlock string
	// Image options
	helperTemplateCreateURL string
	// Number options
	helperTemplateCreateMin      float64
	helperTemplateCreateMax      float64
	helperTemplateCreateStep     float64
	helperTemplateCreateSetValue string
	// Select options
	helperTemplateCreateOptions      []string
	helperTemplateCreateSelectOption string
	// Weather options
	helperTemplateCreateCondition   string
	helperTemplateCreateTemperature string
	helperTemplateCreateHumidity    string
	// Light specific
	helperTemplateCreateBrightness string
	helperTemplateCreateColor      string
	helperTemplateCreateEffect     string
	helperTemplateCreateEffects    []string
	// Fan specific
	helperTemplateCreateSpeed      string
	helperTemplateCreateOscillate  string
	helperTemplateCreateDirection  string
	helperTemplateCreatePreset     string
	helperTemplateCreateSetSpeed   string
	helperTemplateCreateOscOn      string
	helperTemplateCreateOscOff     string
	helperTemplateCreateSetDir     string
	helperTemplateCreateSetPreset  string
	helperTemplateCreatePercentage string
	helperTemplateCreateSetPct     string
	// Vacuum options
	helperTemplateCreateStart        string
	helperTemplateCreatePause        string
	helperTemplateCreateReturnToBase string
	helperTemplateCreateClean        string
	helperTemplateCreateLocate       string
	helperTemplateCreateSetFanSpeed  string
	helperTemplateCreateFanSpeed     string
	helperTemplateCreateBattery      string
)

// setupTemplateCreateFlags registers all flags for the template create command.
func setupTemplateCreateFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&helperTemplateCreateType, "type", "t", "sensor", "Template type: alarm_control_panel, binary_sensor, button, image, number, select, sensor, switch")
	cmd.Flags().StringVar(&helperTemplateCreateStateTemplate, "state", "", "State template (Jinja2)")
	cmd.Flags().StringVar(&helperTemplateCreateIcon, "icon", "", "Icon (e.g., mdi:thermometer)")
	cmd.Flags().StringVar(&helperTemplateCreateUnit, "unit", "", "Unit of measurement (sensor)")
	cmd.Flags().StringVar(&helperTemplateCreateDeviceClass, "device-class", "", "Device class (sensor, binary_sensor, cover)")
	cmd.Flags().StringVar(&helperTemplateCreateStateClass, "state-class", "", "State class: measurement, total, total_increasing (sensor)")
	cmd.Flags().StringVar(&helperTemplateCreateTurnOn, "turn-on", "", "Turn on action (switch, light, fan)")
	cmd.Flags().StringVar(&helperTemplateCreateTurnOff, "turn-off", "", "Turn off action (switch, light, fan)")
	cmd.Flags().StringVar(&helperTemplateCreatePress, "press", "", "Press action (button)")
	cmd.Flags().StringVar(&helperTemplateCreateOpen, "open", "", "Open action (cover)")
	cmd.Flags().StringVar(&helperTemplateCreateClose, "close", "", "Close action (cover)")
	cmd.Flags().StringVar(&helperTemplateCreateStop, "stop", "", "Stop action (cover)")
	cmd.Flags().StringVar(&helperTemplateCreatePosition, "position", "", "Position template (cover)")
	cmd.Flags().StringVar(&helperTemplateCreateSetPos, "set-position", "", "Set position action (cover)")
	cmd.Flags().StringVar(&helperTemplateCreateTilt, "tilt", "", "Tilt template (cover)")
	cmd.Flags().StringVar(&helperTemplateCreateSetTilt, "set-tilt", "", "Set tilt action (cover)")
	cmd.Flags().StringVar(&helperTemplateCreateLock, "lock", "", "Lock action (lock)")
	cmd.Flags().StringVar(&helperTemplateCreateUnlock, "unlock", "", "Unlock action (lock)")
	cmd.Flags().StringVar(&helperTemplateCreateURL, "url", "", "URL template (image)")
	cmd.Flags().Float64Var(&helperTemplateCreateMin, "min", 0, "Minimum value (number)")
	cmd.Flags().Float64Var(&helperTemplateCreateMax, "max", 100, "Maximum value (number)")
	cmd.Flags().Float64Var(&helperTemplateCreateStep, "step", 1, "Step value (number)")
	cmd.Flags().StringVar(&helperTemplateCreateSetValue, "set-value", "", "Set value action (number)")
	cmd.Flags().StringSliceVar(&helperTemplateCreateOptions, "options", nil, "Options (select, comma-separated)")
	cmd.Flags().StringVar(&helperTemplateCreateSelectOption, "select-option", "", "Select option action (select)")
	cmd.Flags().StringVar(&helperTemplateCreateCondition, "condition", "", "Condition template (weather)")
	cmd.Flags().StringVar(&helperTemplateCreateTemperature, "temperature", "", "Temperature template (weather)")
	cmd.Flags().StringVar(&helperTemplateCreateHumidity, "humidity", "", "Humidity template (weather)")
	cmd.Flags().StringVar(&helperTemplateCreateBrightness, "brightness", "", "Brightness template (light)")
	cmd.Flags().StringVar(&helperTemplateCreateColor, "color", "", "Color template (light)")
	cmd.Flags().StringVar(&helperTemplateCreateEffect, "effect", "", "Effect template (light)")
	cmd.Flags().StringSliceVar(&helperTemplateCreateEffects, "effects", nil, "Effect list (light, comma-separated)")
	cmd.Flags().StringVar(&helperTemplateCreatePercentage, "percentage", "", "Percentage template (fan)")
	cmd.Flags().StringVar(&helperTemplateCreateSetPct, "set-percentage", "", "Set percentage action (fan)")
	cmd.Flags().StringVar(&helperTemplateCreatePreset, "preset", "", "Preset mode template (fan)")
	cmd.Flags().StringVar(&helperTemplateCreateSetPreset, "set-preset", "", "Set preset mode action (fan)")
	cmd.Flags().StringVar(&helperTemplateCreateOscillate, "oscillate", "", "Oscillating template (fan)")
	cmd.Flags().StringVar(&helperTemplateCreateOscOn, "oscillate-on", "", "Set oscillating on action (fan)")
	cmd.Flags().StringVar(&helperTemplateCreateOscOff, "oscillate-off", "", "Set oscillating off action (fan)")
	cmd.Flags().StringVar(&helperTemplateCreateDirection, "direction", "", "Direction template (fan)")
	cmd.Flags().StringVar(&helperTemplateCreateSetDir, "set-direction", "", "Set direction action (fan)")
	cmd.Flags().StringVar(&helperTemplateCreateStart, "start", "", "Start action (vacuum)")
	cmd.Flags().StringVar(&helperTemplateCreatePause, "pause", "", "Pause action (vacuum)")
	cmd.Flags().StringVar(&helperTemplateCreateReturnToBase, "return-to-base", "", "Return to base action (vacuum)")
	cmd.Flags().StringVar(&helperTemplateCreateClean, "clean-spot", "", "Clean spot action (vacuum)")
	cmd.Flags().StringVar(&helperTemplateCreateLocate, "locate", "", "Locate action (vacuum)")
	cmd.Flags().StringVar(&helperTemplateCreateSetFanSpeed, "set-fan-speed", "", "Set fan speed action (vacuum)")
	cmd.Flags().StringVar(&helperTemplateCreateFanSpeed, "fan-speed", "", "Fan speed template (vacuum)")
	cmd.Flags().StringVar(&helperTemplateCreateBattery, "battery", "", "Battery level template (vacuum)")
}

// runTemplateCreate handles the template create command.
func runTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	textMode := getTextMode()

	validTypes := map[string]bool{
		"alarm_control_panel": true, "binary_sensor": true, "button": true,
		"image": true, "number": true, "select": true, "sensor": true, "switch": true,
	}
	if !validTypes[helperTemplateCreateType] {
		return fmt.Errorf("invalid template type: %s. Valid types: alarm_control_panel, binary_sensor, button, image, number, select, sensor, switch", helperTemplateCreateType)
	}

	rest, err := getRESTClient()
	if err != nil {
		return err
	}

	flowResult, err := rest.ConfigFlowCreate("template")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}
	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Menu step: select template type
	menuResult, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
		"next_step_id": helperTemplateCreateType,
	})
	if err != nil {
		return fmt.Errorf("failed to select template type: %w", err)
	}
	if t, _ := menuResult["type"].(string); t == "abort" {
		reason, _ := menuResult["reason"].(string)
		return fmt.Errorf("config flow aborted: %s", reason)
	}

	formData := buildTemplateFormData(name)

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create template entity: %w", err)
	}

	resultType, _ := finalResult["type"].(string)
	if resultType == "abort" {
		reason, _ := finalResult["reason"].(string)
		return fmt.Errorf("config flow aborted: %s", reason)
	}
	if resultType != "create_entry" {
		if errors, ok := finalResult["errors"].(map[string]interface{}); ok && len(errors) > 0 {
			return fmt.Errorf("validation errors: %v", errors)
		}
		return fmt.Errorf("unexpected flow result type: %s", resultType)
	}

	result := map[string]interface{}{
		"title": finalResult["title"],
		"type":  helperTemplateCreateType,
	}
	if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
		if entryID, ok := entryResult["entry_id"]; ok {
			result["entry_id"] = entryID
		}
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Template %s '%s' created successfully.", helperTemplateCreateType, name))
	return nil
}

func buildTemplateFormData(name string) map[string]interface{} {
	formData := map[string]interface{}{"name": name}

	if helperTemplateCreateIcon != "" {
		formData["icon"] = helperTemplateCreateIcon
	}

	switch helperTemplateCreateType {
	case "sensor":
		if helperTemplateCreateStateTemplate != "" {
			formData["state"] = helperTemplateCreateStateTemplate
		}
		if helperTemplateCreateUnit != "" {
			formData["unit_of_measurement"] = helperTemplateCreateUnit
		}
		if helperTemplateCreateDeviceClass != "" {
			formData["device_class"] = helperTemplateCreateDeviceClass
		}
		if helperTemplateCreateStateClass != "" {
			formData["state_class"] = helperTemplateCreateStateClass
		}
	case "binary_sensor":
		if helperTemplateCreateStateTemplate != "" {
			formData["state"] = helperTemplateCreateStateTemplate
		}
		if helperTemplateCreateDeviceClass != "" {
			formData["device_class"] = helperTemplateCreateDeviceClass
		}
	case "alarm_control_panel":
		if helperTemplateCreateStateTemplate != "" {
			formData["value_template"] = helperTemplateCreateStateTemplate
		}
	case "switch":
		if helperTemplateCreateStateTemplate != "" {
			formData["value_template"] = helperTemplateCreateStateTemplate
		}
		if helperTemplateCreateTurnOn != "" {
			formData["turn_on"] = buildActionSequence(helperTemplateCreateTurnOn)
		}
		if helperTemplateCreateTurnOff != "" {
			formData["turn_off"] = buildActionSequence(helperTemplateCreateTurnOff)
		}
	case "button":
		if helperTemplateCreatePress != "" {
			formData["press"] = buildActionSequence(helperTemplateCreatePress)
		}
	case "cover":
		if helperTemplateCreateStateTemplate != "" {
			formData["value_template"] = helperTemplateCreateStateTemplate
		}
		if helperTemplateCreateDeviceClass != "" {
			formData["device_class"] = helperTemplateCreateDeviceClass
		}
		if helperTemplateCreatePosition != "" {
			formData["position_template"] = helperTemplateCreatePosition
		}
		if helperTemplateCreateOpen != "" {
			formData["open_cover"] = buildActionSequence(helperTemplateCreateOpen)
		}
		if helperTemplateCreateClose != "" {
			formData["close_cover"] = buildActionSequence(helperTemplateCreateClose)
		}
		if helperTemplateCreateStop != "" {
			formData["stop_cover"] = buildActionSequence(helperTemplateCreateStop)
		}
		if helperTemplateCreateSetPos != "" {
			formData["set_cover_position"] = buildActionSequence(helperTemplateCreateSetPos)
		}
		if helperTemplateCreateTilt != "" {
			formData["tilt_template"] = helperTemplateCreateTilt
		}
		if helperTemplateCreateSetTilt != "" {
			formData["set_cover_tilt_position"] = buildActionSequence(helperTemplateCreateSetTilt)
		}
	case "lock":
		if helperTemplateCreateStateTemplate != "" {
			formData["value_template"] = helperTemplateCreateStateTemplate
		}
		if helperTemplateCreateLock != "" {
			formData["lock"] = buildActionSequence(helperTemplateCreateLock)
		}
		if helperTemplateCreateUnlock != "" {
			formData["unlock"] = buildActionSequence(helperTemplateCreateUnlock)
		}
	case "light":
		if helperTemplateCreateStateTemplate != "" {
			formData["value_template"] = helperTemplateCreateStateTemplate
		}
		if helperTemplateCreateTurnOn != "" {
			formData["turn_on"] = buildActionSequence(helperTemplateCreateTurnOn)
		}
		if helperTemplateCreateTurnOff != "" {
			formData["turn_off"] = buildActionSequence(helperTemplateCreateTurnOff)
		}
		if helperTemplateCreateBrightness != "" {
			formData["level_template"] = helperTemplateCreateBrightness
		}
		if helperTemplateCreateColor != "" {
			formData["color_template"] = helperTemplateCreateColor
		}
		if helperTemplateCreateEffect != "" {
			formData["effect_template"] = helperTemplateCreateEffect
		}
		if len(helperTemplateCreateEffects) > 0 {
			formData["effect_list"] = helperTemplateCreateEffects
		}
	case "fan":
		if helperTemplateCreateStateTemplate != "" {
			formData["value_template"] = helperTemplateCreateStateTemplate
		}
		if helperTemplateCreateTurnOn != "" {
			formData["turn_on"] = buildActionSequence(helperTemplateCreateTurnOn)
		}
		if helperTemplateCreateTurnOff != "" {
			formData["turn_off"] = buildActionSequence(helperTemplateCreateTurnOff)
		}
		if helperTemplateCreatePercentage != "" {
			formData["percentage_template"] = helperTemplateCreatePercentage
		}
		if helperTemplateCreateSetPct != "" {
			formData["set_percentage"] = buildActionSequence(helperTemplateCreateSetPct)
		}
		if helperTemplateCreatePreset != "" {
			formData["preset_mode_template"] = helperTemplateCreatePreset
		}
		if helperTemplateCreateSetPreset != "" {
			formData["set_preset_mode"] = buildActionSequence(helperTemplateCreateSetPreset)
		}
		if helperTemplateCreateOscillate != "" {
			formData["oscillating_template"] = helperTemplateCreateOscillate
		}
		if helperTemplateCreateOscOn != "" {
			formData["set_oscillating"] = buildActionSequence(helperTemplateCreateOscOn)
		}
		if helperTemplateCreateDirection != "" {
			formData["direction_template"] = helperTemplateCreateDirection
		}
		if helperTemplateCreateSetDir != "" {
			formData["set_direction"] = buildActionSequence(helperTemplateCreateSetDir)
		}
	case "vacuum":
		if helperTemplateCreateStateTemplate != "" {
			formData["value_template"] = helperTemplateCreateStateTemplate
		}
		if helperTemplateCreateStart != "" {
			formData["start"] = buildActionSequence(helperTemplateCreateStart)
		}
		if helperTemplateCreatePause != "" {
			formData["pause"] = buildActionSequence(helperTemplateCreatePause)
		}
		if helperTemplateCreateReturnToBase != "" {
			formData["return_to_base"] = buildActionSequence(helperTemplateCreateReturnToBase)
		}
		if helperTemplateCreateClean != "" {
			formData["clean_spot"] = buildActionSequence(helperTemplateCreateClean)
		}
		if helperTemplateCreateLocate != "" {
			formData["locate"] = buildActionSequence(helperTemplateCreateLocate)
		}
		if helperTemplateCreateSetFanSpeed != "" {
			formData["set_fan_speed"] = buildActionSequence(helperTemplateCreateSetFanSpeed)
		}
		if helperTemplateCreateFanSpeed != "" {
			formData["fan_speed_template"] = helperTemplateCreateFanSpeed
		}
		if helperTemplateCreateBattery != "" {
			formData["battery_level_template"] = helperTemplateCreateBattery
		}
	case "image":
		if helperTemplateCreateURL != "" {
			formData["url"] = helperTemplateCreateURL
		}
	case "number":
		if helperTemplateCreateStateTemplate != "" {
			formData["state"] = helperTemplateCreateStateTemplate
		}
		formData["min"] = helperTemplateCreateMin
		formData["max"] = helperTemplateCreateMax
		formData["step"] = helperTemplateCreateStep
		if helperTemplateCreateSetValue != "" {
			formData["set_value"] = buildActionSequence(helperTemplateCreateSetValue)
		}
	case "select":
		if helperTemplateCreateStateTemplate != "" {
			formData["state"] = helperTemplateCreateStateTemplate
		}
		if len(helperTemplateCreateOptions) > 0 {
			formData["options"] = buildOptionsTemplate(helperTemplateCreateOptions)
		}
		if helperTemplateCreateSelectOption != "" {
			formData["select_option"] = buildActionSequence(helperTemplateCreateSelectOption)
		}
	case "weather":
		if helperTemplateCreateCondition != "" {
			formData["condition_template"] = helperTemplateCreateCondition
		}
		if helperTemplateCreateTemperature != "" {
			formData["temperature_template"] = helperTemplateCreateTemperature
		}
		if helperTemplateCreateHumidity != "" {
			formData["humidity_template"] = helperTemplateCreateHumidity
		}
	}

	return formData
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
