package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var backupCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new backup",
	Long:  `Create a new backup of Home Assistant.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBackupCreate,
}

func init() {
	backupCmd.AddCommand(backupCreateCmd)
}

func runBackupCreate(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{}
	if len(args) > 0 {
		params["name"] = args[0]
	}

	result, err := ws.SendCommand("backup/generate", params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, "Backup creation initiated.")
	return nil
}
