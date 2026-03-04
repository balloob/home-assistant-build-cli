package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	entityListID          string
	entityListDomain      string
	entityListArea        string
	entityListFloor       string
	entityListLabel       string
	entityListDevice      string
	entityListDeviceClass string
)

var entityListFlags *ListFlags

var entityListCmd = &cobra.Command{
	Use:   "list",
	Short: "List entities with optional filtering",
	Long:  `List all entities with optional filtering by domain, area, floor, label, device, or device class.`,
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
	entityListCmd.Flags().StringVar(&entityListDeviceClass, "device-class", "", "Filter by device class (e.g., temperature, motion, door)")
	entityListFlags = RegisterListFlags(entityListCmd, "entity_id")
}

func runEntityList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	// Fetch registry data, device data, states (and optionally areas)
	// concurrently — these are independent API calls over the same WS
	// connection which supports multiple in-flight requests.
	var (
		registry []interface{}
		devices  []interface{}
		states   []interface{}
		areas    []interface{}

		registryErr error
		statesErr   error
	)

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		registry, registryErr = ws.EntityRegistryList()
	}()
	go func() {
		defer wg.Done()
		devices, _ = ws.DeviceRegistryList()
	}()
	go func() {
		defer wg.Done()
		states, statesErr = ws.GetStates()
	}()
	if entityListFloor != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			areas, _ = ws.AreaRegistryList()
		}()
	}
	wg.Wait()

	if registryErr != nil {
		return registryErr
	}
	if statesErr != nil {
		return statesErr
	}

	// Build registry lookup map
	registryMap := make(map[string]map[string]interface{})
	for _, e := range registry {
		if entry, ok := e.(map[string]interface{}); ok {
			if entityID, ok := entry["entity_id"].(string); ok {
				registryMap[entityID] = entry
			}
		}
	}

	// Build device lookup maps
	deviceMap := make(map[string]string)     // device_id -> display name
	deviceAreaMap := make(map[string]string) // device_id -> area_id
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
			if areaID, _ := device["area_id"].(string); areaID != "" {
				deviceAreaMap[deviceID] = areaID
			}
		}
	}

	// Build area-to-floor map if floor filter is used
	var areaFloorMap map[string]string
	if entityListFloor != "" {
		areaFloorMap = make(map[string]string)
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

		// Apply area filter — entity-level area_id takes precedence; if the
		// entity has no direct area assignment, fall back to the device's area.
		if entityListArea != "" {
			if regEntry == nil {
				continue
			}
			entityAreaID, _ := regEntry["area_id"].(string)
			if entityAreaID != "" {
				// Entity has an explicit area override — use it directly.
				if entityAreaID != entityListArea {
					continue
				}
			} else {
				// No entity-level area — inherit from the device.
				deviceID, _ := regEntry["device_id"].(string)
				if deviceAreaMap[deviceID] != entityListArea {
					continue
				}
			}
		}

		// Apply floor filter — resolve entity area (with device fallback) then
		// look up whether that area is on the requested floor.
		if entityListFloor != "" {
			if regEntry == nil {
				continue
			}
			areaID, _ := regEntry["area_id"].(string)
			if areaID == "" {
				// Fall back to device area
				deviceID, _ := regEntry["device_id"].(string)
				areaID = deviceAreaMap[deviceID]
			}
			if areaID == "" || areaFloorMap[areaID] != entityListFloor {
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

		// Apply device class filter
		if entityListDeviceClass != "" {
			// Check registry entry first (original_device_class takes precedence)
			var deviceClass string
			if regEntry != nil {
				if dc, ok := regEntry["original_device_class"].(string); ok && dc != "" {
					deviceClass = dc
				} else if dc, ok := regEntry["device_class"].(string); ok && dc != "" {
					deviceClass = dc
				}
			}
			// Fall back to state attributes if not in registry
			if deviceClass == "" {
				if dc, ok := attrs["device_class"].(string); ok {
					deviceClass = dc
				}
			}
			if deviceClass != entityListDeviceClass {
				continue
			}
		}
		friendlyName, _ := attrs["friendly_name"].(string)

		var areaID string
		var deviceID string
		var labels []interface{}
		var disabled bool
		var deviceClass string
		if regEntry != nil {
			areaID, _ = regEntry["area_id"].(string)
			deviceID, _ = regEntry["device_id"].(string)
			labels, _ = regEntry["labels"].([]interface{})
			disabled = regEntry["disabled_by"] != nil
			// Get device class from registry (original_device_class takes precedence)
			if dc, ok := regEntry["original_device_class"].(string); ok && dc != "" {
				deviceClass = dc
			} else if dc, ok := regEntry["device_class"].(string); ok && dc != "" {
				deviceClass = dc
			}
		}
		// Fall back to state attributes if not in registry
		if deviceClass == "" {
			if dc, ok := attrs["device_class"].(string); ok {
				deviceClass = dc
			}
		}

		entities = append(entities, map[string]interface{}{
			"entity_id":    entityID,
			"state":        state["state"],
			"name":         friendlyName,
			"area_id":      areaID,
			"device_id":    deviceID,
			"device_class": deviceClass,
			"labels":       labels,
			"disabled":     disabled,
		})
	}

	if entityListFlags.RenderCount(len(entities), textMode) {
		return nil
	}
	entities = entityListFlags.ApplyLimitMap(entities)

	// Handle brief mode — custom text rendering for name==entityID fallback
	if entityListFlags.Brief {
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
			output.PrintOutput(brief, false, "")
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
		output.PrintOutput(entities, false, "")
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
	deviceClass, _ := e["device_class"].(string)

	// First line: entity with state
	if name != "" && name != entityID {
		fmt.Printf("%s%s (%s): %s\n", indent, name, entityID, state)
	} else {
		fmt.Printf("%s%s: %s\n", indent, entityID, state)
	}

	// Additional details
	if deviceClass != "" {
		fmt.Printf("%s  device_class: %s\n", indent, deviceClass)
	}
	if areaID != "" {
		fmt.Printf("%s  area: %s\n", indent, areaID)
	}
}
