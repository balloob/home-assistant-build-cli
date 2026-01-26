package cmd

import (
	"github.com/spf13/cobra"
)

var threadCmd = &cobra.Command{
	Use:     "thread",
	Short:   "Manage Thread credentials",
	Long:    `List, add, and manage Thread network credentials.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(threadCmd)
}
