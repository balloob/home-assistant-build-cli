package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	entityListDomain string
	entityListArea   string
	entityListFloor  string
	entityListLabel  string
	entityListDevice string
)

var entityListCmd = &cobra.Command{
	Use:   "list",
	Short: "List entities with optional filtering",
	Long:  `List all entities with optional filtering by domain, area, floor, label, or device.`,
	RunE:  runEntityList,
}

func init() {
	entityCmd.AddCommand(entityListCmd)
	entityListCmd.Flags().StringVarP(&entityListDomain, "domain", "d", "", "Filter by domain (e.g., light, switch)")
	entityListCmd.Flags().StringVarP(&entityListArea, "area", "a", "", "Filter by area ID")
	entityListCmd.Flags().StringVarP(&entityListFloor, "floor", "f", "", "Filter by floor ID (includes all areas on that floor)")
	entityListCmd.Flags().StringVarP(&entityListLabel, "label", "l", "", "Filter by label ID")
	entityListCmd.Flags().StringVar(&entityListDevice, "device", "", "Filter by device ID")
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

	client.PrintOutput(entities, textMode, "")
	return nil
}
