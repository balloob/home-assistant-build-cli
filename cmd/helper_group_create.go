package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperGroupCreateEntities    []string
	helperGroupCreateType        string
	helperGroupCreateAll         bool
	helperGroupCreateHideMembers bool
)

var helperGroupCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new group",
	Long: `Create a new group helper using the config entry flow.

Group types available: binary_sensor, cover, event, fan, light, lock, media_player, sensor, switch.

Examples:
  hab helper-group create "Living Room Lights" --type light --entities light.lamp1,light.lamp2
  hab helper-group create "All Motion Sensors" --type binary_sensor --entities binary_sensor.motion1,binary_sensor.motion2 --all`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperGroupCreate,
}

func init() {
	helperGroupParentCmd.AddCommand(helperGroupCreateCmd)
	helperGroupCreateCmd.Flags().StringVarP(&helperGroupCreateType, "type", "t", "light", "Group type: binary_sensor, cover, event, fan, light, lock, media_player, sensor, switch")
	helperGroupCreateCmd.Flags().StringSliceVarP(&helperGroupCreateEntities, "entities", "e", nil, "Entity IDs to include in the group (required)")
	helperGroupCreateCmd.Flags().BoolVar(&helperGroupCreateAll, "all", false, "Set to true if all entities must be on for group to be on (only for binary_sensor, light, switch)")
	helperGroupCreateCmd.Flags().BoolVar(&helperGroupCreateHideMembers, "hide-members", false, "Hide member entities from the UI")
	helperGroupCreateCmd.MarkFlagRequired("entities")
}

func runHelperGroupCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Validate group type
	validTypes := map[string]bool{
		"binary_sensor": true,
		"cover":         true,
		"event":         true,
		"fan":           true,
		"light":         true,
		"lock":          true,
		"media_player":  true,
		"sensor":        true,
		"switch":        true,
	}
	if !validTypes[helperGroupCreateType] {
		return fmt.Errorf("invalid group type: %s. Valid types: binary_sensor, cover, event, fan, light, lock, media_player, sensor, switch", helperGroupCreateType)
	}

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	// Use REST API for config flows
	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// Step 1: Start the config flow for group
	flowResult, err := rest.ConfigFlowCreate("group")
	if err != nil {
		return fmt.Errorf("failed to start config flow: %w", err)
	}

	flowID, ok := flowResult["flow_id"].(string)
	if !ok {
		return fmt.Errorf("no flow_id in response")
	}

	// Step 2: Select the group type (menu step)
	menuResult, err := rest.ConfigFlowStep(flowID, map[string]interface{}{
		"next_step_id": helperGroupCreateType,
	})
	if err != nil {
		return fmt.Errorf("failed to select group type: %w", err)
	}

	// Check if we need another step
	stepType, _ := menuResult["type"].(string)
	if stepType == "abort" {
		reason, _ := menuResult["reason"].(string)
		return fmt.Errorf("config flow aborted: %s", reason)
	}

	// Step 3: Submit the form data
	formData := map[string]interface{}{
		"name":         name,
		"entities":     helperGroupCreateEntities,
		"type":         helperGroupCreateType,
		"hide_members": helperGroupCreateHideMembers,
	}

	// Add "all" flag for types that support it
	if helperGroupCreateType == "binary_sensor" || helperGroupCreateType == "light" || helperGroupCreateType == "switch" {
		formData["all"] = helperGroupCreateAll
	}

	finalResult, err := rest.ConfigFlowStep(flowID, formData)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
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
		"title":    finalResult["title"],
		"type":     helperGroupCreateType,
		"entities": helperGroupCreateEntities,
	}
	if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
		if entryID, ok := entryResult["entry_id"]; ok {
			result["entry_id"] = entryID
		}
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Group '%s' created successfully.", name))
	return nil
}
