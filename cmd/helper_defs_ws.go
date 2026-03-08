package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	// ===== WS-based helpers =====

	registerInputBoolean()
	registerInputNumber()
	registerInputText()
	registerInputSelect()
	registerInputDatetime()
	registerInputButton()
	registerCounter()
	registerTimer()
	registerSchedule()

	// ===== Config-flow-based helpers =====

	registerDerivative()
	registerIntegration()
	registerMinMax()
	registerThreshold()
	registerUtilityMeter()
	registerStatistics()
	registerLocalCalendar()
	registerLocalTodo()
	registerGroup()
	registerTemplate()
}

// ========== WS helpers ==========

func registerInputBoolean() {
	var icon string
	var initial bool

	registerHelperType(HelperDef{
		TypeName:        "input_boolean",
		CommandName:     "input-boolean",
		DisplayName:     "input boolean",
		Short:           "Manage input boolean helpers",
		Long:            "Create, list, and delete input boolean (toggle) helpers.",
		Category:        HelperCategoryWS,
		TypeDescription: "A boolean on/off toggle helper",
		CreateParams:    []string{"name (required)", "icon", "initial (true/false)"},
		CreateShort:     "Create a new input boolean helper",
		CreateLong:      "Create a new input boolean (toggle) helper.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the helper (e.g., mdi:toggle-switch)")
			cmd.Flags().BoolVar(&initial, "initial", false, "Initial value (true/false)")
		},
		RunCreate: helperWSCreate("input_boolean", "input boolean", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			params := map[string]interface{}{"name": name}
			if icon != "" {
				params["icon"] = icon
			}
			if cmd.Flags().Changed("initial") {
				params["initial"] = initial
			}
			return params, nil
		}),
	})
}

func registerInputNumber() {
	var icon, mode, unit string
	var min, max, step, initial float64

	registerHelperType(HelperDef{
		TypeName:        "input_number",
		CommandName:     "input-number",
		DisplayName:     "input number",
		Short:           "Manage input number helpers",
		Long:            "Create, list, and delete input number helpers.",
		Category:        HelperCategoryWS,
		TypeDescription: "A numeric value helper with min/max range",
		CreateParams:    []string{"name (required)", "min (required)", "max (required)", "icon", "initial", "step", "mode (box/slider)", "unit_of_measurement"},
		CreateShort:     "Create a new input number helper",
		CreateLong:      "Create a new input number helper with configurable min/max/step values.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the helper")
			cmd.Flags().Float64Var(&min, "min", 0, "Minimum value (required)")
			cmd.Flags().Float64Var(&max, "max", 100, "Maximum value (required)")
			cmd.Flags().Float64Var(&step, "step", 1, "Step value")
			cmd.Flags().Float64Var(&initial, "initial", 0, "Initial value")
			cmd.Flags().StringVar(&mode, "mode", "slider", "Display mode: slider or box")
			cmd.Flags().StringVar(&unit, "unit", "", "Unit of measurement")
		},
		RequiredFlags: []string{"min", "max"},
		RunCreate: helperWSCreate("input_number", "input number", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			params := map[string]interface{}{
				"name": name,
				"min":  min,
				"max":  max,
			}
			if icon != "" {
				params["icon"] = icon
			}
			if cmd.Flags().Changed("step") {
				params["step"] = step
			}
			if cmd.Flags().Changed("initial") {
				params["initial"] = initial
			}
			if mode != "" {
				params["mode"] = mode
			}
			if unit != "" {
				params["unit_of_measurement"] = unit
			}
			return params, nil
		}),
	})
}

