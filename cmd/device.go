package cmd

import (
	"github.com/spf13/cobra"
)

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Manage devices",
	Long:  `List, view, and manage devices.`,
}

func init() {
	rootCmd.AddCommand(deviceCmd)
}
