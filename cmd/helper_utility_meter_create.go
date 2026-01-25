package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperUtilityMeterCreateSource              string
	helperUtilityMeterCreateCycle               string
	helperUtilityMeterCreateOffset              int
	helperUtilityMeterCreateTariffs             []string
	helperUtilityMeterCreateDeltaValues         bool
	helperUtilityMeterCreateNetConsumption      bool
	helperUtilityMeterCreatePeriodicallyReset   bool
	helperUtilityMeterCreateAlwaysAvailable     bool
)

var helperUtilityMeterCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new utility meter",
	Long: `Create a new utility meter helper that tracks consumption across billing cycles.

Cycle options: quarter-hourly, hourly, daily, weekly, monthly, bimonthly, quarterly, yearly

Examples:
  hab helper-utility-meter create "Monthly Energy" --source sensor.total_energy --cycle monthly
  hab helper-utility-meter create "Daily Water" --source sensor.water_meter --cycle daily --delta-values
  hab helper-utility-meter create "Electric Bill" --source sensor.power --cycle monthly --tariffs "peak,off-peak"`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperUtilityMeterCreate,
}

func init() {
	helperUtilityMeterParentCmd.AddCommand(helperUtilityMeterCreateCmd)
	helperUtilityMeterCreateCmd.Flags().StringVarP(&helperUtilityMeterCreateSource, "source", "s", "", "Source entity ID (required)")
	helperUtilityMeterCreateCmd.Flags().StringVarP(&helperUtilityMeterCreateCycle, "cycle", "c", "monthly", "Reset cycle: quarter-hourly, hourly, daily, weekly, monthly, bimonthly, quarterly, yearly")
	helperUtilityMeterCreateCmd.Flags().IntVar(&helperUtilityMeterCreateOffset, "offset", 0, "Offset in days for cycle reset")
	helperUtilityMeterCreateCmd.Flags().StringSliceVar(&helperUtilityMeterCreateTariffs, "tariffs", nil, "Tariff names for multi-rate billing")
	helperUtilityMeterCreateCmd.Flags().BoolVar(&helperUtilityMeterCreateDeltaValues, "delta-values", false, "Source provides delta values (incremental)")
	helperUtilityMeterCreateCmd.Flags().BoolVar(&helperUtilityMeterCreateNetConsumption, "net-consumption", false, "Net meter that can increase/decrease")
	helperUtilityMeterCreateCmd.Flags().BoolVar(&helperUtilityMeterCreatePeriodicallyReset, "periodically-resetting", true, "Source may reset to 0 independently")
	helperUtilityMeterCreateCmd.Flags().BoolVar(&helperUtilityMeterCreateAlwaysAvailable, "always-available", false, "Maintain last value when source unavailable")
	helperUtilityMeterCreateCmd.MarkFlagRequired("source")
}

func runHelperUtilityMeterCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Map user-friendly cycle names to internal values
	cycleMap := map[string]string{
		"none":           "none",
		"quarter-hourly": "quarter_hourly",
		"hourly":         "hourly",
		"daily":          "daily",
		"weekly":         "weekly",
		"monthly":        "monthly",
		"bimonthly":      "bimonthly",
		"quarterly":      "quarterly",
		"yearly":         "yearly",
	}
	meterType, ok := cycleMap[helperUtilityMeterCreateCycle]
	if !ok {
		return fmt.Errorf("invalid cycle: %s. Valid cycles: none, quarter-hourly, hourly, daily, weekly, monthly, bimonthly, quarterly, yearly", helperUtilityMeterCreateCycle)
	}

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// Start the config flow for utility_meter
	flowResult, err := rest.ConfigFlowCreate("utility_meter")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}

	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Submit the form data with correct field names
	tariffs := helperUtilityMeterCreateTariffs
	if tariffs == nil {
		tariffs = []string{}
	}
	formData := map[string]interface{}{
		"name":                         name,
		"source_sensor":                helperUtilityMeterCreateSource,
		"meter_type":                   meterType,
		"meter_offset":                 helperUtilityMeterCreateOffset,
		"meter_delta_values":           helperUtilityMeterCreateDeltaValues,
		"meter_net_consumption":        helperUtilityMeterCreateNetConsumption,
		"meter_periodically_resetting": helperUtilityMeterCreatePeriodicallyReset,
		"tariffs":                      tariffs,
		"sensor_always_available":      helperUtilityMeterCreateAlwaysAvailable,
	}

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create utility meter: %w", err)
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
		"source": helperUtilityMeterCreateSource,
		"cycle":  helperUtilityMeterCreateCycle,
	}
	if len(helperUtilityMeterCreateTariffs) > 0 {
		result["tariffs"] = helperUtilityMeterCreateTariffs
	}
	if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
		if entryID, ok := entryResult["entry_id"]; ok {
			result["entry_id"] = entryID
		}
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Utility meter '%s' created successfully.", name))
	return nil
}
