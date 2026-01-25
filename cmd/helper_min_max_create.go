package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperMinMaxCreateEntities    []string
	helperMinMaxCreateType        string
	helperMinMaxCreateRoundDigits int
)

var helperMinMaxCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new min/max sensor",
	Long: `Create a new min/max sensor helper that aggregates values from multiple source sensors.

Aggregation types: min, max, mean, median, last, range, sum

Examples:
  hab helper-min-max create "Average Temperature" --entities sensor.temp1,sensor.temp2,sensor.temp3 --type mean
  hab helper-min-max create "Max Power" --entities sensor.power1,sensor.power2 --type max
  hab helper-min-max create "Sum Energy" --entities sensor.energy1,sensor.energy2 --type sum --round 2`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperMinMaxCreate,
}

func init() {
	helperMinMaxParentCmd.AddCommand(helperMinMaxCreateCmd)
	helperMinMaxCreateCmd.Flags().StringSliceVarP(&helperMinMaxCreateEntities, "entities", "e", nil, "Entity IDs to aggregate (required)")
	helperMinMaxCreateCmd.Flags().StringVarP(&helperMinMaxCreateType, "type", "t", "max", "Aggregation type: min, max, mean, median, last, range, sum")
	helperMinMaxCreateCmd.Flags().IntVar(&helperMinMaxCreateRoundDigits, "round", 2, "Decimal places for rounding")
	helperMinMaxCreateCmd.MarkFlagRequired("entities")
}

func runHelperMinMaxCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Validate aggregation type
	validTypes := map[string]bool{
		"min": true, "max": true, "mean": true, "median": true,
		"last": true, "range": true, "sum": true,
	}
	if !validTypes[helperMinMaxCreateType] {
		return fmt.Errorf("invalid aggregation type: %s. Valid types: min, max, mean, median, last, range, sum", helperMinMaxCreateType)
	}

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// Start the config flow for min_max
	flowResult, err := rest.ConfigFlowCreate("min_max")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}

	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Submit the form data
	formData := map[string]interface{}{
		"name":         name,
		"entity_ids":   helperMinMaxCreateEntities,
		"type":         helperMinMaxCreateType,
		"round_digits": helperMinMaxCreateRoundDigits,
	}

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create min/max sensor: %w", err)
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
		"title":    finalResult["title"],
		"entities": helperMinMaxCreateEntities,
		"type":     helperMinMaxCreateType,
	}
	if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
		if entryID, ok := entryResult["entry_id"]; ok {
			result["entry_id"] = entryID
		}
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Min/max sensor '%s' created successfully.", name))
	return nil
}
