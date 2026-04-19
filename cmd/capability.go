package cmd

import "github.com/spf13/cobra"

var capabilityCmd = &cobra.Command{
	Use:     "capability",
	Short:   "Inspect runtime capabilities",
	Long:    "Probe Home Assistant runtime capabilities for safe command/workflow planning.",
	GroupID: "start",
}

func init() {
	rootCmd.AddCommand(capabilityCmd)
	mergeSchemaAnnotation(capabilityCmd, SchemaAnnotation{
		SideEffect:   "meta",
		OutputMode:   "json_envelope",
		Capabilities: []string{"local"},
		ResourceType: "capability",
	})
}
