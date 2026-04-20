package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var backupDeleteForce bool

var backupDeleteCmd = &cobra.Command{
	Use:   "delete <backup_id>",
	Short: "Delete a backup",
	Long:  `Delete a backup by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupDelete,
}

func init() {
	backupCmd.AddCommand(backupDeleteCmd)
	backupDeleteCmd.Flags().BoolVarP(&backupDeleteForce, "force", "f", false, "Skip confirmation")
}

func runBackupDelete(cmd *cobra.Command, args []string) error {
	backupID := args[0]
	textMode := getTextMode()

	if err := confirmAction(backupDeleteForce, fmt.Sprintf("Delete backup %s?", backupID), "delete backup"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.BackupDelete(backupID); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Backup '%s' deleted.", backupID), output.EnvelopeContext{Operation: "delete", ResourceType: "backup"})
	return nil
}
