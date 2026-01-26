package cmd

import (
	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:     "system",
	Short:   "Manage system",
	Long:    `View system info, check config, restart, and more.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(systemCmd)
}
