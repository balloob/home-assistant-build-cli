package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputDatetimeParentCmd = &cobra.Command{
	Use:   "helper-input-datetime",
	Short: "Manage input datetime helpers",
	Long:  `Create, list, and delete input datetime helpers.`,
}

func init() {
	rootCmd.AddCommand(helperInputDatetimeParentCmd)
}
