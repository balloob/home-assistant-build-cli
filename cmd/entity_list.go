package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	entityListID     string
	entityListDomain string
	entityListArea   string
	entityListFloor  string
	entityListLabel  string
	entityListDevice string
	entityListCount  bool
	entityListBrief  bool
	entityListLimit  int
)

var entityListCmd = &cobra.Command{
	Use:   "list",
	Short: "List entities with optional filtering",
	Long:  `List all entities with optional filtering by domain, area, floor, label, or device.`,
	RunE:  runEntityList,
}

func init() {
	entityCmd.AddCommand(entityListCmd)
	entityListCmd.Flags().StringVar(&entityListID, "entity-id", "", "Filter by entity ID")
	entityListCmd.Flags().StringVarP(&entityListDomain, "domain", "d", "", "Filter by domain (e.g., light, switch)")
	entityListCmd.Flags().StringVarP(&entityListArea, "area", "a", "", "Filter by area ID")
	entityListCmd.Flags().StringVarP(&entityListFloor, "floor", "f", "", "Filter by floor ID (includes all areas on that floor)")
	entityListCmd.Flags().StringVarP(&entityListLabel, "label", "l", "", "Filter by label ID")
	entityListCmd.Flags().StringVar(&entityListDevice, "device", "", "Filter by device ID")
	entityListCmd.Flags().BoolVarP(&entityListCount, "count", "c", false, "Return only the count of items")
	entityListCmd.Flags().BoolVarP(&entityListBrief, "brief", "b", false, "Return minimal fields (entity_id and name only)")
	entityListCmd.Flags().IntVarP(&entityListLimit, "limit", "n", 0, "Limit results to N items")
}

func runEntityList(cmd *cobra.Command, args []string) error {
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

	// Get entity registry
	registry, err := ws.EntityRegistryList()
	if err != nil {
		return err
	}

	registryMap := make(map[string]map[string]interface{})
	for _, e := range registry {
		if entry, ok := e.(map[string]interface{}); ok {
			if entityID, ok := entry["entity_id"].(string); ok {
				registryMap[entityID] = entry
			}
		}
	}

	// Get device registry for device names
	deviceMap := make(map[string]string) // device_id -> name
	devices, err := ws.DeviceRegistryList()
	if err == nil {
		for _, d := range devices {
			if device, ok := d.(map[string]interface{}); ok {
				deviceID, _ := device["id"].(string)
				name, _ := device["name"].(string)
				nameByUser, _ := device["name_by_user"].(string)
				if nameByUser != "" {
					deviceMap[deviceID] = nameByUser
				} else if name != "" {
					deviceMap[deviceID] = name
				}
			}
		}
	}

	// Build area-to-floor map if floor filter is used
	var areaFloorMap map[string]string
	if entityListFloor != "" {
		areaFloorMap = make(map[string]string)
		areas, err := ws.AreaRegistryList()
		if err == nil {
			for _, a := range areas {
				if area, ok := a.(map[string]interface{}); ok {
					areaID, _ := area["area_id"].(string)
					floorID, _ := area["floor_id"].(string)
					if areaID != "" {
						areaFloorMap[areaID] = floorID
					}
				}
			}
		}
	}

	// Get states
	states, err := ws.GetStates()
	if err != nil {
		return err
	}

	var entities []map[string]interface{}
	for _, s := range states {
		state, ok := s.(map[string]interface{})
		if !ok {
			continue
		}

		entityID, _ := state["entity_id"].(string)
		parts := strings.SplitN(entityID, ".", 2)
		entityDomain := ""
		if len(parts) > 0 {
			entityDomain = parts[0]
		}

		// Apply entity ID filter
		if entityListID != "" && entityID != entityListID {
			continue
		}

		// Apply domain filter
		if entityListDomain != "" && entityDomain != entityListDomain {
			continue
		}

		regEntry := registryMap[entityID]

		// Apply device filter
		if entityListDevice != "" {
			if regEntry == nil {
				continue
			}
			deviceID, _ := regEntry["device_id"].(string)
			if deviceID != entityListDevice {
				continue
			}
		}

		// Apply area filter
		if entityListArea != "" {
			if regEntry == nil {
				continue
			}
			if areaID, _ := regEntry["area_id"].(string); areaID != entityListArea {
				continue
			}
		}

		// Apply floor filter (check if entity's area is on the specified floor)
		if entityListFloor != "" {
			if regEntry == nil {
				continue
			}
			areaID, _ := regEntry["area_id"].(string)
			if areaID == "" {
				continue
			}
			floorID := areaFloorMap[areaID]
			if floorID != entityListFloor {
				continue
			}
		}

		// Apply label filter
		if entityListLabel != "" {
			if regEntry == nil {
				continue
			}
			labels, _ := regEntry["labels"].([]interface{})
			hasLabel := false
			for _, l := range labels {
				if labelStr, ok := l.(string); ok && labelStr == entityListLabel {
					hasLabel = true
					break
				}
			}
			if !hasLabel {
				continue
			}
		}

		attrs, _ := state["attributes"].(map[string]interface{})
		friendlyName, _ := attrs["friendly_name"].(string)

		var areaID string
		var deviceID string
		var labels []interface{}
		var disabled bool
		if regEntry != nil {
			areaID, _ = regEntry["area_id"].(string)
			deviceID, _ = regEntry["device_id"].(string)
			labels, _ = regEntry["labels"].([]interface{})
			disabled = regEntry["disabled_by"] != nil
		}

		entities = append(entities, map[string]interface{}{
			"entity_id": entityID,
			"state":     state["state"],
			"name":      friendlyName,
			"area_id":   areaID,
			"device_id": deviceID,
			"labels":    labels,
			"disabled":  disabled,
		})
	}

	// Handle count mode
	if entityListCount {
		if textMode {
			fmt.Printf("Count: %d\n", len(entities))
		} else {
			client.PrintOutput(map[string]interface{}{"count": len(entities)}, false, "")
		}
		return nil
	}

	// Apply limit
	if entityListLimit > 0 && len(entities) > entityListLimit {
		entities = entities[:entityListLimit]
	}

	// Handle brief mode
	if entityListBrief {
		if textMode {
			for _, item := range entities {
				entityID, _ := item["entity_id"].(string)
				name, _ := item["name"].(string)
				if name != "" && name != entityID {
					fmt.Printf("%s (%s)\n", name, entityID)
				} else {
					fmt.Println(entityID)
				}
			}
		} else {
			var brief []map[string]interface{}
			for _, item := range entities {
				brief = append(brief, map[string]interface{}{
					"entity_id": item["entity_id"],
					"name":      item["name"],
				})
			}
			client.PrintOutput(brief, false, "")
		}
		return nil
	}

	// Full output
	if textMode {
		if len(entities) == 0 {
			fmt.Println("No entities.")
			return nil
		}
		printEntitiesGroupedByDevice(entities, deviceMap)
	} else {
		client.PrintOutput(entities, false, "")
	}
	return nil
}

