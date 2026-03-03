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
	entityID := entityRenameID
	if entityID == "" && len(args) > 0 {
		entityID = args[0]
	}
	if entityID == "" {
		return fmt.Errorf("entity ID is required (use --entity flag or first positional argument)")
	}
	newName := entityRenameName
	if newName == "" && len(args) > 1 {
		newName = args[1]
	}
	if newName == "" {
		return fmt.Errorf("new name is required (use --name flag or second positional argument)")
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
