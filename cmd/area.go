package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var areaCmd = &cobra.Command{
	Use:     "area",
	Short:   "Manage areas",
	Long:    `Create, update, and delete areas.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(areaCmd)

	RegisterRegistryCRUD(RegistryCRUDConfig{
		ParentCmd:    areaCmd,
		ResourceName: "area",
		IDField:      "area_id",
		SearchType:   "area",
		IDFlagName:   "area",
		ListFilters: []RegistryFilterDef{
			{FlagName: "area-id", FieldKey: "area_id", Usage: "Filter by area ID"},
			{FlagName: "floor", FieldKey: "floor_id", Usage: "Filter by floor ID"},
		},
		BriefFields: []string{"area_id", "name"},
		CreateFlags: []RegistryFlagDef{
			{Name: "floor", ParamKey: "floor_id", Usage: "Floor ID to assign", Type: FlagString},
			{Name: "icon", ParamKey: "icon", Usage: "Icon for the area", Type: FlagString},
		},
		UpdateFlags: []RegistryFlagDef{
			{Name: "floor", ParamKey: "floor_id", Usage: "Floor ID to assign", Type: FlagString},
			{Name: "icon", ParamKey: "icon", Usage: "Icon for the area", Type: FlagString},
		},
		ListFunc: func(ws client.WebSocketAPI) ([]interface{}, error) {
			return ws.AreaRegistryList()
		},
		CreateFunc: func(ws client.WebSocketAPI, name string, params map[string]interface{}) (map[string]interface{}, error) {
			return ws.AreaRegistryCreate(name, params)
		},
		UpdateFunc: func(ws client.WebSocketAPI, id string, params map[string]interface{}) (map[string]interface{}, error) {
			return ws.AreaRegistryUpdate(id, params)
		},
		DeleteFunc: func(ws client.WebSocketAPI, id string) error {
			return ws.AreaRegistryDelete(id)
		},
	})
}
