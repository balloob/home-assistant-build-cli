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
		{
			Type:        "group",
			Description: "A group of entities that can be controlled together (uses config flow)",
			Parameters:  []string{"name (required)", "type (light/switch/binary_sensor/cover/fan/lock/media_player/sensor/event)", "entities (required, array)", "all (true/false, for binary_sensor/light/switch)", "hide-members (true/false)"},
		},
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
