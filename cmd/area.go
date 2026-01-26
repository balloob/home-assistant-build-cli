package cmd

import (
	"github.com/spf13/cobra"
)

var areaCmd = &cobra.Command{
	Use:     "area",
	Short:   "Manage areas",
	Long:    `Create, update, and delete areas.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(areaCmd)
}