func registerInputText() {
	var icon, initial, pattern, mode string
	var minLen, maxLen int

	registerHelperType(HelperDef{
		TypeName:        "input_text",
		CommandName:     "input-text",
		DisplayName:     "input text",
		Short:           "Manage input text helpers",
		Long:            "Create, list, and delete input text helpers.",
		Category:        HelperCategoryWS,
		TypeDescription: "A text input helper",
		CreateParams:    []string{"name (required)", "icon", "initial", "min", "max", "pattern", "mode (text/password)"},
		CreateShort:     "Create a new input text helper",
		CreateLong:      "Create a new input text helper.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the helper")
			cmd.Flags().StringVar(&initial, "initial", "", "Initial value")
			cmd.Flags().IntVar(&minLen, "min", 0, "Minimum text length")
			cmd.Flags().IntVar(&maxLen, "max", 100, "Maximum text length")
			cmd.Flags().StringVar(&pattern, "pattern", "", "Regex pattern for validation")
			cmd.Flags().StringVar(&mode, "mode", "text", "Input mode: text or password")
		},
		RunCreate: helperWSCreate("input_text", "input text", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			params := map[string]interface{}{"name": name}
			if icon != "" {
				params["icon"] = icon
			}
			if initial != "" {
				params["initial"] = initial
			}
			if cmd.Flags().Changed("min") {
				params["min"] = minLen
			}
			if cmd.Flags().Changed("max") {
				params["max"] = maxLen
			}
			if pattern != "" {
				params["pattern"] = pattern
			}
			if mode != "" {
				params["mode"] = mode
			}
			return params, nil
		}),
	})
}

func registerInputSelect() {
	var icon, initial string
	var options []string

	registerHelperType(HelperDef{
		TypeName:        "input_select",
		CommandName:     "input-select",
		DisplayName:     "input select",
		Short:           "Manage input select helpers",
		Long:            "Create, list, and delete input select (dropdown) helpers.",
		Category:        HelperCategoryWS,
		TypeDescription: "A dropdown selection helper",
		CreateParams:    []string{"name (required)", "options (required, array)", "icon", "initial"},
		CreateShort:     "Create a new input select helper",
		CreateLong:      "Create a new input select (dropdown) helper.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the helper")
			cmd.Flags().StringSliceVarP(&options, "options", "o", nil, "Options for the select (required)")
			cmd.Flags().StringVar(&initial, "initial", "", "Initial selected value")
		},
		RequiredFlags: []string{"options"},
		RunCreate: helperWSCreate("input_select", "input select", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			params := map[string]interface{}{
				"name":    name,
				"options": options,
			}
			if icon != "" {
				params["icon"] = icon
			}
			if initial != "" {
				params["initial"] = initial
			}
			return params, nil
		}),
	})
}

func registerInputDatetime() {
	var icon string
	var hasDate, hasTime bool
	var initial string

	registerHelperType(HelperDef{
		TypeName:        "input_datetime",
		CommandName:     "input-datetime",
		DisplayName:     "input datetime",
		Short:           "Manage input datetime helpers",
		Long:            "Create, list, and delete input datetime helpers.",
		Category:        HelperCategoryWS,
		TypeDescription: "A date/time helper",
		CreateParams:    []string{"name (required)", "has_date (required)", "has_time (required)", "icon", "initial"},
		CreateShort:     "Create a new input datetime helper",
		CreateLong:      "Create a new input datetime helper. At least one of --has-date or --has-time must be specified.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the helper")
			cmd.Flags().BoolVar(&hasDate, "has-date", false, "Include date component")
			cmd.Flags().BoolVar(&hasTime, "has-time", false, "Include time component")
			cmd.Flags().StringVar(&initial, "initial", "", "Initial value")
		},
		RunCreate: helperWSCreate("input_datetime", "input datetime", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			if !hasDate && !hasTime {
				return nil, fmt.Errorf("at least one of --has-date or --has-time must be specified")
			}
			params := map[string]interface{}{
				"name":     name,
				"has_date": hasDate,
				"has_time": hasTime,
			}
			if icon != "" {
				params["icon"] = icon
			}
			if initial != "" {
				params["initial"] = initial
			}
			return params, nil
		}),
	})
}

