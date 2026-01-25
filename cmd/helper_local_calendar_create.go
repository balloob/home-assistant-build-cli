package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperLocalCalendarCreateIcon string
)

var helperLocalCalendarCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new local calendar helper",
	Long: `Create a new local calendar helper for storing calendar events locally.

Local calendars allow you to create and manage calendar events directly in Home Assistant
without relying on external services.`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperLocalCalendarCreate,
}

func init() {
	helperLocalCalendarParentCmd.AddCommand(helperLocalCalendarCreateCmd)
	helperLocalCalendarCreateCmd.Flags().StringVarP(&helperLocalCalendarCreateIcon, "icon", "i", "", "Icon for the calendar (e.g., mdi:calendar)")
}

func runHelperLocalCalendarCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	// Use REST API for config flows
	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// Step 1: Start the config flow for local_calendar
	flowResult, err := rest.ConfigFlowCreate("local_calendar")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}

	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Step 2: Submit the form data with calendar name
	formData := map[string]interface{}{
		"calendar_name": name,
	}

	if helperLocalCalendarCreateIcon != "" {
		formData["icon"] = helperLocalCalendarCreateIcon
	}

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create local calendar: %w", err)
	}

	// Check result type
	resultType, _ := finalResult["type"].(string)
	if resultType == "abort" {
		reason, _ := finalResult["reason"].(string)
		return fmt.Errorf("config flow aborted: %s", reason)
	}

	if resultType != "create_entry" {
		return fmt.Errorf("unexpected flow result type: %s", resultType)
	}

	// Extract result data
	result := map[string]interface{}{
		"title": finalResult["title"],
	}
	if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
		if entryID, ok := entryResult["entry_id"]; ok {
			result["entry_id"] = entryID
		}
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Local calendar '%s' created successfully.", name))
	return nil
}
