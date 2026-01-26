package cmd

import (
	"github.com/spf13/cobra"
)

var actionCmd = &cobra.Command{
	Use:     "action",
	Short:   "Call actions (services)",
	Long:    `List and call Home Assistant actions.`,
	GroupID: "automation",
}

func init() {
	rootCmd.AddCommand(actionCmd)
}
