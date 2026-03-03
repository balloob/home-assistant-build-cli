package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
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
		TypeName:    "input_boolean",
		CommandName: "input-boolean",
		DisplayName: "input boolean",
		Short:       "Manage input boolean helpers",
		Long:        "Create, list, and delete input boolean (toggle) helpers.",
		Category:    HelperCategoryWS,
		CreateShort: "Create a new input boolean helper",
		CreateLong:  "Create a new input boolean (toggle) helper.",
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
		TypeName:    "input_number",
		CommandName: "input-number",
		DisplayName: "input number",
		Short:       "Manage input number helpers",
		Long:        "Create, list, and delete input number helpers.",
		Category:    HelperCategoryWS,
		CreateShort: "Create a new input number helper",
		CreateLong:  "Create a new input number helper with configurable min/max/step values.",
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
		TypeName:    "input_text",
		CommandName: "input-text",
		DisplayName: "input text",
		Short:       "Manage input text helpers",
		Long:        "Create, list, and delete input text helpers.",
		Category:    HelperCategoryWS,
		CreateShort: "Create a new input text helper",
		CreateLong:  "Create a new input text helper.",
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
		TypeName:    "input_select",
		CommandName: "input-select",
		DisplayName: "input select",
		Short:       "Manage input select helpers",
		Long:        "Create, list, and delete input select (dropdown) helpers.",
		Category:    HelperCategoryWS,
		CreateShort: "Create a new input select helper",
		CreateLong:  "Create a new input select (dropdown) helper.",
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
		TypeName:    "input_datetime",
		CommandName: "input-datetime",
		DisplayName: "input datetime",
		Short:       "Manage input datetime helpers",
		Long:        "Create, list, and delete input datetime helpers.",
		Category:    HelperCategoryWS,
		CreateShort: "Create a new input datetime helper",
		CreateLong:  "Create a new input datetime helper. At least one of --has-date or --has-time must be specified.",
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
		TypeName:    "input_button",
		CommandName: "input-button",
		DisplayName: "input button",
		Short:       "Manage input button helpers",
		Long:        "Create, list, and delete input button helpers.",
		Category:    HelperCategoryWS,
		CreateShort: "Create a new input button helper",
		CreateLong:  "Create a new input button helper.",
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
		TypeName:    "counter",
		CommandName: "counter",
		DisplayName: "counter",
		Short:       "Manage counter helpers",
		Long:        "Create, list, and delete counter helpers.",
		Category:    HelperCategoryWS,
		CreateShort: "Create a new counter helper",
		CreateLong:  "Create a new counter helper that can be incremented/decremented.",
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
		TypeName:    "timer",
		CommandName: "timer",
		DisplayName: "timer",
		Short:       "Manage timer helpers",
		Long:        "Create, list, and delete timer helpers.",
		Category:    HelperCategoryWS,
		CreateShort: "Create a new timer helper",
		CreateLong:  "Create a new timer helper.",
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
		TypeName:    "schedule",
		CommandName: "schedule",
		DisplayName: "schedule",
		Short:       "Manage schedule helpers",
		Long:        "Create, list, and delete schedule helpers.",
		Category:    HelperCategoryWS,
		CreateShort: "Create a new schedule helper",
		CreateLong:  "Create a new schedule helper.",
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

// ========== Config flow helpers ==========

