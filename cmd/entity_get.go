package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var entityGetRelated bool

var entityGetCmd = &cobra.Command{
	Use:   "get <entity_id>",
	Short: "Get entity state, attributes, and registry data",
	Long:  `Get the current state, attributes, and registry data of an entity. Use --related to also show related automations, scripts, scenes, and devices.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEntityGet,
}

func init() {
	entityCmd.AddCommand(entityGetCmd)
	entityGetCmd.Flags().BoolVarP(&entityGetRelated, "related", "r", false, "Include related items (automations, scripts, scenes, devices)")
}

func runEntityGet(cmd *cobra.Command, args []string) error {
	entityID := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)

	// Get state from REST API
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	state, err := restClient.GetState(entityID)
	if err != nil {
		return err
	}

	// Get registry data and optionally related items via WebSocket
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		// Fall back to just state if we can't get WebSocket connection
		client.PrintOutput(state, textMode, "")
		return nil
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		// Fall back to just state
		client.PrintOutput(state, textMode, "")
		return nil
	}
	defer ws.Close()

	// Get entity registry data
	registry, err := ws.EntityRegistryGet(entityID)
	if err != nil {
		// Entity might not be in registry, just return state
		client.PrintOutput(state, textMode, "")
		return nil
	}

	// Build result combining state and registry data
	result := make(map[string]interface{})

	// Copy state data
	for k, v := range state {
		result[k] = v
	}

	// Add registry data under "registry" key
	result["registry"] = registry

	// Get related items if requested
	if entityGetRelated {
		related, err := ws.SearchRelated("entity", entityID)
		if err == nil && len(related) > 0 {
			result["related"] = related
		}
	}

	client.PrintOutput(result, textMode, "")
	return nil
}
