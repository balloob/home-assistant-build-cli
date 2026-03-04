package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	entityRenameID   string
	entityRenameName string
)

var entityRenameCmd = &cobra.Command{
	Use:   "rename [entity_id] [new_name]",
	Short: "Rename an entity",
	Long:  `Rename an entity by setting its friendly name.`,
	Args:  cobra.MaximumNArgs(2),
	RunE:  runEntityRename,
}

func init() {
	entityCmd.AddCommand(entityRenameCmd)
	entityRenameCmd.Flags().StringVar(&entityRenameID, "entity", "", "Entity ID to rename")
	entityRenameCmd.Flags().StringVar(&entityRenameName, "name", "", "New friendly name")
}

func runEntityRename(cmd *cobra.Command, args []string) error {
	entityID, err := resolveArg(entityRenameID, args, 0, "entity ID")
	if err != nil {
		return err
	}
	newName, err := resolveArg(entityRenameName, args, 1, "new name")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.EntityRegistryUpdate(entityID, map[string]interface{}{
		"name": newName,
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Entity renamed to %s.", newName))
	return nil
}
