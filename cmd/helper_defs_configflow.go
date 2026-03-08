package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

// ========== Config flow helpers ==========

func registerDerivative() {
	var source, unitPrefix, unitTime, timeWindow string
	var round int

	registerHelperType(HelperDef{
		TypeName:        "derivative",
		CommandName:     "derivative",
		DisplayName:     "derivative sensor",
		Short:           "Manage derivative sensor helpers",
		Long:            "Create, list, and delete derivative sensor helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "Calculates the rate of change of a source sensor (config flow)",
		CreateParams:    []string{"name (required)", "source (required)", "round", "unit-prefix (n/µ/m/k/M/G/T)", "unit-time (s/min/h/d)", "time-window"},
		CreateShort:     "Create a new derivative sensor",
		CreateLong: `Create a new derivative sensor helper that calculates the rate of change of a source sensor.

Unit prefixes: n (nano), µ (micro), m (milli), k (kilo), M (mega), G (giga), T (tera)
Time units: s (seconds), min (minutes), h (hours), d (days)`,
		CreateExample: `  hab helper-derivative create "Power Rate" --source sensor.energy_usage
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
			if err := validateOneOf(unitTime, "time unit", validTimeUnits, false); err != nil {
				return nil, err
			}
			if err := validateOneOf(unitPrefix, "unit prefix", validMetricPrefixes, true); err != nil {
				return nil, err
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
		TypeName:        "integration",
		CommandName:     "integration",
		DisplayName:     "integration sensor",
		Short:           "Manage integration (integral) sensor helpers",
		Long:            "Create, list, and delete integration (Riemann sum integral) sensor helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "Calculates the Riemann sum (integral) of a source sensor (config flow)",
		CreateParams:    []string{"name (required)", "source (required)", "round", "unit-prefix (k/M/G/T)", "unit-time (s/min/h/d)", "method (trapezoidal/left/right)"},
		CreateShort:     "Create a new integration sensor",
		CreateLong: `Create a new integration (Riemann sum integral) sensor helper.`,
		CreateExample: `  hab helper-integration create "Total Energy" --source sensor.power
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
			if err := validateOneOf(unitTime, "time unit", validTimeUnits, false); err != nil {
				return nil, err
			}
			validMethods := map[string]bool{"left": true, "right": true, "trapezoidal": true}
			if err := validateOneOf(method, "method", validMethods, false); err != nil {
				return nil, err
			}
			if err := validateOneOf(unitPrefix, "unit prefix", validMetricPrefixes, true); err != nil {
				return nil, err
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
		TypeName:        "min_max",
		CommandName:     "min-max",
		DisplayName:     "min/max sensor",
		Short:           "Manage min/max sensor helpers",
		Long:            "Create, list, and delete min/max sensor helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "Aggregates values from multiple sensors (min/max/mean/etc) (config flow)",
		CreateParams:    []string{"name (required)", "entities (required, array)", "type (min/max/mean/median/last/range/sum)", "round"},
		CreateShort:     "Create a new min/max sensor",
		CreateLong: `Create a new min/max sensor helper.`,
		CreateExample: `  hab helper-min-max create "Highest Temp" --type max --entities sensor.temp1,sensor.temp2`,
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
		TypeName:        "threshold",
		CommandName:     "threshold",
		DisplayName:     "threshold sensor",
		Short:           "Manage threshold sensor helpers",
		Long:            "Create, list, and delete threshold binary sensor helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "Monitors a sensor value against configurable thresholds (config flow)",
		CreateParams:    []string{"name (required)", "entity (required)", "lower", "upper", "hysteresis"},
		CreateShort:     "Create a new threshold sensor",
		CreateLong: `Create a new threshold binary sensor helper. At least one of --lower or --upper must be specified.`,
		CreateExample: `  hab helper-threshold create "Freezing Alert" --entity sensor.temperature --lower 0
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
		TypeName:        "utility_meter",
		CommandName:     "utility-meter",
		DisplayName:     "utility meter",
		Short:           "Manage utility meter helpers",
		Long:            "Create, list, and delete utility meter helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "Tracks consumption across billing cycles (config flow)",
		CreateParams:    []string{"name (required)", "source (required)", "cycle (quarter-hourly/hourly/daily/weekly/monthly/bimonthly/quarterly/yearly)", "offset", "tariffs (array)", "delta-values", "net-consumption"},
		CreateShort:     "Create a new utility meter",
		CreateLong: `Create a new utility meter helper that tracks consumption across billing cycles.

Cycle options: quarter-hourly, hourly, daily, weekly, monthly, bimonthly, quarterly, yearly`,
		CreateExample: `  hab helper-utility-meter create "Monthly Energy" --source sensor.total_energy --cycle monthly
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
		TypeName:        "statistics",
		CommandName:     "statistics",
		DisplayName:     "statistics sensor",
		Short:           "Manage statistics sensor helpers",
		Long:            "Create, list, and delete statistics sensor helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "Provides statistical analysis of sensor history (config flow)",
		CreateParams:    []string{"name (required)", "entity (required)", "characteristic (mean/median/standard_deviation/etc)", "sampling-size", "max-age", "precision", "percentile"},
		CreateShort:     "Create a new statistics sensor",
		CreateLong: `Create a new statistics sensor helper that provides statistical analysis of sensor history.

State characteristics: mean, median, standard_deviation, variance, sum, min, max, count,
                       datetime_newest, datetime_oldest, change, change_second,
                       average_linear, average_step, average_timeless, total,
                       change_sample, count_on, count_off, percentile, noisiness

At least one of --sampling-size or --max-age must be specified.`,
		CreateExample: `  hab helper-statistics create "Temp Average" --entity sensor.temperature --characteristic mean --sampling-size 100
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
			if err := validateOneOf(characteristic, "characteristic", validCharacteristics, false); err != nil {
				return err
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

			flowID, err := configFlowStart(rest, "statistics")
			if err != nil {
				return err
			}

			// Step 1: entity selection
			step1Result, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
				"name":      name,
				"entity_id": entity,
			})
			if err != nil {
				return fmt.Errorf("failed to submit entity selection: %w", err)
			}
			if err := configFlowCheckAbort(step1Result); err != nil {
				return err
			}

			// Step 2: state characteristic
			step2Result, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
				"state_characteristic": characteristic,
			})
			if err != nil {
				return fmt.Errorf("failed to submit state characteristic: %w", err)
			}
			if err := configFlowCheckAbort(step2Result); err != nil {
				return err
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

			result, err := configFlowCheckFinal(finalResult)
			if err != nil {
				return err
			}
			result["entity"] = entity
			result["characteristic"] = characteristic

			output.PrintSuccess(result, textMode, fmt.Sprintf("Statistics sensor '%s' created successfully.", name))
			return nil
		},
	})
}

func registerLocalCalendar() {
	var icon string

	registerHelperType(HelperDef{
		TypeName:        "local_calendar",
		CommandName:     "local-calendar",
		DisplayName:     "local calendar",
		Short:           "Manage local calendar helpers",
		Long:            "Create, list, and delete local calendar helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "A local calendar for storing events in Home Assistant (config flow)",
		CreateParams:    []string{"name (required)", "icon"},
		CreateShort:     "Create a new local calendar",
		CreateLong:      "Create a new local calendar helper.",
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
		TypeName:        "local_todo",
		CommandName:     "local-todo",
		DisplayName:     "local to-do list",
		Short:           "Manage local to-do list helpers",
		Long:            "Create, list, and delete local to-do list helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "A local to-do list for storing tasks in Home Assistant (config flow)",
		CreateParams:    []string{"name (required)", "icon"},
		CreateShort:     "Create a new local to-do list",
		CreateLong:      "Create a new local to-do list helper.",
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
		TypeName:        "group",
		CommandName:     "group",
		DisplayName:     "group",
		Short:           "Manage group helpers",
		Long:            "Create, list, and delete group helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "A group of entities that can be controlled together (config flow)",
		CreateParams:    []string{"name (required)", "type (light/switch/binary_sensor/cover/fan/lock/media_player/sensor/event)", "entities (required, array)", "all (true/false, for binary_sensor/light/switch)", "hide-members (true/false)"},
		CreateShort:     "Create a new group",
		CreateLong: `Create a new group helper using the config entry flow.

Group types available: binary_sensor, cover, event, fan, light, lock, media_player, sensor, switch.

For sensor groups, use --sensor-type to specify aggregation: last, max, mean, median, min, product, range, stdev, sum.`,
		CreateExample: `  hab helper-group create "Living Room Lights" --type light --entities light.lamp1,light.lamp2
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
			if err := validateOneOf(groupType, "group type", validTypes, false); err != nil {
				return err
			}

			if groupType == "sensor" {
				validSensorTypes := map[string]bool{
					"last": true, "max": true, "mean": true, "median": true,
					"min": true, "product": true, "range": true, "stdev": true, "sum": true,
				}
				if err := validateOneOf(sensorType, "sensor type", validSensorTypes, false); err != nil {
					return err
				}
			}

			rest, err := getRESTClient()
			if err != nil {
				return err
			}

			flowID, err := configFlowStart(rest, "group")
			if err != nil {
				return err
			}

			// Menu step: select group type
			menuResult, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
				"next_step_id": groupType,
			})
			if err != nil {
				return fmt.Errorf("failed to select group type: %w", err)
			}
			if err := configFlowCheckAbort(menuResult); err != nil {
				return err
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

			result, err := configFlowCheckFinal(finalResult)
			if err != nil {
				return err
			}
			result["type"] = groupType
			result["entities"] = entities

			output.PrintSuccess(result, textMode, fmt.Sprintf("Group '%s' created successfully.", name))
			return nil
		},
	})
}

