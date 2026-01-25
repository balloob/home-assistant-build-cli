package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperDerivativeCreateSource     string
	helperDerivativeCreateRound      int
	helperDerivativeCreateUnitPrefix string
	helperDerivativeCreateUnitTime   string
	helperDerivativeCreateTimeWindow string
)

var helperDerivativeCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new derivative sensor",
	Long: `Create a new derivative sensor helper that calculates the rate of change of a source sensor.

The derivative sensor tracks how fast a sensor value is changing over time.

Unit prefixes: n (nano), µ (micro), m (milli), k (kilo), M (mega), G (giga), T (tera)
Time units: s (seconds), min (minutes), h (hours), d (days)

Examples:
  hab helper-derivative create "Power Rate" --source sensor.energy_usage
  hab helper-derivative create "Temperature Change" --source sensor.temperature --unit-time min --round 2
  hab helper-derivative create "Smooth Power Rate" --source sensor.power --time-window 00:05:00`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperDerivativeCreate,
}

func init() {
	helperDerivativeParentCmd.AddCommand(helperDerivativeCreateCmd)
	helperDerivativeCreateCmd.Flags().StringVarP(&helperDerivativeCreateSource, "source", "s", "", "Source entity ID (required)")
	helperDerivativeCreateCmd.Flags().IntVar(&helperDerivativeCreateRound, "round", 2, "Decimal places for rounding (0-6)")
	helperDerivativeCreateCmd.Flags().StringVar(&helperDerivativeCreateUnitPrefix, "unit-prefix", "", "Metric unit prefix (n, µ, m, k, M, G, T, P)")
	helperDerivativeCreateCmd.Flags().StringVar(&helperDerivativeCreateUnitTime, "unit-time", "h", "Time unit for derivative (s, min, h, d)")
	helperDerivativeCreateCmd.Flags().StringVar(&helperDerivativeCreateTimeWindow, "time-window", "00:00:00", "Time window for smoothing (HH:MM:SS format)")
	helperDerivativeCreateCmd.MarkFlagRequired("source")
}

func runHelperDerivativeCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Validate time unit
	validTimeUnits := map[string]bool{"s": true, "min": true, "h": true, "d": true}
	if !validTimeUnits[helperDerivativeCreateUnitTime] {
		return fmt.Errorf("invalid time unit: %s. Valid units: s, min, h, d", helperDerivativeCreateUnitTime)
	}

	// Validate unit prefix if provided
	if helperDerivativeCreateUnitPrefix != "" {
		validPrefixes := map[string]bool{"n": true, "µ": true, "m": true, "k": true, "M": true, "G": true, "T": true, "P": true}
		if !validPrefixes[helperDerivativeCreateUnitPrefix] {
			return fmt.Errorf("invalid unit prefix: %s. Valid prefixes: n, µ, m, k, M, G, T, P", helperDerivativeCreateUnitPrefix)
		}
	}

	// Parse time window (HH:MM:SS format)
	timeWindow := parseDuration(helperDerivativeCreateTimeWindow)

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// Start the config flow for derivative
	flowResult, err := rest.ConfigFlowCreate("derivative")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}

	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Submit the form data
	formData := map[string]interface{}{
		"name":        name,
		"source":      helperDerivativeCreateSource,
		"round":       helperDerivativeCreateRound,
		"unit_time":   helperDerivativeCreateUnitTime,
		"time_window": timeWindow,
	}

	if helperDerivativeCreateUnitPrefix != "" {
		formData["unit_prefix"] = helperDerivativeCreateUnitPrefix
	}

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create derivative sensor: %w", err)
	}

	resultType, _ := finalResult["type"].(string)
	if resultType == "abort" {
		reason, _ := finalResult["reason"].(string)
		return fmt.Errorf("config flow aborted: %s", reason)
	}

	if resultType == "form" {
		// Check for errors in the form response
		if errors, ok := finalResult["errors"].(map[string]interface{}); ok && len(errors) > 0 {
			return fmt.Errorf("validation error: %v", errors)
		}
		return fmt.Errorf("unexpected form step required: %v", finalResult)
	}

	if resultType != "create_entry" {
		return fmt.Errorf("unexpected flow result type: %s", resultType)
	}

	result := map[string]interface{}{
		"title":  finalResult["title"],
		"source": helperDerivativeCreateSource,
	}
	if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
		if entryID, ok := entryResult["entry_id"]; ok {
			result["entry_id"] = entryID
		}
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Derivative sensor '%s' created successfully.", name))
	return nil
}

// parseDuration converts HH:MM:SS format to a duration object
func parseDuration(s string) map[string]interface{} {
	hours, minutes, seconds := 0, 0, 0
	fmt.Sscanf(s, "%d:%d:%d", &hours, &minutes, &seconds)
	return map[string]interface{}{
		"hours":   hours,
		"minutes": minutes,
		"seconds": seconds,
	}
}
