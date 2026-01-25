package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperThresholdCreateEntity     string
	helperThresholdCreateLower      float64
	helperThresholdCreateUpper      float64
	helperThresholdCreateHysteresis float64
)

var helperThresholdCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new threshold binary sensor",
	Long: `Create a new threshold binary sensor helper that monitors a sensor value against thresholds.

At least one of --lower or --upper must be specified.
The sensor will be "on" when the value is above the upper threshold or below the lower threshold.

Examples:
  hab helper-threshold create "High Temperature" --entity sensor.temperature --upper 30
  hab helper-threshold create "Low Battery" --entity sensor.battery --lower 20
  hab helper-threshold create "Temperature Range" --entity sensor.temp --lower 18 --upper 25 --hysteresis 1`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperThresholdCreate,
}

func init() {
	helperThresholdParentCmd.AddCommand(helperThresholdCreateCmd)
	helperThresholdCreateCmd.Flags().StringVarP(&helperThresholdCreateEntity, "entity", "e", "", "Source entity ID to monitor (required)")
	helperThresholdCreateCmd.Flags().Float64VarP(&helperThresholdCreateLower, "lower", "l", 0, "Lower threshold value")
	helperThresholdCreateCmd.Flags().Float64VarP(&helperThresholdCreateUpper, "upper", "u", 0, "Upper threshold value")
	helperThresholdCreateCmd.Flags().Float64Var(&helperThresholdCreateHysteresis, "hysteresis", 0, "Hysteresis to prevent rapid state changes")
	helperThresholdCreateCmd.MarkFlagRequired("entity")
}

func runHelperThresholdCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// At least one threshold must be set
	hasLower := cmd.Flags().Changed("lower")
	hasUpper := cmd.Flags().Changed("upper")
	if !hasLower && !hasUpper {
		return fmt.Errorf("at least one of --lower or --upper must be specified")
	}

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// Start the config flow for threshold
	flowResult, err := rest.ConfigFlowCreate("threshold")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}

	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Submit the form data
	formData := map[string]interface{}{
		"name":       name,
		"entity_id":  helperThresholdCreateEntity,
		"hysteresis": helperThresholdCreateHysteresis,
	}

	if hasLower {
		formData["lower"] = helperThresholdCreateLower
	}
	if hasUpper {
		formData["upper"] = helperThresholdCreateUpper
	}

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create threshold sensor: %w", err)
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
		"entity": helperThresholdCreateEntity,
	}
	if hasLower {
		result["lower"] = helperThresholdCreateLower
	}
	if hasUpper {
		result["upper"] = helperThresholdCreateUpper
	}
	if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
		if entryID, ok := entryResult["entry_id"]; ok {
			result["entry_id"] = entryID
		}
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Threshold sensor '%s' created successfully.", name))
	return nil
}