func registerTemplate() {
	registerHelperType(HelperDef{
		TypeName:        "template",
		CommandName:     "template",
		DisplayName:     "template entity",
		Short:           "Manage template entity helpers",
		Long:            "Create, list, and delete template entity helpers.",
		Category:        HelperCategoryConfigFlow,
		TypeDescription: "Create template entities using Jinja2 expressions (config flow)",
		CreateParams:    []string{"name (required)", "type (alarm_control_panel/binary_sensor/button/image/number/select/sensor/switch)", "state (Jinja2 template)", "icon", "turn-on", "turn-off"},
		CreateShort:     "Create a new template entity",
		CreateLong: `Create a new template entity helper using the config entry flow.

Template types available: alarm_control_panel, binary_sensor, button, image, number, select, sensor, switch.

Templates use Jinja2 syntax. State templates should return valid values for the entity type.`,
		CreateExample: `  hab helper-template create "Is Sun Up" --type binary_sensor --state "{{ is_state('sun.sun', 'above_horizon') }}"
  hab helper-template create "Room Temperature" --type sensor --state "{{ states('sensor.temp1') | float + states('sensor.temp2') | float }}" --unit "°C"
  hab helper-template create "All Lights" --type switch --state "{{ is_state('light.living_room', 'on') }}" --turn-on "homeassistant.turn_on" --turn-off "homeassistant.turn_off"`,
		SetupFlags: setupTemplateCreateFlags,
		RunCreate:  runTemplateCreate,
	})
}
