package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputButtonParentCmd = &cobra.Command{
	Use:   "helper-input-button",
	Short: "Manage input button helpers",
	Long:  `Create, list, and delete input button helpers.`,
}

func init() {
	rootCmd.AddCommand(helperInputButtonParentCmd)
}
