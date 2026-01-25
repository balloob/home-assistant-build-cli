package cmd

import (
	"github.com/spf13/cobra"
)

var helperGroupParentCmd = &cobra.Command{
	Use:   "helper-group",
	Short: "Manage group helpers",
	Long:  `Create, list, and delete group helpers.`,
}

func init() {
	rootCmd.AddCommand(helperGroupParentCmd)
}
