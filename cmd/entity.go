package cmd

import (
	"github.com/spf13/cobra"
)

var entityCmd = &cobra.Command{
	Use:     "entity",
	Short:   "Manage entities",
	Long:    `List, get, search, and manage entities.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(entityCmd)
}
