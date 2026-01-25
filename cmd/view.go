package cmd

import (
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Manage dashboard views",
	Long:  `Create, update, list, and delete views in a dashboard.`,
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
