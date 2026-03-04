package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var overviewCmd = &cobra.Command{
	Use:     "overview",
	Short:   "Show an overview of the Home Assistant instance",
	Long:    `Show aggregated counts of floors, areas, devices, entities, automations, scripts, and helpers.`,
	RunE:    runOverview,
	GroupID: "start",
}

func init() {
	rootCmd.AddCommand(overviewCmd)
}

func runOverview(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	// Get REST client for config
	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	// Fetch all data concurrently — the REST call runs on a separate HTTP
	// connection while the 7 WS calls share one connection (with
	// concurrency-safe SendCommand).  All calls are optional; errors are
	// silently ignored per the original behavior.
	var (
		configData interface{}
		floors     []interface{}
		areas      []interface{}
		devices    []interface{}
		dashboards interface{}
		labels     []interface{}
		states     []interface{}
		entities   []interface{}

		configErr     error
		floorsErr     error
		areasErr      error
		devicesErr    error
		dashboardsErr error
		labelsErr     error
		statesErr     error
		entitiesErr   error
	)

	var wg sync.WaitGroup
	wg.Add(8)
	go func() { defer wg.Done(); configData, configErr = restClient.Get("config") }()
	go func() { defer wg.Done(); floors, floorsErr = ws.FloorRegistryList() }()
	go func() { defer wg.Done(); areas, areasErr = ws.AreaRegistryList() }()
	go func() { defer wg.Done(); devices, devicesErr = ws.DeviceRegistryList() }()
	go func() { defer wg.Done(); dashboards, dashboardsErr = ws.SendCommand("lovelace/dashboards/list", nil) }()
	go func() { defer wg.Done(); labels, labelsErr = ws.LabelRegistryList() }()
	go func() { defer wg.Done(); states, statesErr = ws.GetStates() }()
	go func() { defer wg.Done(); entities, entitiesErr = ws.EntityRegistryList() }()
	wg.Wait()

	// Assemble result from fetched data
	result := make(map[string]interface{})

	if configErr == nil {
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

	if floorsErr == nil {
		result["floors"] = len(floors)
	}
	if areasErr == nil {
		result["areas"] = len(areas)
	}
	if devicesErr == nil {
		result["devices"] = len(devices)
	}
	if dashboardsErr == nil {
		if dashboardList, ok := dashboards.([]interface{}); ok {
			result["dashboards"] = len(dashboardList)
		}
	}
	if labelsErr == nil {
		result["labels"] = len(labels)
	}

	if statesErr == nil {
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

	if entitiesErr == nil {
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
		output.PrintOutput(result, false, "")
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
