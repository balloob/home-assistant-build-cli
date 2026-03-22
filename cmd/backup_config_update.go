package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var backupConfigUpdateInput InputFlags

var backupConfigUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update backup configuration",
	Long:  `Update backup configuration using JSON or YAML input.`,
	RunE:  runBackupConfigUpdate,
}

func init() {
	backupConfigCmd.AddCommand(backupConfigUpdateCmd)
	backupConfigUpdateInput.Register(backupConfigUpdateCmd)
}

func runBackupConfigUpdate(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	if backupConfigUpdateInput.Data == "" && backupConfigUpdateInput.File == "" {
		return fmt.Errorf("backup configuration data is required (use --data or --file)")
	}

	params, err := backupConfigUpdateInput.Parse()
	if err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.BackupConfigUpdate(params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, "Backup configuration updated.")
	return nil
}
