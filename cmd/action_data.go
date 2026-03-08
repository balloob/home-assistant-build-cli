package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
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

	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	services, err := restClient.GetServices()
	if err != nil {
		return err
	}

	// Filter to actions that always return data (response.optional == false).
	actions := collectActions(services, domain, func(actionInfo map[string]interface{}) bool {
		response, ok := actionInfo["response"].(map[string]interface{})
		if !ok {
			return false
		}
		optional, hasOptional := response["optional"].(bool)
		return hasOptional && !optional
	})

	output.PrintOutput(actions, textMode, "")
	return nil
}
