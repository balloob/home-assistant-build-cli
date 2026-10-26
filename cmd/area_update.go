package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/registryupdate"
)

func validateAreaUpdate(ws client.WebSocketAPI, params map[string]any) error {
	if floor, ok := params["floor_id"].(string); ok {
		if err := registryupdate.RequireIDs(ws.FloorRegistryList, "floor_id", []string{floor}); err != nil {
			return err
		}
	}
	if labels, ok := params["labels"].([]string); ok {
		if err := registryupdate.RequireIDs(ws.LabelRegistryList, "label_id", labels); err != nil {
			return err
		}
	}
	return registryupdate.ValidateAreaSensors(ws.GetStates, params)
}
