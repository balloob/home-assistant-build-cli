package cmd

import (
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Manage authentication with Home Assistant.`,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
