package cmd

import "github.com/spf13/cobra"

var networkCmd = &cobra.Command{
	Use:     "network",
	Short:   "Manage network settings",
	Long:    `Inspect and configure network adapters and URLs.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(networkCmd)
}
