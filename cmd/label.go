package cmd

import (
	"github.com/spf13/cobra"
)

var labelCmd = &cobra.Command{
	Use:     "label",
	Short:   "Manage labels",
	Long:    `Create, update, and delete labels.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(labelCmd)
}
