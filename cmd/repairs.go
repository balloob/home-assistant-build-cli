package cmd

import (
	"github.com/spf13/cobra"
)

var repairsCmd = &cobra.Command{
	Use:     "repairs",
	Short:   "Manage Home Assistant repairs",
	Long:    `List and manage repair issues reported by Home Assistant.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(repairsCmd)
}
