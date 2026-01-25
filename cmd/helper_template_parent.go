package cmd

import (
	"github.com/spf13/cobra"
)

var helperTemplateParentCmd = &cobra.Command{
	Use:   "helper-template",
	Short: "Manage template entity helpers",
	Long: `Create, list, and delete template entity helpers.

Template entities allow you to create entities that derive their values from templates.
Supported domains: alarm_control_panel, binary_sensor, button, image, number, select, sensor, switch.`,
}

func init() {
	rootCmd.AddCommand(helperTemplateParentCmd)
}
