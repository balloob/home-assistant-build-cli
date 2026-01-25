package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
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
		client.PrintOutput(result, textMode, "")
		return nil
	}

	// Handle count mode
	if backupListCount {
		client.PrintOutput(map[string]interface{}{"count": len(backups)}, textMode, "")
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
		client.PrintOutput(brief, textMode, "")
		return nil
	}

	client.PrintOutput(backups, textMode, "")
	return nil
}
