package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var backupConfigGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get backup configuration",
	Long:  `Get current backup configuration.`,
	RunE:  runBackupConfigGet,
}

func init() {
	backupConfigCmd.AddCommand(backupConfigGetCmd)
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
		output.PrintOutput(cfg, textMode, "")
		return nil
	}

	output.PrintOutput(result, textMode, "")
	return nil
}