func registerDerivative() {
	var source, unitPrefix, unitTime, timeWindow string
	var round int

	registerHelperType(HelperDef{
		TypeName:    "derivative",
		CommandName: "derivative",
		DisplayName: "derivative sensor",
		Short:       "Manage derivative sensor helpers",
		Long:        "Create, list, and delete derivative sensor helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new derivative sensor",
		CreateLong: `Create a new derivative sensor helper that calculates the rate of change of a source sensor.

Unit prefixes: n (nano), µ (micro), m (milli), k (kilo), M (mega), G (giga), T (tera)
Time units: s (seconds), min (minutes), h (hours), d (days)

Examples:
  hab helper-derivative create "Power Rate" --source sensor.energy_usage
  hab helper-derivative create "Temperature Change" --source sensor.temperature --unit-time min --round 2
  hab helper-derivative create "Smooth Power Rate" --source sensor.power --time-window 00:05:00`,
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&source, "source", "s", "", "Source entity ID (required)")
			cmd.Flags().IntVar(&round, "round", 2, "Decimal places for rounding (0-6)")
			cmd.Flags().StringVar(&unitPrefix, "unit-prefix", "", "Metric unit prefix (n, µ, m, k, M, G, T, P)")
			cmd.Flags().StringVar(&unitTime, "unit-time", "h", "Time unit for derivative (s, min, h, d)")
			cmd.Flags().StringVar(&timeWindow, "time-window", "00:00:00", "Time window for smoothing (HH:MM:SS format)")
		},
		RequiredFlags: []string{"source"},
		RunCreate: helperConfigFlowCreate("derivative", "derivative sensor", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			validTimeUnits := map[string]bool{"s": true, "min": true, "h": true, "d": true}
			if !validTimeUnits[unitTime] {
				return nil, fmt.Errorf("invalid time unit: %s. Valid units: s, min, h, d", unitTime)
			}
			if unitPrefix != "" {
				validPrefixes := map[string]bool{"n": true, "µ": true, "m": true, "k": true, "M": true, "G": true, "T": true, "P": true}
				if !validPrefixes[unitPrefix] {
					return nil, fmt.Errorf("invalid unit prefix: %s. Valid prefixes: n, µ, m, k, M, G, T, P", unitPrefix)
				}
			}
			formData := map[string]interface{}{
				"name":        name,
				"source":      source,
				"round":       round,
				"unit_time":   unitTime,
				"time_window": parseDuration(timeWindow),
			}
			if unitPrefix != "" {
				formData["unit_prefix"] = unitPrefix
			}
			return formData, nil
		}),
	})
}

func registerIntegration() {
	var source, unitPrefix, unitTime, method string
	var round int

	registerHelperType(HelperDef{
		TypeName:    "integration",
		CommandName: "integration",
		DisplayName: "integration sensor",
		Short:       "Manage integration (integral) sensor helpers",
		Long:        "Create, list, and delete integration (Riemann sum integral) sensor helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new integration sensor",
		CreateLong: `Create a new integration (Riemann sum integral) sensor helper.

Examples:
  hab helper-integration create "Total Energy" --source sensor.power
  hab helper-integration create "Water Usage" --source sensor.flow_rate --method trapezoidal`,
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&source, "source", "s", "", "Source entity ID (required)")
			cmd.Flags().IntVar(&round, "round", 3, "Decimal places for rounding")
			cmd.Flags().StringVar(&unitPrefix, "unit-prefix", "", "Metric unit prefix (n, µ, m, k, M, G, T, P)")
			cmd.Flags().StringVar(&unitTime, "unit-time", "h", "Time unit (s, min, h, d)")
			cmd.Flags().StringVar(&method, "method", "trapezoidal", "Integration method: left, right, trapezoidal")
		},
		RequiredFlags: []string{"source"},
		RunCreate: helperConfigFlowCreate("integration", "integration sensor", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			validTimeUnits := map[string]bool{"s": true, "min": true, "h": true, "d": true}
			if !validTimeUnits[unitTime] {
				return nil, fmt.Errorf("invalid time unit: %s. Valid units: s, min, h, d", unitTime)
			}
			validMethods := map[string]bool{"left": true, "right": true, "trapezoidal": true}
			if !validMethods[method] {
				return nil, fmt.Errorf("invalid method: %s. Valid methods: left, right, trapezoidal", method)
			}
			if unitPrefix != "" {
				validPrefixes := map[string]bool{"n": true, "µ": true, "m": true, "k": true, "M": true, "G": true, "T": true, "P": true}
				if !validPrefixes[unitPrefix] {
					return nil, fmt.Errorf("invalid unit prefix: %s", unitPrefix)
				}
			}
			formData := map[string]interface{}{
				"name":      name,
				"source":    source,
				"round":     round,
				"unit_time": unitTime,
				"method":    method,
			}
			if unitPrefix != "" {
				formData["unit_prefix"] = unitPrefix
			}
			return formData, nil
		}),
	})
}

func registerMinMax() {
	var minMaxType string
	var entities []string
	var round int

	registerHelperType(HelperDef{
		TypeName:    "min_max",
		CommandName: "min-max",
		DisplayName: "min/max sensor",
		Short:       "Manage min/max sensor helpers",
		Long:        "Create, list, and delete min/max sensor helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new min/max sensor",
		CreateLong: `Create a new min/max sensor helper.

Examples:
  hab helper-min-max create "Highest Temp" --type max --entities sensor.temp1,sensor.temp2`,
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringSliceVarP(&entities, "entities", "e", nil, "Source entity IDs (required)")
			cmd.Flags().StringVarP(&minMaxType, "type", "t", "max", "Aggregation type: min, max, mean, median, last, range")
			cmd.Flags().IntVar(&round, "round", 2, "Decimal places for rounding")
		},
		RequiredFlags: []string{"entities"},
		RunCreate: helperConfigFlowCreate("min_max", "min/max sensor", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			return map[string]interface{}{
				"name":         name,
				"entity_ids":   entities,
				"type":         minMaxType,
				"round_digits": round,
			}, nil
		}),
	})
}

