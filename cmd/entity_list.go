package cmd

import (
	"fmt"
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

	// Build lookup maps
	registryMap := buildRegistryMap(registry)
	deviceMap, deviceAreaMap := buildDeviceMaps(devices)

	var areaFloorMap map[string]string
	if entityListFloor != "" {
		areaFloorMap = buildAreaFloorMap(areas)
	}

	// Filter and build entity list
	entities := filterAndBuildEntities(states, registryMap, deviceAreaMap, areaFloorMap, entityFilterCriteria{
		EntityID:    entityListID,
		Domain:      entityListDomain,
		Area:        entityListArea,
		Floor:       entityListFloor,
		Label:       entityListLabel,
		Device:      entityListDevice,
		DeviceClass: entityListDeviceClass,
	})

	// Count mode
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