func registerInputButton() {
	var icon string

	registerHelperType(HelperDef{
		TypeName:        "input_button",
		CommandName:     "input-button",
		DisplayName:     "input button",
		Short:           "Manage input button helpers",
		Long:            "Create, list, and delete input button helpers.",
		Category:        HelperCategoryWS,
		TypeDescription: "A button helper that can be pressed",
		CreateParams:    []string{"name (required)", "icon"},
		CreateShort:     "Create a new input button helper",
		CreateLong:      "Create a new input button helper.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the helper (e.g., mdi:button-pointer)")
		},
		RunCreate: helperWSCreate("input_button", "input button", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			params := map[string]interface{}{"name": name}
			if icon != "" {
				params["icon"] = icon
			}
			return params, nil
		}),
	})
}

func registerCounter() {
	var icon string
	var initial, minimum, maximum, step int
	var restore bool

	registerHelperType(HelperDef{
		TypeName:        "counter",
		CommandName:     "counter",
		DisplayName:     "counter",
		Short:           "Manage counter helpers",
		Long:            "Create, list, and delete counter helpers.",
		Category:        HelperCategoryWS,
		TypeDescription: "A counter helper that can be incremented/decremented",
		CreateParams:    []string{"name (required)", "icon", "initial", "minimum", "maximum", "step", "restore (true/false)"},
		CreateShort:     "Create a new counter helper",
		CreateLong:      "Create a new counter helper that can be incremented/decremented.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the helper")
			cmd.Flags().IntVar(&initial, "initial", 0, "Initial value")
			cmd.Flags().IntVar(&minimum, "minimum", 0, "Minimum value")
			cmd.Flags().IntVar(&maximum, "maximum", 0, "Maximum value (0 for no limit)")
			cmd.Flags().IntVar(&step, "step", 1, "Step value for increment/decrement")
			cmd.Flags().BoolVar(&restore, "restore", true, "Restore value after restart")
		},
		RunCreate: helperWSCreate("counter", "counter", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			params := map[string]interface{}{"name": name}
			if icon != "" {
				params["icon"] = icon
			}
			if cmd.Flags().Changed("initial") {
				params["initial"] = initial
			}
			if cmd.Flags().Changed("minimum") {
				params["minimum"] = minimum
			}
			if cmd.Flags().Changed("maximum") && maximum != 0 {
				params["maximum"] = maximum
			}
			if cmd.Flags().Changed("step") {
				params["step"] = step
			}
			if cmd.Flags().Changed("restore") {
				params["restore"] = restore
			}
			return params, nil
		}),
	})
}

func registerTimer() {
	var icon, duration string
	var restore bool

	registerHelperType(HelperDef{
		TypeName:        "timer",
		CommandName:     "timer",
		DisplayName:     "timer",
		Short:           "Manage timer helpers",
		Long:            "Create, list, and delete timer helpers.",
		Category:        HelperCategoryWS,
		TypeDescription: "A timer helper that counts down",
		CreateParams:    []string{"name (required)", "icon", "duration", "restore (true/false)"},
		CreateShort:     "Create a new timer helper",
		CreateLong:      "Create a new timer helper.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the helper")
			cmd.Flags().StringVarP(&duration, "duration", "d", "", "Default duration (e.g., 00:05:00)")
			cmd.Flags().BoolVar(&restore, "restore", true, "Restore timer after restart")
		},
		RunCreate: helperWSCreate("timer", "timer", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			params := map[string]interface{}{"name": name}
			if icon != "" {
				params["icon"] = icon
			}
			if duration != "" {
				params["duration"] = duration
			}
			if cmd.Flags().Changed("restore") {
				params["restore"] = restore
			}
			return params, nil
		}),
	})
}

func registerSchedule() {
	var icon string

	registerHelperType(HelperDef{
		TypeName:        "schedule",
		CommandName:     "schedule",
		DisplayName:     "schedule",
		Short:           "Manage schedule helpers",
		Long:            "Create, list, and delete schedule helpers.",
		Category:        HelperCategoryWS,
		TypeDescription: "A schedule helper for time-based automation",
		CreateParams:    []string{"name (required)", "icon"},
		CreateShort:     "Create a new schedule helper",
		CreateLong:      "Create a new schedule helper.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the helper")
		},
		RunCreate: helperWSCreate("schedule", "schedule", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			params := map[string]interface{}{"name": name}
			if icon != "" {
				params["icon"] = icon
			}
			return params, nil
		}),
	})
}
