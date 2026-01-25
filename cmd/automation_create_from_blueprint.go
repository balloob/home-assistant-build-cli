package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	automationFromBlueprintData   string
	automationFromBlueprintFile   string
	automationFromBlueprintFormat string
)

var automationCreateFromBlueprintCmd = &cobra.Command{
	Use:   "create-from-blueprint <id> <blueprint-path>",
	Short: "Create an automation from a blueprint",
	Long: `Create a new automation from a blueprint. Provide the automation ID, blueprint path, and blueprint inputs.

Example:
  hab automation create-from-blueprint my_motion_light homeassistant/motion_light \
    -d '{"alias":"Kitchen Motion Light","motion_entity":"binary_sensor.kitchen_motion","light_target":{"entity_id":"light.kitchen"}}'`,
	Args: cobra.ExactArgs(2),
	RunE: runAutomationCreateFromBlueprint,
}

func init() {
	automationCmd.AddCommand(automationCreateFromBlueprintCmd)
	automationCreateFromBlueprintCmd.Flags().StringVarP(&automationFromBlueprintData, "data", "d", "", "Blueprint inputs as JSON (must include alias)")
	automationCreateFromBlueprintCmd.Flags().StringVarP(&automationFromBlueprintFile, "file", "f", "", "Path to inputs file")
	automationCreateFromBlueprintCmd.Flags().StringVar(&automationFromBlueprintFormat, "format", "", "Input format (json, yaml)")
}

func runAutomationCreateFromBlueprint(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	blueprintPath := args[1]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	inputs, err := input.ParseInput(automationFromBlueprintData, automationFromBlueprintFile, automationFromBlueprintFormat)
	if err != nil {
		return err
	}

	// Extract alias from inputs (required)
	alias, ok := inputs["alias"].(string)
	if !ok || alias == "" {
		return fmt.Errorf("automation must have an 'alias' field in the inputs")
	}
	delete(inputs, "alias")

	// Build the automation config with blueprint reference
	config := map[string]interface{}{
		"alias": alias,
		"use_blueprint": map[string]interface{}{
			"path":  blueprintPath,
			"input": inputs,
		},
	}

	// Include description if provided
	if desc, ok := inputs["description"].(string); ok {
		config["description"] = desc
		delete(inputs, "description")
	}

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	result, err := restClient.Post(fmt.Sprintf("config/automation/config/%s", automationID), config)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Automation %s created from blueprint %s.", automationID, blueprintPath))
	return nil
}
