package cmd

import (
	"fmt"
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

	// Get REST client for config
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	// Gather all data
	result := make(map[string]interface{})

	// Get system config
	configData, err := restClient.Get("config")
	if err == nil {
		if config, ok := configData.(map[string]interface{}); ok {
			result["location_name"] = config["location_name"]
			result["version"] = config["version"]
			result["state"] = config["state"]
			result["time_zone"] = config["time_zone"]
			result["elevation"] = config["elevation"]
			result["latitude"] = config["latitude"]
			result["longitude"] = config["longitude"]
			if unitSystem, ok := config["unit_system"].(map[string]interface{}); ok {
				result["temperature_unit"] = unitSystem["temperature"]
			}
		}
	}

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

	if textMode {
		printOverviewText(result)
	} else {
		client.PrintOutput(result, false, "")
	}
	return nil
}

func printOverviewText(data map[string]interface{}) {
	// Instance info
	locationName, _ := data["location_name"].(string)
	version, _ := data["version"].(string)
	state, _ := data["state"].(string)
	timeZone, _ := data["time_zone"].(string)
	tempUnit, _ := data["temperature_unit"].(string)
	latitude, hasLat := data["latitude"].(float64)
	longitude, hasLon := data["longitude"].(float64)
	elevation, hasElev := data["elevation"].(float64)

	if locationName != "" {
		fmt.Printf("%s\n", locationName)
		fmt.Println(strings.Repeat("=", len(locationName)))
	} else {
		fmt.Println("Home Assistant")
		fmt.Println("==============")
	}

	fmt.Println()

	// System info
	if version != "" {
		fmt.Printf("Version: %s", version)
		if state != "" && state != "RUNNING" {
			fmt.Printf(" (%s)", state)
		}
		fmt.Println()
	}
	if timeZone != "" {
		fmt.Printf("Timezone: %s\n", timeZone)
	}
	if hasLat && hasLon {
		fmt.Printf("Location: %.4f, %.4f", latitude, longitude)
		if hasElev && elevation != 0 {
			fmt.Printf(" (elevation: %.0fm)", elevation)
		}
		fmt.Println()
	}
	if tempUnit != "" {
		fmt.Printf("Unit system: %s\n", tempUnit)
	}

	fmt.Println()

	// Registry counts
	fmt.Println("Registry:")
	if floors, ok := data["floors"].(int); ok {
		fmt.Printf("  Floors: %d\n", floors)
	}
	if areas, ok := data["areas"].(int); ok {
		fmt.Printf("  Areas: %d\n", areas)
	}
	if devices, ok := data["devices"].(int); ok {
		fmt.Printf("  Devices: %d\n", devices)
	}
	if labels, ok := data["labels"].(int); ok {
		fmt.Printf("  Labels: %d\n", labels)
	}

	fmt.Println()

	// Entities
	fmt.Println("Entities:")
	if entities, ok := data["entities"].(int); ok {
		fmt.Printf("  Total: %d\n", entities)
	}
	if byDomain, ok := data["entities_by_domain"].(map[string]int); ok && len(byDomain) > 0 {
		// Show top domains
		fmt.Print("  By domain: ")
		count := 0
		for domain, cnt := range byDomain {
			if count > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s (%d)", domain, cnt)
			count++
			if count >= 5 {
				if len(byDomain) > 5 {
					fmt.Printf(", ... +%d more", len(byDomain)-5)
				}
				break
			}
		}
		fmt.Println()
	}

	fmt.Println()

	// Automations & Scripts
	fmt.Println("Automation:")
	if automations, ok := data["automations"].(int); ok {
		fmt.Printf("  Automations: %d\n", automations)
	}
	if scripts, ok := data["scripts"].(int); ok {
		fmt.Printf("  Scripts: %d\n", scripts)
	}
	if helpers, ok := data["helpers"].(int); ok {
		fmt.Printf("  Helpers: %d\n", helpers)
	}

	fmt.Println()

	// Dashboards
	if dashboards, ok := data["dashboards"].(int); ok {
		fmt.Printf("Dashboards: %d\n", dashboards)
	}
}
