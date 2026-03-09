package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
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
	Example: `  hab action call light.turn_on -e light.kitchen
  hab action call climate.set_temperature -e climate.living_room -d '{"temperature": 22}'
  hab action call weather.get_forecasts -e weather.home -d '{"type":"daily"}' -r
  hab action call light.turn_off -a living_room`,
	Args: cobra.MaximumNArgs(1),
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
	actionName, err := resolveArg(actionCallName, args, 0, "action name")
	if err != nil {
		return err
	}
	textMode := getTextMode()

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

	restClient, err := getRESTClient()
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
		output.PrintOutput(result, textMode, "")
	} else {
		output.PrintSuccess(nil, textMode, fmt.Sprintf("Action %s called successfully.", actionName))
	}

	return nil
}