func printEntitiesGroupedByDevice(entities []map[string]interface{}, deviceNames map[string]string) {
	// Group by device_id
	byDevice := make(map[string][]map[string]interface{})
	var deviceOrder []string

	for _, e := range entities {
		deviceID, _ := e["device_id"].(string)
		if _, exists := byDevice[deviceID]; !exists {
			deviceOrder = append(deviceOrder, deviceID)
		}
		byDevice[deviceID] = append(byDevice[deviceID], e)
	}

	// Print entities without device first
	if noDevice, ok := byDevice[""]; ok {
		fmt.Println("No device:")
		for _, e := range noDevice {
			printEntityText(e, "  ")
		}
		fmt.Println()
	}

	// Print each device group
	for _, deviceID := range deviceOrder {
		if deviceID == "" {
			continue
		}
		deviceEntities := byDevice[deviceID]

		if name, ok := deviceNames[deviceID]; ok && name != "" {
			fmt.Printf("%s (%s):\n", name, deviceID)
		} else {
			fmt.Printf("Device %s:\n", deviceID)
		}
		for _, e := range deviceEntities {
			printEntityText(e, "  ")
		}
		fmt.Println()
	}
}

func printEntityText(e map[string]interface{}, indent string) {
	entityID, _ := e["entity_id"].(string)
	state, _ := e["state"].(string)
	name, _ := e["name"].(string)
	areaID, _ := e["area_id"].(string)

	// First line: entity with state
	if name != "" && name != entityID {
		fmt.Printf("%s%s (%s): %s\n", indent, name, entityID, state)
	} else {
		fmt.Printf("%s%s: %s\n", indent, entityID, state)
	}

	// Additional details
	if areaID != "" {
		fmt.Printf("%s  area: %s\n", indent, areaID)
	}
}
