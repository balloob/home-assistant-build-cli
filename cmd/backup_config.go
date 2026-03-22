package cmd

import "github.com/spf13/cobra"

var backupConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage backup configuration",
	Long:  `Get and update backup configuration settings.`,
}

func init() {
	backupCmd.AddCommand(backupConfigCmd)
}
