package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List available helper types",
	Long:  `List all available helper types that can be created.`,
	RunE:  runHelperTypes,
}

func init() {
	helperCmd.AddCommand(helperTypesCmd)
}

// HelperTypeInfo contains information about a helper type
type HelperTypeInfo struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Parameters  []string `json:"parameters"`
}

func runHelperTypes(cmd *cobra.Command, args []string) error {
	textMode := viper.GetBool("text")

	types := []HelperTypeInfo{
		// Storage-based helpers (WebSocket API)
		{
			Type:        "input_boolean",
			Description: "A boolean on/off toggle helper",
			Parameters:  []string{"name (required)", "icon", "initial (true/false)"},
		},
		{
			Type:        "input_number",
			Description: "A numeric value helper with min/max range",
			Parameters:  []string{"name (required)", "min (required)", "max (required)", "icon", "initial", "step", "mode (box/slider)", "unit_of_measurement"},
		},
		{
			Type:        "input_text",
			Description: "A text input helper",
			Parameters:  []string{"name (required)", "icon", "initial", "min", "max", "pattern", "mode (text/password)"},
		},
		{
			Type:        "input_select",
			Description: "A dropdown selection helper",
			Parameters:  []string{"name (required)", "options (required, array)", "icon", "initial"},
		},
		{
			Type:        "input_datetime",
			Description: "A date/time helper",
			Parameters:  []string{"name (required)", "has_date (required)", "has_time (required)", "icon", "initial"},
		},
		{
			Type:        "input_button",
			Description: "A button helper that can be pressed",
			Parameters:  []string{"name (required)", "icon"},
		},
		{
			Type:        "counter",
			Description: "A counter helper that can be incremented/decremented",
			Parameters:  []string{"name (required)", "icon", "initial", "minimum", "maximum", "step", "restore (true/false)"},
		},
		{
			Type:        "timer",
			Description: "A timer helper that counts down",
			Parameters:  []string{"name (required)", "icon", "duration", "restore (true/false)"},
		},
		{
			Type:        "schedule",
			Description: "A schedule helper for time-based automation",
			Parameters:  []string{"name (required)", "icon"},
		},
		// Config flow-based helpers (REST API)
		{
			Type:        "group",
			Description: "A group of entities that can be controlled together (config flow)",
			Parameters:  []string{"name (required)", "type (light/switch/binary_sensor/cover/fan/lock/media_player/sensor/event)", "entities (required, array)", "all (true/false, for binary_sensor/light/switch)", "hide-members (true/false)"},
		},
		{
			Type:        "derivative",
			Description: "Calculates the rate of change of a source sensor (config flow)",
			Parameters:  []string{"name (required)", "source (required)", "round", "unit-prefix (n/µ/m/k/M/G/T)", "unit-time (s/min/h/d)", "unit", "time-window"},
		},
		{
			Type:        "integration",
			Description: "Calculates the Riemann sum (integral) of a source sensor (config flow)",
			Parameters:  []string{"name (required)", "source (required)", "round", "unit-prefix (k/M/G/T)", "unit-time (s/min/h/d)", "method (trapezoidal/left/right)"},
		},
		{
			Type:        "min_max",
			Description: "Aggregates values from multiple sensors (min/max/mean/etc) (config flow)",
			Parameters:  []string{"name (required)", "entities (required, array)", "type (min/max/mean/median/last/range/sum)", "round"},
		},
		{
			Type:        "threshold",
			Description: "Monitors a sensor value against configurable thresholds (config flow)",
			Parameters:  []string{"name (required)", "entity (required)", "lower", "upper", "hysteresis"},
		},
		{
			Type:        "utility_meter",
			Description: "Tracks consumption across billing cycles (config flow)",
			Parameters:  []string{"name (required)", "source (required)", "cycle (quarter-hourly/hourly/daily/weekly/monthly/bimonthly/quarterly/yearly)", "offset", "tariffs (array)", "delta-values", "net-consumption"},
		},
		{
			Type:        "statistics",
			Description: "Provides statistical analysis of sensor history (config flow)",
			Parameters:  []string{"name (required)", "entity (required)", "characteristic (mean/median/standard_deviation/etc)", "sampling-size", "max-age", "precision", "percentile"},
		},
		{
			Type:        "local_calendar",
			Description: "A local calendar for storing events in Home Assistant (config flow)",
			Parameters:  []string{"name (required)", "icon"},
		},
		{
			Type:        "local_todo",
			Description: "A local to-do list for storing tasks in Home Assistant (config flow)",
			Parameters:  []string{"name (required)", "icon"},
		},
	}

	// Convert to interface slice for output
	result := make([]interface{}, len(types))
	for i, t := range types {
		result[i] = map[string]interface{}{
			"type":        t.Type,
			"description": t.Description,
			"parameters":  t.Parameters,
		}
	}

	client.PrintOutput(result, textMode, "")
	return nil
}