func registerThreshold() {
	var entity string
	var lower, upper, hysteresis float64

	registerHelperType(HelperDef{
		TypeName:    "threshold",
		CommandName: "threshold",
		DisplayName: "threshold sensor",
		Short:       "Manage threshold sensor helpers",
		Long:        "Create, list, and delete threshold binary sensor helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new threshold sensor",
		CreateLong: `Create a new threshold binary sensor helper. At least one of --lower or --upper must be specified.

Examples:
  hab helper-threshold create "Freezing Alert" --entity sensor.temperature --lower 0
  hab helper-threshold create "Overheat Alert" --entity sensor.cpu_temp --upper 80 --hysteresis 5`,
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&entity, "entity", "e", "", "Source entity ID (required)")
			cmd.Flags().Float64VarP(&lower, "lower", "l", 0, "Lower threshold")
			cmd.Flags().Float64VarP(&upper, "upper", "u", 0, "Upper threshold")
			cmd.Flags().Float64Var(&hysteresis, "hysteresis", 0, "Hysteresis value")
		},
		RequiredFlags: []string{"entity"},
		RunCreate: helperConfigFlowCreate("threshold", "threshold sensor", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			if !cmd.Flags().Changed("lower") && !cmd.Flags().Changed("upper") {
				return nil, fmt.Errorf("at least one of --lower or --upper must be specified")
			}
			formData := map[string]interface{}{
				"name":       name,
				"entity_id":  entity,
				"hysteresis": hysteresis,
			}
			if cmd.Flags().Changed("lower") {
				formData["lower"] = lower
			}
			if cmd.Flags().Changed("upper") {
				formData["upper"] = upper
			}
			return formData, nil
		}),
	})
}

func registerUtilityMeter() {
	var source, cycle string
	var offset int
	var tariffs []string
	var deltaValues, netConsumption, periodicallyResetting, alwaysAvailable bool

	registerHelperType(HelperDef{
		TypeName:    "utility_meter",
		CommandName: "utility-meter",
		DisplayName: "utility meter",
		Short:       "Manage utility meter helpers",
		Long:        "Create, list, and delete utility meter helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new utility meter",
		CreateLong: `Create a new utility meter helper that tracks consumption across billing cycles.

Cycle options: quarter-hourly, hourly, daily, weekly, monthly, bimonthly, quarterly, yearly

Examples:
  hab helper-utility-meter create "Monthly Energy" --source sensor.total_energy --cycle monthly
  hab helper-utility-meter create "Daily Water" --source sensor.water_meter --cycle daily --delta-values
  hab helper-utility-meter create "Electric Bill" --source sensor.power --cycle monthly --tariffs "peak,off-peak"`,
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&source, "source", "s", "", "Source entity ID (required)")
			cmd.Flags().StringVarP(&cycle, "cycle", "c", "monthly", "Reset cycle: quarter-hourly, hourly, daily, weekly, monthly, bimonthly, quarterly, yearly")
			cmd.Flags().IntVar(&offset, "offset", 0, "Offset in days for cycle reset")
			cmd.Flags().StringSliceVar(&tariffs, "tariffs", nil, "Tariff names for multi-rate billing")
			cmd.Flags().BoolVar(&deltaValues, "delta-values", false, "Source provides delta values (incremental)")
			cmd.Flags().BoolVar(&netConsumption, "net-consumption", false, "Net meter that can increase/decrease")
			cmd.Flags().BoolVar(&periodicallyResetting, "periodically-resetting", true, "Source may reset to 0 independently")
			cmd.Flags().BoolVar(&alwaysAvailable, "always-available", false, "Maintain last value when source unavailable")
		},
		RequiredFlags: []string{"source"},
		RunCreate: helperConfigFlowCreate("utility_meter", "utility meter", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			cycleMap := map[string]string{
				"none": "none", "quarter-hourly": "quarter_hourly",
				"hourly": "hourly", "daily": "daily", "weekly": "weekly",
				"monthly": "monthly", "bimonthly": "bimonthly",
				"quarterly": "quarterly", "yearly": "yearly",
			}
			meterType, ok := cycleMap[cycle]
			if !ok {
				return nil, fmt.Errorf("invalid cycle: %s. Valid cycles: none, quarter-hourly, hourly, daily, weekly, monthly, bimonthly, quarterly, yearly", cycle)
			}
			t := tariffs
			if t == nil {
				t = []string{}
			}
			return map[string]interface{}{
				"name":                   name,
				"source":                 source,
				"cycle":                  meterType,
				"offset":                 offset,
				"delta_values":           deltaValues,
				"net_consumption":        netConsumption,
				"periodically_resetting": periodicallyResetting,
				"tariffs":               t,
				"always_available":       alwaysAvailable,
			}, nil
		}),
	})
}

