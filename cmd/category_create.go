package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	categoryCreateScope string
	categoryCreateIcon  string
)

var categoryCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new category",
	Long:  `Create a new category in a given scope.`,
	Example: `  hab category create "Notifications" --scope automation
  hab category create "Security" --scope automation --icon mdi:shield
  hab category create "Climate" --scope helpers`,
	Args: cobra.ExactArgs(1),
	RunE: runCategoryCreate,
}

func init() {
	categoryCmd.AddCommand(categoryCreateCmd)
	categoryCreateCmd.Flags().StringVar(&categoryCreateScope, "scope", "", "Scope for the category: automation, script, scene, helpers (required)")
	categoryCreateCmd.Flags().StringVar(&categoryCreateIcon, "icon", "", "Icon for the category (e.g. mdi:bell)")
	categoryCreateCmd.MarkFlagRequired("scope")
}

func runCategoryCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	if !validCategoryScopes[categoryCreateScope] {
		return fmt.Errorf("invalid scope '%s'. Valid values: automation, script, scene, helpers", categoryCreateScope)
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{}
	if categoryCreateIcon != "" {
		params["icon"] = categoryCreateIcon
	}

	result, err := ws.CategoryRegistryCreate(categoryCreateScope, name, params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Category '%s' created in scope '%s'.", name, categoryCreateScope))
	return nil
}
