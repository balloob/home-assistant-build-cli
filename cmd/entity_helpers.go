package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
)

// ──────────────────────────────────────────────────────────
// Data resolution helpers
// ──────────────────────────────────────────────────────────

// resolveDeviceClass resolves the device class for an entity.
// Registry's original_device_class takes precedence, then registry device_class,
// then state attributes device_class.
func resolveDeviceClass(regEntry map[string]interface{}, attrs map[string]interface{}) string {
	if regEntry != nil {
		if dc, ok := regEntry["original_device_class"].(string); ok && dc != "" {
			return dc
		}
		if dc, ok := regEntry["device_class"].(string); ok && dc != "" {
			return dc
		}
	}
	if attrs != nil {
		if dc, ok := attrs["device_class"].(string); ok {
			return dc
		}
	}
	return ""
}

// resolveEntityArea resolves the area ID for an entity, falling back to the
// device's area if the entity has no direct area assignment.
func resolveEntityArea(regEntry map[string]interface{}, deviceAreaMap map[string]string) string {
	if regEntry == nil {
		return ""
	}
	if areaID, _ := regEntry["area_id"].(string); areaID != "" {
		return areaID
	}
	deviceID, _ := regEntry["device_id"].(string)
	return deviceAreaMap[deviceID]
}

// ──────────────────────────────────────────────────────────
// Lookup map builders
// ──────────────────────────────────────────────────────────

// buildRegistryMap creates a lookup map from entity_id to registry entry.
func buildRegistryMap(registry []interface{}) map[string]map[string]interface{} {
	m := make(map[string]map[string]interface{}, len(registry))
	for _, e := range registry {
		if entry, ok := e.(map[string]interface{}); ok {
			if entityID, ok := entry["entity_id"].(string); ok {
				m[entityID] = entry
			}
		}
	}
	return m
}

// buildDeviceMaps creates device name and device area lookup maps from device
// registry data. The names map prefers name_by_user over name.
func buildDeviceMaps(devices []interface{}) (deviceNames map[string]string, deviceAreas map[string]string) {
	deviceNames = make(map[string]string, len(devices))
	deviceAreas = make(map[string]string, len(devices))
	for _, d := range devices {
		if device, ok := d.(map[string]interface{}); ok {
			deviceID, _ := device["id"].(string)
			name, _ := device["name"].(string)
			nameByUser, _ := device["name_by_user"].(string)
			if nameByUser != "" {
				deviceNames[deviceID] = nameByUser
			} else if name != "" {
				deviceNames[deviceID] = name
			}
			if areaID, _ := device["area_id"].(string); areaID != "" {
				deviceAreas[deviceID] = areaID
			}
		}
	}
	return
}

// buildAreaFloorMap creates an area_id to floor_id lookup map.
func buildAreaFloorMap(areas []interface{}) map[string]string {
	m := make(map[string]string, len(areas))
	for _, a := range areas {
		if area, ok := a.(map[string]interface{}); ok {
			areaID, _ := area["area_id"].(string)
			floorID, _ := area["floor_id"].(string)
			if areaID != "" {
				m[areaID] = floorID
			}
		}
	}
	return m
}

// ──────────────────────────────────────────────────────────
// Filtering
// ──────────────────────────────────────────────────────────

// entityFilterCriteria holds the filter criteria for entity listing.
type entityFilterCriteria struct {
	EntityID    string
	Domain      string
	Area        string
	Floor       string
	Label       string
	Device      string
	DeviceClass string
}

// filterAndBuildEntities filters states against the registry and criteria,
// returning enriched entity maps. Each entity map includes entity_id, state,
// name, area_id, device_id, device_class, labels, and disabled.
func filterAndBuildEntities(
	states []interface{},
	registryMap map[string]map[string]interface{},
	deviceAreaMap map[string]string,
	areaFloorMap map[string]string,
	f entityFilterCriteria,
) []map[string]interface{} {
	entities := make([]map[string]interface{}, 0, len(states))

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
		if f.EntityID != "" && entityID != f.EntityID {
			continue
		}

		// Apply domain filter
		if f.Domain != "" && entityDomain != f.Domain {
			continue
		}

		regEntry := registryMap[entityID]

		// Apply device filter
		if f.Device != "" {
			if regEntry == nil {
				continue
			}
			deviceID, _ := regEntry["device_id"].(string)
			if deviceID != f.Device {
				continue
			}
		}

		// Apply area filter (entity area takes precedence, falls back to device area)
		if f.Area != "" {
			areaID := resolveEntityArea(regEntry, deviceAreaMap)
			if areaID != f.Area {
				continue
			}
		}

		// Apply floor filter (resolve area, then check area-to-floor mapping)
		if f.Floor != "" {
			areaID := resolveEntityArea(regEntry, deviceAreaMap)
			if areaID == "" || areaFloorMap[areaID] != f.Floor {
				continue
			}
		}

		// Apply label filter
		if f.Label != "" {
			if regEntry == nil {
				continue
			}
			labels, _ := regEntry["labels"].([]interface{})
			hasLabel := false
			for _, l := range labels {
				if labelStr, ok := l.(string); ok && labelStr == f.Label {
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
		deviceClass := resolveDeviceClass(regEntry, attrs)
		if f.DeviceClass != "" && deviceClass != f.DeviceClass {
			continue
		}

		friendlyName, _ := attrs["friendly_name"].(string)

		var areaID, deviceID string
		var labels []interface{}
		var disabled bool
		if regEntry != nil {
			areaID, _ = regEntry["area_id"].(string)
			deviceID, _ = regEntry["device_id"].(string)
			labels, _ = regEntry["labels"].([]interface{})
			disabled = regEntry["disabled_by"] != nil
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

	return entities
}

// ──────────────────────────────────────────────────────────
// Text-mode rendering
// ──────────────────────────────────────────────────────────

// printEntitiesGroupedByDevice renders entities grouped by device in text mode.
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

// printEntityText renders a single entity line in text mode.
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

// modifyEntityLabels fetches an entity's current labels, applies a modifier
// function, and updates the entity registry.  The modifier returns the new
// label list and a message.  If the returned list is nil, the message is
// treated as an early-exit notice (e.g. "already has label") and no update
// is performed.
func modifyEntityLabels(entityID, labelID string, modify func(labels []string) ([]string, string)) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	entity, err := ws.EntityRegistryGet(entityID)
	if err != nil {
		return err
	}

	currentLabels, _ := entity["labels"].([]interface{})
	labels := make([]string, 0, len(currentLabels))
	for _, l := range currentLabels {
		if ls, ok := l.(string); ok {
			labels = append(labels, ls)
		}
	}

	newLabels, msg := modify(labels)
	if newLabels == nil {
		output.PrintSuccess(nil, textMode, msg)
		return nil
	}

	result, err := ws.EntityRegistryUpdate(entityID, map[string]interface{}{
		"labels": newLabels,
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, msg)
	return nil
}