func registerStatistics() {
	var entity, characteristic, maxAge string
	var samplingSize, precision, percentile int

	registerHelperType(HelperDef{
		TypeName:    "statistics",
		CommandName: "statistics",
		DisplayName: "statistics sensor",
		Short:       "Manage statistics sensor helpers",
		Long:        "Create, list, and delete statistics sensor helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new statistics sensor",
		CreateLong: `Create a new statistics sensor helper that provides statistical analysis of sensor history.

State characteristics: mean, median, standard_deviation, variance, sum, min, max, count,
                       datetime_newest, datetime_oldest, change, change_second,
                       average_linear, average_step, average_timeless, total,
                       change_sample, count_on, count_off, percentile, noisiness

At least one of --sampling-size or --max-age must be specified.

Examples:
  hab helper-statistics create "Temp Average" --entity sensor.temperature --characteristic mean --sampling-size 100
  hab helper-statistics create "Temp Std Dev" --entity sensor.temp --characteristic standard_deviation --max-age "01:00:00"
  hab helper-statistics create "Temp 95th" --entity sensor.temp --characteristic percentile --percentile 95 --sampling-size 50`,
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&entity, "entity", "e", "", "Source entity ID (required)")
			cmd.Flags().StringVarP(&characteristic, "characteristic", "c", "mean", "Statistical characteristic to calculate")
			cmd.Flags().IntVar(&samplingSize, "sampling-size", 0, "Maximum number of samples to store")
			cmd.Flags().StringVar(&maxAge, "max-age", "", "Maximum age of samples (e.g., 01:00:00)")
			cmd.Flags().IntVar(&precision, "precision", 2, "Decimal precision for results")
			cmd.Flags().IntVar(&percentile, "percentile", 50, "Percentile value (1-99, for percentile characteristic)")
		},
		RequiredFlags: []string{"entity"},
		RunCreate: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			textMode := getTextMode()

			// Validate
			validCharacteristics := map[string]bool{
				"mean": true, "median": true, "standard_deviation": true, "variance": true,
				"sum": true, "min": true, "max": true, "count": true,
				"datetime_newest": true, "datetime_oldest": true,
				"change": true, "change_second": true, "change_sample": true,
				"average_linear": true, "average_step": true, "average_timeless": true,
				"total": true, "count_on": true, "count_off": true,
				"percentile": true, "noisiness": true,
			}
			if !validCharacteristics[characteristic] {
				return fmt.Errorf("invalid characteristic: %s", characteristic)
			}

			hasSamplingSize := cmd.Flags().Changed("sampling-size") && samplingSize > 0
			hasMaxAge := cmd.Flags().Changed("max-age") && maxAge != ""
			if !hasSamplingSize && !hasMaxAge {
				return fmt.Errorf("at least one of --sampling-size or --max-age must be specified")
			}

			if characteristic == "percentile" {
				if percentile < 1 || percentile > 99 {
					return fmt.Errorf("percentile must be between 1 and 99")
				}
			}

			rest, err := getRESTClient()
			if err != nil {
				return err
			}

			flowResult, err := rest.ConfigFlowCreate("statistics")
			if err != nil {
				return fmt.Errorf("failed to start config flow: %w", err)
			}
			flowID, ok := flowResult["flow_id"].(string)
			if !ok {
				return fmt.Errorf("no flow_id in response")
			}

			// Step 1: entity selection
			step1Result, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
				"name":      name,
				"entity_id": entity,
			})
			if err != nil {
				return fmt.Errorf("failed to submit entity selection: %w", err)
			}
			if t, _ := step1Result["type"].(string); t == "abort" {
				reason, _ := step1Result["reason"].(string)
				return fmt.Errorf("config flow aborted: %s", reason)
			}

			// Step 2: state characteristic
			step2Result, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
				"state_characteristic": characteristic,
			})
			if err != nil {
				return fmt.Errorf("failed to submit state characteristic: %w", err)
			}
			if t, _ := step2Result["type"].(string); t == "abort" {
				reason, _ := step2Result["reason"].(string)
				return fmt.Errorf("config flow aborted: %s", reason)
			}

			// Step 3: options
			step3Data := map[string]interface{}{
				"precision":        precision,
				"keep_last_sample": false,
			}
			if hasSamplingSize {
				step3Data["sampling_size"] = samplingSize
			}
			if hasMaxAge {
				step3Data["max_age"] = parseDuration(maxAge)
			}
			if characteristic == "percentile" {
				step3Data["percentile"] = percentile
			}

			finalResult, err := rest.ConfigFlowStep(flowID, step3Data)
			if err != nil {
				return fmt.Errorf("failed to create statistics sensor: %w", err)
			}

			resultType, _ := finalResult["type"].(string)
			if resultType == "abort" {
				reason, _ := finalResult["reason"].(string)
				return fmt.Errorf("config flow aborted: %s", reason)
			}
			if resultType == "form" {
				if errors, ok := finalResult["errors"].(map[string]interface{}); ok && len(errors) > 0 {
					return fmt.Errorf("validation error: %v", errors)
				}
				return fmt.Errorf("unexpected form step required: %v", finalResult)
			}
			if resultType != "create_entry" {
				return fmt.Errorf("unexpected flow result type: %s", resultType)
			}

			result := map[string]interface{}{
				"title":          finalResult["title"],
				"entity":         entity,
				"characteristic": characteristic,
			}
			if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
				if entryID, ok := entryResult["entry_id"]; ok {
					result["entry_id"] = entryID
				}
			}

			output.PrintSuccess(result, textMode, fmt.Sprintf("Statistics sensor '%s' created successfully.", name))
			return nil
		},
	})
}

