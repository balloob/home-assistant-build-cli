package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperStatisticsCreateEntity          string
	helperStatisticsCreateCharacteristic  string
	helperStatisticsCreateSamplingSize    int
	helperStatisticsCreateMaxAge          string
	helperStatisticsCreatePrecision       int
	helperStatisticsCreatePercentile      int
)

var helperStatisticsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new statistics sensor",
	Long: `Create a new statistics sensor helper that provides statistical analysis of sensor history.

State characteristics: mean, median, standard_deviation, variance, sum, min, max, count,
                       datetime_newest, datetime_oldest, change, change_second,
                       average_linear, average_step, average_timeless, total,
                       change_sample, count_on, count_off, percentile, noisiness

At least one of --sampling-size or --max-age must be specified.

Examples:
  hab helper-statistics create "Temp Average" --entity sensor.temperature --characteristic mean --sampling-size 100
  hab helper-statistics create "Temp Std Dev" --entity sensor.temp --characteristic standard_deviation --max-age "01:00:00"
  hab helper-statistics create "Temp 95th" --entity sensor.temp --characteristic percentile --percentile 95 --sampling-size 50`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperStatisticsCreate,
}

func init() {
	helperStatisticsParentCmd.AddCommand(helperStatisticsCreateCmd)
	helperStatisticsCreateCmd.Flags().StringVarP(&helperStatisticsCreateEntity, "entity", "e", "", "Source entity ID (required)")
	helperStatisticsCreateCmd.Flags().StringVarP(&helperStatisticsCreateCharacteristic, "characteristic", "c", "mean", "Statistical characteristic to calculate")
	helperStatisticsCreateCmd.Flags().IntVar(&helperStatisticsCreateSamplingSize, "sampling-size", 0, "Maximum number of samples to store")
	helperStatisticsCreateCmd.Flags().StringVar(&helperStatisticsCreateMaxAge, "max-age", "", "Maximum age of samples (e.g., 01:00:00)")
	helperStatisticsCreateCmd.Flags().IntVar(&helperStatisticsCreatePrecision, "precision", 2, "Decimal precision for results")
	helperStatisticsCreateCmd.Flags().IntVar(&helperStatisticsCreatePercentile, "percentile", 50, "Percentile value (1-99, for percentile characteristic)")
	helperStatisticsCreateCmd.MarkFlagRequired("entity")
}

func runHelperStatisticsCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Validate characteristic
	validCharacteristics := map[string]bool{
		"mean": true, "median": true, "standard_deviation": true, "variance": true,
		"sum": true, "min": true, "max": true, "count": true,
		"datetime_newest": true, "datetime_oldest": true,
		"change": true, "change_second": true, "change_sample": true,
		"average_linear": true, "average_step": true, "average_timeless": true,
		"total": true, "count_on": true, "count_off": true,
		"percentile": true, "noisiness": true,
	}
	if !validCharacteristics[helperStatisticsCreateCharacteristic] {
		return fmt.Errorf("invalid characteristic: %s", helperStatisticsCreateCharacteristic)
	}

	// At least one of sampling_size or max_age must be set
	hasSamplingSize := cmd.Flags().Changed("sampling-size") && helperStatisticsCreateSamplingSize > 0
	hasMaxAge := cmd.Flags().Changed("max-age") && helperStatisticsCreateMaxAge != ""
	if !hasSamplingSize && !hasMaxAge {
		return fmt.Errorf("at least one of --sampling-size or --max-age must be specified")
	}

	// Validate percentile if using percentile characteristic
	if helperStatisticsCreateCharacteristic == "percentile" {
		if helperStatisticsCreatePercentile < 1 || helperStatisticsCreatePercentile > 99 {
			return fmt.Errorf("percentile must be between 1 and 99")
		}
	}

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// Start the config flow for statistics
	flowResult, err := rest.ConfigFlowCreate("statistics")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}

	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Step 1: Submit name and entity_id
	step1Data := map[string]interface{}{
		"name":      name,
		"entity_id": helperStatisticsCreateEntity,
	}

	step1Result, err := rest.ConfigFlowStep(flowID, step1Data)
	if err != nil {
		return fmt.Errorf("failed to submit entity selection: %w", err)
	}

	step1Type, _ := step1Result["type"].(string)
	if step1Type == "abort" {
		reason, _ := step1Result["reason"].(string)
		return fmt.Errorf("config flow aborted: %s", reason)
	}

	// Step 2: Submit the state characteristic and options
	step2Data := map[string]interface{}{
		"state_characteristic": helperStatisticsCreateCharacteristic,
		"precision":            helperStatisticsCreatePrecision,
		"keep_last_sample":     false,
	}

	if hasSamplingSize {
		step2Data["samples_max_buffer_size"] = helperStatisticsCreateSamplingSize
	}
	if hasMaxAge {
		step2Data["max_age"] = parseDurationForStatistics(helperStatisticsCreateMaxAge)
	}
	if helperStatisticsCreateCharacteristic == "percentile" {
		step2Data["percentile"] = helperStatisticsCreatePercentile
	}

	finalResult, err := rest.ConfigFlowStep(flowID, step2Data)
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
		"entity":         helperStatisticsCreateEntity,
		"characteristic": helperStatisticsCreateCharacteristic,
	}
	if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
		if entryID, ok := entryResult["entry_id"]; ok {
			result["entry_id"] = entryID
		}
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Statistics sensor '%s' created successfully.", name))
	return nil
}

// parseDurationForStatistics converts HH:MM:SS format to a duration object for statistics
func parseDurationForStatistics(s string) map[string]interface{} {
	hours, minutes, seconds := 0, 0, 0
	fmt.Sscanf(s, "%d:%d:%d", &hours, &minutes, &seconds)
	return map[string]interface{}{
		"hours":   hours,
		"minutes": minutes,
		"seconds": seconds,
	}
}
