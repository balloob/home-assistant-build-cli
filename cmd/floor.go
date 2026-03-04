package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var floorCmd = &cobra.Command{
	Use:     "floor",
	Short:   "Manage floors",
	Long:    `Create, update, and delete floors.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(floorCmd)

	RegisterRegistryCRUD(RegistryCRUDConfig{
		ParentCmd:    floorCmd,
		ResourceName: "floor",
		IDField:      "floor_id",
		SearchType:   "floor",
		IDFlagName:   "floor",
		ListFilters: []RegistryFilterDef{
			{FlagName: "floor-id", FieldKey: "floor_id", Usage: "Filter by floor ID"},
		},
		BriefFields: []string{"floor_id", "name"},
		CreateFlags: []RegistryFlagDef{
			{Name: "icon", ParamKey: "icon", Usage: "Icon for the floor", Type: FlagString},
			{Name: "level", ParamKey: "level", Usage: "Floor level (0 = ground)", Type: FlagInt},
		},
		UpdateFlags: []RegistryFlagDef{
			{Name: "icon", ParamKey: "icon", Usage: "Icon for the floor", Type: FlagString},
			{Name: "level", ParamKey: "level", Usage: "Floor level", Type: FlagInt},
		},
		ListFunc: func(ws client.WebSocketAPI) ([]interface{}, error) {
			return ws.FloorRegistryList()
		},
		CreateFunc: func(ws client.WebSocketAPI, name string, params map[string]interface{}) (map[string]interface{}, error) {
			return ws.FloorRegistryCreate(name, params)
		},
		UpdateFunc: func(ws client.WebSocketAPI, id string, params map[string]interface{}) (map[string]interface{}, error) {
			return ws.FloorRegistryUpdate(id, params)
		},
		DeleteFunc: func(ws client.WebSocketAPI, id string) error {
			return ws.FloorRegistryDelete(id)
		},
	})
}
