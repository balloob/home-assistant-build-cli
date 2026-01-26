package cmd

import (
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage backups",
	Long:  `Create, restore, and manage backups.`,
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
