package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

var helperTemplateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new template entity",
	Long: `Create a new template entity helper using the config entry flow.

Template types available: alarm_control_panel, binary_sensor, button, image, number, select, sensor, switch.

Templates use Jinja2 syntax. State templates should return valid values for the entity type.

Examples:
  # Create a template binary sensor
  hab helper-template create "Is Sun Up" --type binary_sensor --state "{{ is_state('sun.sun', 'above_horizon') }}"

  # Create a template sensor with unit
  hab helper-template create "Room Temperature" --type sensor --state "{{ states('sensor.temp1') | float + states('sensor.temp2') | float }}" --unit "°C"

  # Create a template switch with actions
  hab helper-template create "All Lights" --type switch --state "{{ is_state('light.living_room', 'on') }}" --turn-on "homeassistant.turn_on" --turn-off "homeassistant.turn_off"

  # Create a template button
  hab helper-template create "Reset Counter" --type button --press "counter.reset"

  # Create a template number
  hab helper-template create "Volume" --type number --state "{{ states('input_number.volume') | float }}" --min 0 --max 100 --step 5 --set-value "input_number.set_value"

  # Create a template select
  hab helper-template create "Mode" --type select --state "{{ states('input_select.mode') }}" --options "auto,cool,heat,off" --select-option "input_select.select_option"

  # Create a template image
  hab helper-template create "Camera Snapshot" --type image --url "http://example.com/image.jpg"`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperTemplateCreate,
}

func init() {
	helperTemplateParentCmd.AddCommand(helperTemplateCreateCmd)

	// Type selection
	helperTemplateCreateCmd.Flags().StringVarP(&helperTemplateCreateType, "type", "t", "sensor", "Template type: alarm_control_panel, binary_sensor, button, image, number, select, sensor, switch")

	// Common options
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateStateTemplate, "state", "", "State template (Jinja2)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateIcon, "icon", "", "Icon (e.g., mdi:thermometer)")

	// Sensor specific
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateUnit, "unit", "", "Unit of measurement (sensor)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateDeviceClass, "device-class", "", "Device class (sensor, binary_sensor, cover)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateStateClass, "state-class", "", "State class: measurement, total, total_increasing (sensor)")

	// Switch/Light/Fan actions
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateTurnOn, "turn-on", "", "Turn on action (switch, light, fan)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateTurnOff, "turn-off", "", "Turn off action (switch, light, fan)")

	// Button
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreatePress, "press", "", "Press action (button)")

	// Cover
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateOpen, "open", "", "Open action (cover)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateClose, "close", "", "Close action (cover)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateStop, "stop", "", "Stop action (cover)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreatePosition, "position", "", "Position template (cover)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateSetPos, "set-position", "", "Set position action (cover)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateTilt, "tilt", "", "Tilt template (cover)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateSetTilt, "set-tilt", "", "Set tilt action (cover)")

	// Lock
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateLock, "lock", "", "Lock action (lock)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateUnlock, "unlock", "", "Unlock action (lock)")

	// Image
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateURL, "url", "", "URL template (image)")

	// Number
	helperTemplateCreateCmd.Flags().Float64Var(&helperTemplateCreateMin, "min", 0, "Minimum value (number)")
	helperTemplateCreateCmd.Flags().Float64Var(&helperTemplateCreateMax, "max", 100, "Maximum value (number)")
	helperTemplateCreateCmd.Flags().Float64Var(&helperTemplateCreateStep, "step", 1, "Step value (number)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateSetValue, "set-value", "", "Set value action (number)")

	// Select
	helperTemplateCreateCmd.Flags().StringSliceVar(&helperTemplateCreateOptions, "options", nil, "Options (select, comma-separated)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateSelectOption, "select-option", "", "Select option action (select)")

	// Weather
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateCondition, "condition", "", "Condition template (weather)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateTemperature, "temperature", "", "Temperature template (weather)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateHumidity, "humidity", "", "Humidity template (weather)")

	// Light specific
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateBrightness, "brightness", "", "Brightness template (light)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateColor, "color", "", "Color template (light)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateEffect, "effect", "", "Effect template (light)")
	helperTemplateCreateCmd.Flags().StringSliceVar(&helperTemplateCreateEffects, "effects", nil, "Effect list (light, comma-separated)")

	// Fan specific
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreatePercentage, "percentage", "", "Percentage template (fan)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateSetPct, "set-percentage", "", "Set percentage action (fan)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreatePreset, "preset", "", "Preset mode template (fan)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateSetPreset, "set-preset", "", "Set preset mode action (fan)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateOscillate, "oscillate", "", "Oscillating template (fan)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateOscOn, "oscillate-on", "", "Set oscillating on action (fan)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateOscOff, "oscillate-off", "", "Set oscillating off action (fan)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateDirection, "direction", "", "Direction template (fan)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateSetDir, "set-direction", "", "Set direction action (fan)")

	// Vacuum
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateStart, "start", "", "Start action (vacuum)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreatePause, "pause", "", "Pause action (vacuum)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateReturnToBase, "return-to-base", "", "Return to base action (vacuum)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateClean, "clean-spot", "", "Clean spot action (vacuum)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateLocate, "locate", "", "Locate action (vacuum)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateSetFanSpeed, "set-fan-speed", "", "Set fan speed action (vacuum)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateFanSpeed, "fan-speed", "", "Fan speed template (vacuum)")
	helperTemplateCreateCmd.Flags().StringVar(&helperTemplateCreateBattery, "battery", "", "Battery level template (vacuum)")
}

func runHelperTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Validate template type
	// These are the types supported by the Home Assistant template config flow
	validTypes := map[string]bool{
		"alarm_control_panel": true,
		"binary_sensor":       true,
		"button":              true,
		"image":               true,
		"number":              true,
		"select":              true,
		"sensor":              true,
		"switch":              true,
	}
	if !validTypes[helperTemplateCreateType] {
		return fmt.Errorf("invalid template type: %s. Valid types: alarm_control_panel, binary_sensor, button, image, number, select, sensor, switch", helperTemplateCreateType)
	}

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	// Use REST API for config flows
	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// Step 1: Start the config flow for template
	flowResult, err := rest.ConfigFlowCreate("template")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}

	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Step 2: Select the template type (menu step)
	menuResult, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
		"next_step_id": helperTemplateCreateType,
	})
	if err != nil {
		return fmt.Errorf("failed to select template type: %w", err)
	}

	// Check if we need another step
	stepType, _ := menuResult["type"].(string)
	if stepType == "abort" {
		reason, _ := menuResult["reason"].(string)
		return fmt.Errorf("config flow aborted: %s", reason)
	}

	// Step 3: Submit the form data based on template type
	formData := buildTemplateFormData(name)

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create template entity: %w", err)
	}

	// Check result type
	resultType, _ := finalResult["type"].(string)
	if resultType == "abort" {
		reason, _ := finalResult["reason"].(string)
		return fmt.Errorf("config flow aborted: %s", reason)
	}

	if resultType != "create_entry" {
		// If we got a form back, there might be validation errors
		if errors, ok := finalResult["errors"].(map[string]interface{}); ok && len(errors) > 0 {
			return fmt.Errorf("validation errors: %v", errors)
		}
		return fmt.Errorf("unexpected flow result type: %s", resultType)
	}

	// Extract result data
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
	formData := map[string]interface{}{
		"name": name,
	}

	// Add icon if provided
	if helperTemplateCreateIcon != "" {
		formData["icon"] = helperTemplateCreateIcon
	}

	// Type-specific configuration
	// Note: Different template types use different field names for the state template
	// binary_sensor, sensor, number, select use "state"
	// switch, light, fan, lock, cover, vacuum use "value_template"
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
			// Convert options slice to a Jinja2 template string that returns a list
			// e.g., ["option1", "option2"] becomes "{{ ['option1', 'option2'] }}"
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

// buildOptionsTemplate converts a slice of options to a Jinja2 template string
// e.g., ["option1", "option2"] becomes "{{ ['option1', 'option2'] }}"
func buildOptionsTemplate(options []string) string {
	var items []string
	for _, opt := range options {
		items = append(items, fmt.Sprintf("'%s'", opt))
	}
	return fmt.Sprintf("{{ [%s] }}", strings.Join(items, ", "))
}

// buildActionSequence creates an action sequence from a simple action string
// For simple cases like "homeassistant.turn_on", it creates a basic action
func buildActionSequence(action string) []map[string]interface{} {
	return []map[string]interface{}{
		{"action": action},
	}
}
