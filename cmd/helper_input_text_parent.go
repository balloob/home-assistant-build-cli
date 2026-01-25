package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputTextParentCmd = &cobra.Command{
	Use:   "helper-input-text",
	Short: "Manage input text helpers",
	Long:  `Create, list, and delete input text helpers.`,
}

func init() {
	rootCmd.AddCommand(helperInputTextParentCmd)
}
