package cmd

import (
	"github.com/spf13/cobra"
)

var floorCmd = &cobra.Command{
	Use:     "floor",
	Short:   "Manage floors",
	Long:    `Create, update, and delete floors.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(floorCmd)
}
