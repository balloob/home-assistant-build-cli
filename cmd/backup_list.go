package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var backupListFlags *ListFlags

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	Long:  `List all available backups.`,
	RunE:  runBackupList,
}

func init() {
	backupCmd.AddCommand(backupListCmd)
	backupListFlags = RegisterListFlags(backupListCmd, "backup_id")
}

func runBackupList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("backup/info", nil)
	if err != nil {
		return err
	}

	var backups []interface{}
	if resultMap, ok := result.(map[string]interface{}); ok {
		if b, ok := resultMap["backups"].([]interface{}); ok {
			backups = b
		}
	}

	if backups == nil {
		output.PrintOutput(result, textMode, "")
		return nil
	}

	if backupListFlags.RenderCount(len(backups), textMode) {
		return nil
	}
	backups = backupListFlags.ApplyLimit(backups)
	if backupListFlags.RenderBrief(backups, textMode, "backup_id", "name") {
		return nil
	}

	output.PrintOutput(backups, textMode, "")
	return nil
}
