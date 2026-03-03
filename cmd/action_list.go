package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var actionListCmd = &cobra.Command{
	Use:   "list [domain]",
	Short: "List available actions",
	Long:  `List all available actions, optionally filtered by domain.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runActionList,
}

func init() {
	actionCmd.AddCommand(actionListCmd)
}

func runActionList(cmd *cobra.Command, args []string) error {
	var domain string
	if len(args) > 0 {
		domain = args[0]
	}

	textMode := getTextMode()

	restClient, err := getRESTClient()
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

	output.PrintOutput(actions, textMode, "")
	return nil
}
