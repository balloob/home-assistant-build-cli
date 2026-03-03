package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var entityEnableCmd = &cobra.Command{
	Use:   "enable <entity_id>",
	Short: "Enable a disabled entity",
	Long:  `Enable an entity that was previously disabled.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEntityEnable,
}

func init() {
	entityCmd.AddCommand(entityEnableCmd)
}

func runEntityEnable(cmd *cobra.Command, args []string) error {
	entityID := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.EntityRegistryUpdate(entityID, map[string]interface{}{
		"disabled_by": nil,
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Entity %s enabled.", entityID))
	return nil
}
