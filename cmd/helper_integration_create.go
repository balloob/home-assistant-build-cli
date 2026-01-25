package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperIntegrationCreateSource     string
	helperIntegrationCreateRound      int
	helperIntegrationCreateUnitPrefix string
	helperIntegrationCreateUnitTime   string
	helperIntegrationCreateMethod     string
)

var helperIntegrationCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new integration (integral) sensor",
	Long: `Create a new integration sensor helper that calculates the Riemann sum (integral) of a source sensor.

The integration sensor tracks the accumulated value over time (e.g., power to energy).

Unit prefixes: k (kilo), M (mega), G (giga), T (tera)
Time units: s (seconds), min (minutes), h (hours), d (days)
Methods: trapezoidal (default), left, right

Examples:
  hab helper-integration create "Total Energy" --source sensor.power_usage
  hab helper-integration create "Energy kWh" --source sensor.power --unit-prefix k --unit-time h
  hab helper-integration create "Energy Left" --source sensor.power --method left`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperIntegrationCreate,
}

func init() {
	helperIntegrationParentCmd.AddCommand(helperIntegrationCreateCmd)
	helperIntegrationCreateCmd.Flags().StringVarP(&helperIntegrationCreateSource, "source", "s", "", "Source entity ID (required)")
	helperIntegrationCreateCmd.Flags().IntVar(&helperIntegrationCreateRound, "round", 3, "Decimal places for rounding")
	helperIntegrationCreateCmd.Flags().StringVar(&helperIntegrationCreateUnitPrefix, "unit-prefix", "", "Metric unit prefix (k, M, G, T)")
	helperIntegrationCreateCmd.Flags().StringVar(&helperIntegrationCreateUnitTime, "unit-time", "h", "Time unit for integration (s, min, h, d)")
	helperIntegrationCreateCmd.Flags().StringVar(&helperIntegrationCreateMethod, "method", "trapezoidal", "Integration method (trapezoidal, left, right)")
	helperIntegrationCreateCmd.MarkFlagRequired("source")
}

func runHelperIntegrationCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Validate time unit
	validTimeUnits := map[string]bool{"s": true, "min": true, "h": true, "d": true}
	if !validTimeUnits[helperIntegrationCreateUnitTime] {
		return fmt.Errorf("invalid time unit: %s. Valid units: s, min, h, d", helperIntegrationCreateUnitTime)
	}

	// Validate method
	validMethods := map[string]bool{"trapezoidal": true, "left": true, "right": true}
	if !validMethods[helperIntegrationCreateMethod] {
		return fmt.Errorf("invalid method: %s. Valid methods: trapezoidal, left, right", helperIntegrationCreateMethod)
	}

	// Validate unit prefix if provided
	if helperIntegrationCreateUnitPrefix != "" {
		validPrefixes := map[string]bool{"k": true, "M": true, "G": true, "T": true}
		if !validPrefixes[helperIntegrationCreateUnitPrefix] {
			return fmt.Errorf("invalid unit prefix: %s. Valid prefixes: k, M, G, T", helperIntegrationCreateUnitPrefix)
		}
	}

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// Start the config flow for integration
	flowResult, err := rest.ConfigFlowCreate("integration")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}

	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Submit the form data
	formData := map[string]interface{}{
		"name":      name,
		"source":    helperIntegrationCreateSource,
		"round":     helperIntegrationCreateRound,
		"unit_time": helperIntegrationCreateUnitTime,
		"method":    helperIntegrationCreateMethod,
	}

	if helperIntegrationCreateUnitPrefix != "" {
		formData["unit_prefix"] = helperIntegrationCreateUnitPrefix
	}

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create integration sensor: %w", err)
	}

	resultType, _ := finalResult["type"].(string)
	if resultType == "abort" {
		reason, _ := finalResult["reason"].(string)
		return fmt.Errorf("config flow aborted: %s", reason)
	}

	if resultType != "create_entry" {
		if resultType == "form" {
			return fmt.Errorf("unexpected form step required: %v", finalResult)
		}
		return fmt.Errorf("unexpected flow result type: %s", resultType)
	}

	result := map[string]interface{}{
		"title":  finalResult["title"],
		"source": helperIntegrationCreateSource,
		"method": helperIntegrationCreateMethod,
	}
	if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
		if entryID, ok := entryResult["entry_id"]; ok {
			result["entry_id"] = entryID
		}
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Integration sensor '%s' created successfully.", name))
	return nil
}
