package cmd

import (
	"errors"
	"fmt"
	"sync"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	entityGetRelated bool
	entityGetDevice  bool
	entityGetID      string
)

var entityGetCmd = &cobra.Command{
	Use:   "get [entity_id]",
	Short: "Get entity state, attributes, and registry data",
	Long: `Get the current state, attributes, and registry data of an entity.

Use --related to show related automations, scripts, scenes, and devices.
Use --device to include the parent device information.`,
	Example: `  hab entity get light.kitchen
  hab entity get sensor.temperature -D
  hab entity get light.kitchen -r`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEntityGet,
}

func init() {
	entityCmd.AddCommand(entityGetCmd)
	entityGetCmd.Flags().StringVar(&entityGetID, "entity", "", "Entity ID to get")
	entityGetCmd.Flags().StringVar(&entityGetID, "entity-id", "", "Alias for --entity")
	entityGetCmd.Flags().BoolVarP(&entityGetRelated, "related", "r", false, "Include related items (automations, scripts, scenes, devices)")
	entityGetCmd.Flags().BoolVarP(&entityGetDevice, "device", "D", false, "Include parent device information")
}

func runEntityGet(cmd *cobra.Command, args []string) error {
	entityID, err := resolveArg(entityGetID, args, 0, "entity ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	// Get both clients up-front so we can issue parallel requests.
	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		// Fall back to just REST state if WS is unavailable.
		noteFallback("websocket unavailable; returned REST state without registry, device, or related data")
		noteMissingSection("registry")
		noteMissingSection("device")
		noteMissingSection("related")
		state, stateErr := restClient.GetState(entityID)
		if stateErr != nil {
			return stateErr
		}
		output.PrintOutputWithContext(state, textMode, "", output.EnvelopeContext{Operation: "get", ResourceType: "entity"})
		return nil
	}
	defer ws.Close()

	// Phase 1: fetch state (REST) and registry (WS) concurrently.
	var (
		state       map[string]interface{}
		registry    map[string]interface{}
		stateErr    error
		registryErr error
	)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		state, stateErr = restClient.GetState(entityID)
	}()
	go func() {
		defer wg.Done()
		registry, registryErr = ws.EntityRegistryGet(entityID)
	}()
	wg.Wait()

	// A disabled entity has a registry entry but no state; only a 404 may
	// fall through to the registry so auth/network errors still surface.
	if stateErr != nil && (!isNotFoundError(stateErr) || registryErr != nil) {
		return stateErr
	}
	if registryErr != nil {
		// Entity might not be in registry, just return state
		noteFallback("entity registry unavailable for target; returned REST state only")
		noteMissingSection("registry")
		noteMissingSection("device")
		noteMissingSection("related")
		output.PrintOutputWithContext(state, textMode, "", output.EnvelopeContext{Operation: "get", ResourceType: "entity"})
		return nil
	}

	// Phase 2: fetch optional device and related data concurrently.
	var (
		deviceResult  map[string]interface{}
		relatedResult map[string][]string
	)

	var wg2 sync.WaitGroup
	if entityGetDevice {
		if deviceID, ok := registry["device_id"].(string); ok && deviceID != "" {
			wg2.Add(1)
			go func() {
				defer wg2.Done()
				deviceResult = findDevice(ws, deviceID)
			}()
		}
	}
	if entityGetRelated {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			var relatedErr error
			relatedResult, relatedErr = ws.SearchRelated("entity", entityID)
			if relatedErr != nil {
				noteFallback("entity related search failed; returned state and registry data without related items")
				noteMissingSection("related")
			}
		}()
	}
	wg2.Wait()

	// Build result combining state and registry data
	result := make(map[string]interface{}, len(state)+3)
	for k, v := range state {
		result[k] = v
	}
	result["registry"] = registry

	ctx := output.EnvelopeContext{Operation: "get", ResourceType: "entity"}
	if state == nil {
		result["entity_id"] = entityID
		result["state"] = nil
		note := statelessEntityNote(entityID, registry)
		ctx.Warnings = []string{note}
		if textMode {
			fmt.Println(note)
		}
	}

	if deviceResult != nil {
		result["device"] = deviceResult
	}
	if len(relatedResult) > 0 {
		result["related"] = relatedResult
	}

	output.PrintOutputWithContext(result, textMode, "", ctx)
	return nil
}

// isNotFoundError reports whether err is a NOT_FOUND API error.
func isNotFoundError(err error) bool {
	var apiErr *client.APIError
	return errors.As(err, &apiErr) && apiErr.Code == client.ErrCodeNotFound
}

// statelessEntityNote explains why a registered entity has no state.
func statelessEntityNote(entityID string, registry map[string]interface{}) string {
	if by, ok := registry["disabled_by"].(string); ok && by != "" {
		return fmt.Sprintf("%s is disabled (disabled_by: %s) and has no state.", entityID, by)
	}
	return fmt.Sprintf("%s is in the entity registry but has no state.", entityID)
}

// findDevice fetches the full device registry and returns the matching device.
func findDevice(ws client.WebSocketAPI, deviceID string) map[string]interface{} {
	devices, err := ws.DeviceRegistryList()
	if err != nil {
		return nil
	}
	for _, d := range devices {
		if device, ok := d.(map[string]interface{}); ok {
			if device["id"] == deviceID {
				return device
			}
		}
	}
	return nil
}
