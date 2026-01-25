package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputSelectParentCmd = &cobra.Command{
	Use:   "helper-input-select",
	Short: "Manage input select helpers",
	Long:  `Create, list, and delete input select (dropdown) helpers.`,
}

func init() {
	rootCmd.AddCommand(helperInputSelectParentCmd)
}
