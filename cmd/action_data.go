package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var actionDataCmd = &cobra.Command{
	Use:   "data [domain]",
	Short: "List actions that return data",
	Long:  `List all actions that return data (response type = always), optionally filtered by domain.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runActionData,
}

func init() {
	actionCmd.AddCommand(actionDataCmd)
}

func runActionData(cmd *cobra.Command, args []string) error {
	var domain string
	if len(args) > 0 {
		domain = args[0]
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	services, err := restClient.GetServices()
	if err != nil {
		return err
	}

	var actions []map[string]interface{}
	for _, s := range services {
		svc, ok := s.(map[string]interface{})
		if !ok {
			continue
		}

		svcDomain, _ := svc["domain"].(string)
		if domain != "" && svcDomain != domain {
			continue
		}

		svcServices, ok := svc["services"].(map[string]interface{})
		if !ok {
			continue
		}

		for actionName, actionData := range svcServices {
			actionInfo, _ := actionData.(map[string]interface{})

			// Check if this action returns data (response type = always)
			// In Home Assistant, if the "response" field exists with optional=false,
			// the action always returns data
			response, ok := actionInfo["response"].(map[string]interface{})
			if !ok {
				// No response field means action doesn't return data
				continue
			}
			optional, hasOptional := response["optional"].(bool)
			if !hasOptional || optional {
				// If optional field doesn't exist or is true, skip
				// We only want actions where optional=false (always returns data)
				continue
			}

			name, _ := actionInfo["name"].(string)
			if name == "" {
				name = actionName
			}
			description, _ := actionInfo["description"].(string)

			actions = append(actions, map[string]interface{}{
				"action":      svcDomain + "." + actionName,
				"name":        name,
				"description": description,
			})
		}
	}

	client.PrintOutput(actions, textMode, "")
	return nil
}
