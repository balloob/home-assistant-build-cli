package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var sceneActivateTransition float64

var sceneActivateCmd = &cobra.Command{
	Use:   "activate <scene_id>",
	Short: "Activate a scene",
	Long:  `Activate a scene, optionally with a transition duration.`,
	Example: `  hab scene activate scene.romantic
  hab scene activate romantic --transition 2.5`,
	Args: cobra.ExactArgs(1),
	RunE: runSceneActivate,
}

func init() {
	sceneCmd.AddCommand(sceneActivateCmd)
	sceneActivateCmd.Flags().Float64Var(&sceneActivateTransition, "transition", -1, "Transition duration in seconds")
}

func runSceneActivate(cmd *cobra.Command, args []string) error {
	sceneID := ensureDomainPrefix(args[0], "scene")
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	serviceData := map[string]interface{}{
		"entity_id": sceneID,
	}
	if cmd.Flags().Changed("transition") {
		serviceData["transition"] = sceneActivateTransition
	}

	_, err = restClient.CallService("scene", "turn_on", serviceData)
	if err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Scene %s activated.", sceneID))
	return nil
}
