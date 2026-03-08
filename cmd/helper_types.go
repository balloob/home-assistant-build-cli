package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var helperTypesCmd = &cobra.Command{
	Use:     "types",
	Short:   "List available helper types",
	Long:    `List all available helper types that can be created.`,
	GroupID: helperGroupCommands,
	RunE:    runHelperTypes,
}

func init() {
	helperCmd.AddCommand(helperTypesCmd)
}

func runHelperTypes(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	result := make([]interface{}, len(helperTypeRegistry))
	for i, def := range helperTypeRegistry {
		result[i] = map[string]interface{}{
			"type":        def.TypeName,
			"description": def.TypeDescription,
			"parameters":  def.CreateParams,
		}
	}

	output.PrintOutput(result, textMode, "")
	return nil
}
