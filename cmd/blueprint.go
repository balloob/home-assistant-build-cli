package cmd

import (
	"github.com/spf13/cobra"
)

var blueprintCmd = &cobra.Command{
	Use:     "blueprint",
	Short:   "Manage blueprints",
	Long:    `List, import, and manage blueprints.`,
	GroupID: "automation",
}

func init() {
	rootCmd.AddCommand(blueprintCmd)
}
