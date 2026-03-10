package cmd

import (
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:     "template",
	Short:   "Work with Home Assistant templates",
	Long:    `Render and work with Jinja2 templates using the Home Assistant template engine.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(templateCmd)
}
