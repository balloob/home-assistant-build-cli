package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	backupRestoreAgent         string
	backupRestorePassword      string
	backupRestoreAddons        []string
	backupRestoreFolders       []string
	backupRestoreDatabase      bool
	backupRestoreHomeAssistant bool
)

var backupRestoreCmd = &cobra.Command{
	Use:   "restore <backup_id>",
	Short: "Restore a backup",
	Long:  `Restore a backup by ID using a specific backup agent.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupRestore,
}

func init() {
	backupCmd.AddCommand(backupRestoreCmd)

	backupRestoreCmd.Flags().StringVar(&backupRestoreAgent, "agent", "", "Backup agent ID to restore from")
	backupRestoreCmd.Flags().StringVar(&backupRestorePassword, "password", "", "Backup password")
	backupRestoreCmd.Flags().StringSliceVar(&backupRestoreAddons, "restore-addon", nil, "Addon slug to restore (repeatable)")
	backupRestoreCmd.Flags().StringSliceVar(&backupRestoreFolders, "restore-folder", nil, "Folder to restore (repeatable)")
	backupRestoreCmd.Flags().BoolVar(&backupRestoreDatabase, "restore-database", false, "Restore the Home Assistant database")
	backupRestoreCmd.Flags().BoolVar(&backupRestoreHomeAssistant, "restore-homeassistant", false, "Restore Home Assistant core data")
	backupRestoreCmd.MarkFlagRequired("agent")
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
	backupID := args[0]
	textMode := getTextMode()

	params := map[string]interface{}{
		"backup_id": backupID,
		"agent_id":  backupRestoreAgent,
	}

	if backupRestorePassword != "" {
		params["password"] = backupRestorePassword
	}
	if len(backupRestoreAddons) > 0 {
		params["restore_addons"] = backupRestoreAddons
	}
	if len(backupRestoreFolders) > 0 {
		params["restore_folders"] = backupRestoreFolders
	}
	if cmd.Flags().Changed("restore-database") {
		params["restore_database"] = backupRestoreDatabase
	}
	if cmd.Flags().Changed("restore-homeassistant") {
		params["restore_homeassistant"] = backupRestoreHomeAssistant
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.BackupRestore(params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Backup restore started for '%s'.", backupID))
	return nil
}
