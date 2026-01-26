package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	actionCallName           string
	actionCallData           string
	actionCallEntity         string
	actionCallArea           string
	actionCallReturnResponse bool
)

var actionCallCmd = &cobra.Command{
	Use:   "call [domain.action]",
	Short: "Call an action with data",
	Long:  `Call a Home Assistant action (service) with optional data.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runActionCall,
}

func init() {
	actionCmd.AddCommand(actionCallCmd)
	actionCallCmd.Flags().StringVar(&actionCallName, "action", "", "Action name in domain.action format")
	actionCallCmd.Flags().StringVarP(&actionCallData, "data", "d", "", "Action data as JSON")
	actionCallCmd.Flags().StringVarP(&actionCallEntity, "entity", "e", "", "Target entity ID")
	actionCallCmd.Flags().StringVarP(&actionCallArea, "area", "a", "", "Target area ID")
	actionCallCmd.Flags().BoolVarP(&actionCallReturnResponse, "return-response", "r", false, "Return action response")
}

func runActionCall(cmd *cobra.Command, args []string) error {
	actionName := actionCallName
	if actionName == "" && len(args) > 0 {
		actionName = args[0]
	}
	if actionName == "" {
		return fmt.Errorf("action name is required (use --action flag or positional argument)")
	}
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Parse action name
	parts := strings.SplitN(actionName, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid action format: %s. Expected domain.action", actionName)
	}
	domain := parts[0]
	service := parts[1]

	// Parse data
	serviceData := make(map[string]interface{})
	if actionCallData != "" {
		if err := json.Unmarshal([]byte(actionCallData), &serviceData); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
	}

	// Add target
	if actionCallEntity != "" {
		serviceData["entity_id"] = actionCallEntity
	}
	if actionCallArea != "" {
		serviceData["area_id"] = actionCallArea
	}

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	if actionCallReturnResponse {
		serviceData["return_response"] = true
	}

	result, err := restClient.CallService(domain, service, serviceData)
	if err != nil {
		return err
	}

	if actionCallReturnResponse && result != nil {
		client.PrintOutput(result, textMode, "")
	} else {
		client.PrintSuccess(nil, textMode, fmt.Sprintf("Action %s called successfully.", actionName))
	}

	return nil
}
