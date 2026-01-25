package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Show an overview of the Home Assistant instance",
	Long:  `Show aggregated counts of floors, areas, devices, entities, automations, scripts, and helpers.`,
	RunE:  runOverview,
}

func init() {
	rootCmd.AddCommand(overviewCmd)
}

func runOverview(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	// Gather all counts
	result := make(map[string]interface{})

	// Count floors
	floors, err := ws.FloorRegistryList()
	if err == nil {
		result["floors"] = len(floors)
	}

	// Count areas
	areas, err := ws.AreaRegistryList()
	if err == nil {
		result["areas"] = len(areas)
	}

	// Count devices
	devices, err := ws.DeviceRegistryList()
	if err == nil {
		result["devices"] = len(devices)
	}

	// Count dashboards
	dashboards, err := ws.SendCommand("lovelace/dashboards/list", nil)
	if err == nil {
		if dashboardList, ok := dashboards.([]interface{}); ok {
			result["dashboards"] = len(dashboardList)
		}
	}

	// Count labels
	labels, err := ws.LabelRegistryList()
	if err == nil {
		result["labels"] = len(labels)
	}

	// Get states to count entities, automations, scripts
	states, err := ws.GetStates()
	if err == nil {
		entityCount := 0
		automationCount := 0
		scriptCount := 0
		entitiesByDomain := make(map[string]int)

		for _, s := range states {
			state, ok := s.(map[string]interface{})
			if !ok {
				continue
			}

			entityID, _ := state["entity_id"].(string)
			parts := strings.SplitN(entityID, ".", 2)
			if len(parts) < 2 {
				continue
			}

			domain := parts[0]
			entityCount++
			entitiesByDomain[domain]++

			if domain == "automation" {
				automationCount++
			} else if domain == "script" {
				scriptCount++
			}
		}

		result["entities"] = entityCount
		result["entities_by_domain"] = entitiesByDomain
		result["automations"] = automationCount
		result["scripts"] = scriptCount
	}

	// Count helpers from entity registry
	entities, err := ws.EntityRegistryList()
	if err == nil {
		helperDomains := map[string]bool{
			"input_boolean":  true,
			"input_number":   true,
			"input_text":     true,
			"input_select":   true,
			"input_datetime": true,
			"input_button":   true,
			"counter":        true,
			"timer":          true,
			"schedule":       true,
		}

		helperCount := 0
		for _, e := range entities {
			entity, ok := e.(map[string]interface{})
			if !ok {
				continue
			}

			entityID, _ := entity["entity_id"].(string)
			parts := strings.SplitN(entityID, ".", 2)
			if len(parts) < 2 {
				continue
			}

			if helperDomains[parts[0]] {
				helperCount++
			}
		}
		result["helpers"] = helperCount
	}

	client.PrintOutput(result, textMode, "")
	return nil
}