func registerLocalCalendar() {
	var icon string

	registerHelperType(HelperDef{
		TypeName:    "local_calendar",
		CommandName: "local-calendar",
		DisplayName: "local calendar",
		Short:       "Manage local calendar helpers",
		Long:        "Create, list, and delete local calendar helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new local calendar",
		CreateLong:  "Create a new local calendar helper.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the calendar")
		},
		RunCreate: helperConfigFlowCreate("local_calendar", "local calendar", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			formData := map[string]interface{}{
				"calendar_name": name,
			}
			if icon != "" {
				formData["icon"] = icon
			}
			return formData, nil
		}),
	})
}

func registerLocalTodo() {
	var icon string

	registerHelperType(HelperDef{
		TypeName:    "local_todo",
		CommandName: "local-todo",
		DisplayName: "local to-do list",
		Short:       "Manage local to-do list helpers",
		Long:        "Create, list, and delete local to-do list helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new local to-do list",
		CreateLong:  "Create a new local to-do list helper.",
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&icon, "icon", "i", "", "Icon for the to-do list")
		},
		RunCreate: helperConfigFlowCreate("local_todo", "local to-do list", func(cmd *cobra.Command, name string) (map[string]interface{}, error) {
			formData := map[string]interface{}{
				"todo_list_name": name,
			}
			if icon != "" {
				formData["icon"] = icon
			}
			return formData, nil
		}),
	})
}

