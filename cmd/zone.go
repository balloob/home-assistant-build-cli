package cmd

import (
	"github.com/spf13/cobra"
)

var zoneCmd = &cobra.Command{
	Use:     "zone",
	Short:   "Manage zones",
	Long:    `Create, update, and delete zones.`,
	GroupID: "registry",
}

func init() {
	rootCmd.AddCommand(zoneCmd)
}
