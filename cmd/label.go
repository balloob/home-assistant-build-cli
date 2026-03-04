package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var labelCmd = &cobra.Command{
	Use:     "label",
	Short:   "Manage labels",
	Long:    `Create, update, and delete labels.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(labelCmd)

	RegisterRegistryCRUD(RegistryCRUDConfig{
		ParentCmd:    labelCmd,
		ResourceName: "label",
		IDField:      "label_id",
		SearchType:   "label",
		IDFlagName:   "label",
		ListFilters: []RegistryFilterDef{
			{FlagName: "label-id", FieldKey: "label_id", Usage: "Filter by label ID"},
		},
		BriefFields: []string{"label_id", "name"},
		CreateFlags: []RegistryFlagDef{
			{Name: "icon", ParamKey: "icon", Usage: "Icon for the label", Type: FlagString},
			{Name: "color", ParamKey: "color", Usage: "Color for the label", Type: FlagString},
			{Name: "description", ParamKey: "description", Usage: "Description of the label", Type: FlagString},
		},
		UpdateFlags: []RegistryFlagDef{
			{Name: "icon", ParamKey: "icon", Usage: "Icon for the label", Type: FlagString},
			{Name: "color", ParamKey: "color", Usage: "Color for the label", Type: FlagString},
			{Name: "description", ParamKey: "description", Usage: "Description of the label", Type: FlagString},
		},
		ListFunc: func(ws client.WebSocketAPI) ([]interface{}, error) {
			return ws.LabelRegistryList()
		},
		CreateFunc: func(ws client.WebSocketAPI, name string, params map[string]interface{}) (map[string]interface{}, error) {
			return ws.LabelRegistryCreate(name, params)
		},
		UpdateFunc: func(ws client.WebSocketAPI, id string, params map[string]interface{}) (map[string]interface{}, error) {
			return ws.LabelRegistryUpdate(id, params)
		},
		DeleteFunc: func(ws client.WebSocketAPI, id string) error {
			return ws.LabelRegistryDelete(id)
		},
	})
}
