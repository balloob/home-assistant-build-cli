package cmd

import (
	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Manage system",
	Long: `View system info, check config, restart, and more.

For workflow guidance, run 'hab guide operations'.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(systemCmd)
}
