package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputNumberParentCmd = &cobra.Command{
	Use:   "helper-input-number",
	Short: "Manage input number helpers",
	Long:  `Create, list, and delete input number helpers.`,
}

func init() {
	rootCmd.AddCommand(helperInputNumberParentCmd)
}