func registerGroup() {
	var groupType, sensorType string
	var entities []string
	var all, hideMembers bool

	registerHelperType(HelperDef{
		TypeName:    "group",
		CommandName: "group",
		DisplayName: "group",
		Short:       "Manage group helpers",
		Long:        "Create, list, and delete group helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new group",
		CreateLong: `Create a new group helper using the config entry flow.

Group types available: binary_sensor, cover, event, fan, light, lock, media_player, sensor, switch.

For sensor groups, use --sensor-type to specify aggregation: last, max, mean, median, min, product, range, stdev, sum.

Examples:
  hab helper-group create "Living Room Lights" --type light --entities light.lamp1,light.lamp2
  hab helper-group create "All Motion Sensors" --type binary_sensor --entities binary_sensor.motion1,binary_sensor.motion2 --all
  hab helper-group create "Average Temperature" --type sensor --sensor-type mean --entities sensor.temp1,sensor.temp2`,
		SetupFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVarP(&groupType, "type", "t", "light", "Group type: binary_sensor, cover, event, fan, light, lock, media_player, sensor, switch")
			cmd.Flags().StringSliceVarP(&entities, "entities", "e", nil, "Entity IDs to include in the group (required)")
			cmd.Flags().BoolVar(&all, "all", false, "Set to true if all entities must be on for group to be on (only for binary_sensor, light, switch)")
			cmd.Flags().BoolVar(&hideMembers, "hide-members", false, "Hide member entities from the UI")
			cmd.Flags().StringVar(&sensorType, "sensor-type", "mean", "Sensor aggregation type: last, max, mean, median, min, product, range, stdev, sum")
		},
		RequiredFlags: []string{"entities"},
		RunCreate: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			textMode := getTextMode()

			validTypes := map[string]bool{
				"binary_sensor": true, "cover": true, "event": true, "fan": true,
				"light": true, "lock": true, "media_player": true, "sensor": true, "switch": true,
			}
			if !validTypes[groupType] {
				return fmt.Errorf("invalid group type: %s. Valid types: binary_sensor, cover, event, fan, light, lock, media_player, sensor, switch", groupType)
			}

			if groupType == "sensor" {
				validSensorTypes := map[string]bool{
					"last": true, "max": true, "mean": true, "median": true,
					"min": true, "product": true, "range": true, "stdev": true, "sum": true,
				}
				if !validSensorTypes[sensorType] {
					return fmt.Errorf("invalid sensor type: %s. Valid types: last, max, mean, median, min, product, range, stdev, sum", sensorType)
				}
			}

			rest, err := getRESTClient()
			if err != nil {
				return err
			}

			flowResult, err := rest.ConfigFlowCreate("group")
			if err != nil {
				return fmt.Errorf("failed to start config flow: %w", err)
			}
			flowID, ok := flowResult["flow_id"].(string)
			if !ok {
				return fmt.Errorf("no flow_id in response")
			}

			// Menu step: select group type
			menuResult, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
				"next_step_id": groupType,
			})
			if err != nil {
				return fmt.Errorf("failed to select group type: %w", err)
			}
			if t, _ := menuResult["type"].(string); t == "abort" {
				reason, _ := menuResult["reason"].(string)
				return fmt.Errorf("config flow aborted: %s", reason)
			}

			formData := map[string]interface{}{
				"name":         name,
				"entities":     entities,
				"hide_members": hideMembers,
			}
			if groupType == "sensor" {
				formData["type"] = sensorType
			} else if groupType == "binary_sensor" || groupType == "light" || groupType == "switch" {
				formData["all"] = all
			}

			finalResult, err := rest.ConfigFlowStep(flowID, formData)
			if err != nil {
				return fmt.Errorf("failed to create group: %w", err)
			}

			resultType, _ := finalResult["type"].(string)
			if resultType == "abort" {
				reason, _ := finalResult["reason"].(string)
				return fmt.Errorf("config flow aborted: %s", reason)
			}
			if resultType != "create_entry" {
				return fmt.Errorf("unexpected flow result type: %s", resultType)
			}

			result := map[string]interface{}{
				"title":    finalResult["title"],
				"type":     groupType,
				"entities": entities,
			}
			if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
				if entryID, ok := entryResult["entry_id"]; ok {
					result["entry_id"] = entryID
				}
			}

			output.PrintSuccess(result, textMode, fmt.Sprintf("Group '%s' created successfully.", name))
			return nil
		},
	})
}

func registerTemplate() {
	registerHelperType(HelperDef{
		TypeName:    "template",
		CommandName: "template",
		DisplayName: "template entity",
		Short:       "Manage template entity helpers",
		Long:        "Create, list, and delete template entity helpers.",
		Category:    HelperCategoryConfigFlow,
		CreateShort: "Create a new template entity",
		CreateLong: `Create a new template entity helper using the config entry flow.

Template types available: alarm_control_panel, binary_sensor, button, image, number, select, sensor, switch.

Templates use Jinja2 syntax. State templates should return valid values for the entity type.

Examples:
  hab helper-template create "Is Sun Up" --type binary_sensor --state "{{ is_state('sun.sun', 'above_horizon') }}"
  hab helper-template create "Room Temperature" --type sensor --state "{{ states('sensor.temp1') | float + states('sensor.temp2') | float }}" --unit "°C"
  hab helper-template create "All Lights" --type switch --state "{{ is_state('light.living_room', 'on') }}" --turn-on "homeassistant.turn_on" --turn-off "homeassistant.turn_off"`,
		SetupFlags: setupTemplateCreateFlags,
		RunCreate:  runTemplateCreate,
	})
}
