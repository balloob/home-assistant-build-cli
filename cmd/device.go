package cmd

import (
	"github.com/spf13/cobra"
)

var deviceCmd = &cobra.Command{
	Use:     "device",
	Short:   "Manage devices",
	Long:    `List, view, and manage devices.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(deviceCmd)
}
