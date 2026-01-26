package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputDatetimeParentCmd = &cobra.Command{
	Use:     "input-datetime",
	Short:   "Manage input datetime helpers",
	Long:    `Create, list, and delete input datetime helpers.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperInputDatetimeParentCmd)
}
