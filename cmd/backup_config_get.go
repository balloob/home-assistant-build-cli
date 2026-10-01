package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var backupConfigGetShowPassword bool

var backupConfigGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get backup configuration",
	Long: `Get current backup configuration.

The backup encryption password is left out of the output. The output shows
create_backup.password_set instead. Use --show-password to include it.`,
	RunE: runBackupConfigGet,
}

func init() {
	backupConfigCmd.AddCommand(backupConfigGetCmd)
	backupConfigGetCmd.Flags().BoolVar(&backupConfigGetShowPassword, "show-password", false, "Include the backup encryption password in the output")
}

func runBackupConfigGet(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.BackupConfigInfo()
	if err != nil {
		return err
	}

	if cfg, ok := result["config"].(map[string]interface{}); ok {
		if !backupConfigGetShowPassword {
			redactBackupPassword(cfg)
		}
		output.PrintOutput(cfg, textMode, "")
		return nil
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

// redactBackupPassword replaces create_backup.password with a password_set
// flag. The key is removed rather than masked, so a config copied into
// 'hab backup config update' cannot overwrite the password with a placeholder.
func redactBackupPassword(cfg map[string]interface{}) {
	createBackup, ok := cfg["create_backup"].(map[string]interface{})
	if !ok {
		return
	}
	password, _ := createBackup["password"].(string)
	delete(createBackup, "password")
	createBackup["password_set"] = password != ""
}
