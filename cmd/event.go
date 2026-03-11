package cmd

import (
	"github.com/spf13/cobra"
)

var eventCmd = &cobra.Command{
	Use:     "event",
	Short:   "Manage Home Assistant events",
	Long:    `List registered event types and fire custom events on the Home Assistant event bus.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(eventCmd)
}
