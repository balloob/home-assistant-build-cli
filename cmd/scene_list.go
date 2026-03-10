package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var sceneListFlags *ListFlags

var sceneListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all scenes",
	Long:    `List all scenes in Home Assistant.`,
	GroupID: sceneGroupCommands,
	RunE:    runSceneList,
}

func init() {
	sceneCmd.AddCommand(sceneListCmd)
	sceneListFlags = RegisterListFlags(sceneListCmd, "entity_id")
}

func runSceneList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	states, err := ws.GetStates()
	if err != nil {
		return err
	}

	// Collect scene entities
	scenes := make([]map[string]interface{}, 0)
	for _, s := range states {
		state, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		entityID, _ := state["entity_id"].(string)
		if !strings.HasPrefix(entityID, "scene.") {
			continue
		}
		attrs, _ := state["attributes"].(map[string]interface{})
		friendlyName, _ := attrs["friendly_name"].(string)
		configID, _ := attrs["id"].(string)

		scenes = append(scenes, map[string]interface{}{
			"entity_id": entityID,
			"name":      friendlyName,
			"id":        configID,
			"state":     state["state"],
		})
	}

	if sceneListFlags.RenderCount(len(scenes), textMode) {
		return nil
	}
	scenes = sceneListFlags.ApplyLimitMap(scenes)
	if sceneListFlags.RenderBriefMap(scenes, textMode, "entity_id", "name") {
		return nil
	}

	if textMode {
		if len(scenes) == 0 {
			fmt.Println("No scenes.")
			return nil
		}
		for _, s := range scenes {
			name, _ := s["name"].(string)
			entityID, _ := s["entity_id"].(string)
			fmt.Printf("%s (%s)\n", name, entityID)
		}
	} else {
		output.PrintOutput(scenes, false, "")
	}
	return nil
}
