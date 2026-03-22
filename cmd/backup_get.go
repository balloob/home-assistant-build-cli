package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var backupGetCmd = &cobra.Command{
	Use:   "get <backup_id>",
	Short: "Get backup details",
	Long:  `Get detailed information for a backup by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupGet,
}

func init() {
	backupCmd.AddCommand(backupGetCmd)
}

func runBackupGet(cmd *cobra.Command, args []string) error {
	backupID := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.BackupDetails(backupID)
	if err != nil {
		return err
	}

	if backup, ok := result["backup"].(map[string]interface{}); ok {
		output.PrintOutput(backup, textMode, "")
		return nil
	}

	output.PrintOutput(result, textMode, "")
	return nil
}
