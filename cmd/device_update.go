package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/registryupdate"
	"github.com/spf13/cobra"
)

func init() { deviceCmd.AddCommand(newDeviceUpdateCommand()) }

func newDeviceUpdateCommand() *cobra.Command {
	cmd := newRegistryUpdateCommand(RegistryCRUDConfig{
		ResourceName: "device", IDField: "id", NameField: "name_by_user", NullableName: true,
		UpdateFlags: []RegistryFlagDef{
			{Name: "area", ParamKey: "area_id", Usage: "Area ID to assign", Type: FlagString},
			{Name: "labels", ParamKey: "labels", Usage: "Label IDs (comma-separated or repeated)", Type: FlagStringSet},
			{Name: "disabled", ParamKey: "disabled_by", Usage: "Disable the device and its entities", Type: FlagUserBool},
		},
		ListFunc: func(ws client.WebSocketAPI) ([]any, error) { return ws.DeviceRegistryList() },
		UpdateFunc: func(ws client.WebSocketAPI, id string, params map[string]any) (map[string]any, error) {
			return ws.DeviceRegistryUpdate(id, params)
		},
		ValidateUpdate: validateDeviceUpdate,
	})
	cmd.Use = "update <device_id>"
	cmd.Flags().Lookup("name").Usage = "Name override (empty restores the integration's name)"
	cmd.Example = `  hab device update <device_id> --name "Kitchen plug" --area kitchen --plan
  hab device update <device_id> --labels lighting,downstairs
  hab device update <device_id> --name "" --area "" --labels "" --disabled=false`
	return cmd
}

func validateDeviceUpdate(ws client.WebSocketAPI, params map[string]any) error {
	if area, ok := params["area_id"].(string); ok {
		if err := registryupdate.RequireIDs(ws.AreaRegistryList, "area_id", []string{area}); err != nil {
			return err
		}
	}
	if labels, ok := params["labels"].([]string); ok {
		return registryupdate.RequireIDs(ws.LabelRegistryList, "label_id", labels)
	}
	return nil
}
