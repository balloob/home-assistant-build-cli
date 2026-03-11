package cmd

import (
	"github.com/spf13/cobra"
)

var integrationCmd = &cobra.Command{
	Use:     "integration",
	Short:   "Manage integrations",
	Long:    `List, inspect, reload, enable, and disable Home Assistant integrations (config entries).`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(integrationCmd)
}
