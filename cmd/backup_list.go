package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	Long:  `List all available backups.`,
	RunE:  runBackupList,
}

var (
	backupListCount bool
	backupListBrief bool
	backupListLimit int
)

func init() {
	backupCmd.AddCommand(backupListCmd)
	backupListCmd.Flags().BoolVarP(&backupListCount, "count", "c", false, "Return only the count of items")
	backupListCmd.Flags().BoolVarP(&backupListBrief, "brief", "b", false, "Return minimal fields (backup_id and name only)")
	backupListCmd.Flags().IntVarP(&backupListLimit, "limit", "n", 0, "Limit results to N items")
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

	// Handle count mode
	if backupListCount {
		output.PrintOutput(map[string]interface{}{"count": len(backups)}, textMode, "")
		return nil
	}

	// Apply limit
	if backupListLimit > 0 && len(backups) > backupListLimit {
		backups = backups[:backupListLimit]
	}

	// Handle brief mode
	if backupListBrief {
		var brief []map[string]interface{}
		for _, b := range backups {
			if backup, ok := b.(map[string]interface{}); ok {
				brief = append(brief, map[string]interface{}{
					"backup_id": backup["backup_id"],
					"name":      backup["name"],
				})
			}
		}
		output.PrintOutput(brief, textMode, "")
		return nil
	}

	output.PrintOutput(backups, textMode, "")
	return nil
}
