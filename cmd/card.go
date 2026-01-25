package cmd

import (
	"github.com/spf13/cobra"
)

var cardCmd = &cobra.Command{
	Use:   "card",
	Short: "Manage dashboard cards",
	Long:  `Create, update, list, and delete cards in a dashboard view or section.`,
}

func init() {
	rootCmd.AddCommand(cardCmd)
}
