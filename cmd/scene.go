package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

const (
	sceneGroupCommands = groupCommands
)

var sceneCmd = &cobra.Command{
	Use:     "scene",
	Short:   "Manage scenes",
	Long:    `Create, update, delete, and activate scenes.`,
	GroupID: "automation",
}

// resolveSceneConfigID converts a scene entity_id (e.g. "scene.romantic" or
// "romantic") to the internal config ID stored in attributes.id.
func resolveSceneConfigID(restClient client.RestAPI, entityOrConfigID string) (string, error) {
	return resolveStateConfigID(restClient, "scene", entityOrConfigID)
}

func init() {
	rootCmd.AddCommand(sceneCmd)

	sceneCmd.AddGroup(
		&cobra.Group{ID: sceneGroupCommands, Title: "Commands:"},
	)

	RegisterConfigResourceCRUD(ConfigResourceConfig{
		ParentCmd:           sceneCmd,
		ResourceName:        "scene",
		APIPrefix:           "config/scene/config/",
		IDFlagName:          "scene",
		ResolveID:           resolveSceneConfigID,
		GroupID:             sceneGroupCommands,
		RequiredCreateField: "name",
		CreateExample: `  hab scene create my_scene -d '{"name":"Romantic","entities":{"light.living_room":{"state":"on","brightness":80}}}'
  hab scene create cozy_night -f scene.yaml`,
	})
}
