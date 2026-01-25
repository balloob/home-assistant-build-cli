package cmd

import (
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for items and relationships",
	Long:  `Search commands for finding related items in Home Assistant.`,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
