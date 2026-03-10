package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var categoryListScope string
var categoryListFlags *ListFlags

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List categories",
	Long:  `List all categories for a given scope (automation, script, scene, or helpers).`,
	Example: `  hab category list --scope automation
  hab category list --scope helpers`,
	RunE: runCategoryList,
}

func init() {
	categoryCmd.AddCommand(categoryListCmd)
	categoryListCmd.Flags().StringVar(&categoryListScope, "scope", "", "Scope to list categories for: automation, script, scene, helpers (required)")
	categoryListCmd.MarkFlagRequired("scope")
	categoryListFlags = RegisterListFlags(categoryListCmd, "category_id")
}

func runCategoryList(cmd *cobra.Command, args []string) error {
	if !validCategoryScopes[categoryListScope] {
		return fmt.Errorf("invalid scope '%s'. Valid values: automation, script, scene, helpers", categoryListScope)
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	categories, err := ws.CategoryRegistryList(categoryListScope)
	if err != nil {
		return err
	}

	if categoryListFlags.RenderCount(len(categories), textMode) {
		return nil
	}
	categories = categoryListFlags.ApplyLimit(categories)
	if categoryListFlags.RenderBrief(categories, textMode, "category_id", "name") {
		return nil
	}

	if textMode {
		if len(categories) == 0 {
			fmt.Printf("No categories for scope '%s'.\n", categoryListScope)
			return nil
		}
		for _, c := range categories {
			if m, ok := c.(map[string]interface{}); ok {
				name, _ := m["name"].(string)
				id, _ := m["category_id"].(string)
				fmt.Printf("%s (%s)\n", name, id)
			}
		}
	} else {
		output.PrintOutput(categories, false, "")
	}
	return nil
}
