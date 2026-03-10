package cmd

import (
	"github.com/spf13/cobra"
)

var personCmd = &cobra.Command{
	Use:     "person",
	Short:   "Manage persons",
	Long:    `Create, update, and delete person entries for presence tracking.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(personCmd)
}
