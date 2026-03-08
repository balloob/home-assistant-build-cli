package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var entityDisableCmd = &cobra.Command{
	Use:   "disable <entity_id>",
	Short: "Disable an entity",
	Long:  `Disable an entity so it is no longer active.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEntityDisable,
}

func init() {
	entityCmd.AddCommand(entityDisableCmd)
}

func runEntityDisable(cmd *cobra.Command, args []string) error {
	return entitySetDisabled(args[0], true)
}

// entitySetDisabled enables or disables an entity via the registry.
func entitySetDisabled(entityID string, disable bool) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	var disabledBy interface{}
	verb := "enabled"
	if disable {
		disabledBy = "user"
		verb = "disabled"
	}

	result, err := ws.EntityRegistryUpdate(entityID, map[string]interface{}{
		"disabled_by": disabledBy,
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Entity %s %s.", entityID, verb))
	return nil
}
