package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	categoryDeleteScope string
	categoryDeleteForce bool
)

var categoryDeleteCmd = &cobra.Command{
	Use:   "delete <category_id>",
	Short: "Delete a category",
	Long:  `Delete a category from a given scope.`,
	Example: `  hab category delete abc123 --scope automation
  hab category delete abc123 --scope automation --force`,
	Args: cobra.ExactArgs(1),
	RunE: runCategoryDelete,
}

func init() {
	categoryCmd.AddCommand(categoryDeleteCmd)
	categoryDeleteCmd.Flags().StringVar(&categoryDeleteScope, "scope", "", "Scope of the category: automation, script, scene, helpers (required)")
	categoryDeleteCmd.Flags().BoolVarP(&categoryDeleteForce, "force", "f", false, "Skip confirmation")
	categoryDeleteCmd.MarkFlagRequired("scope")
}

func runCategoryDelete(cmd *cobra.Command, args []string) error {
	categoryID := args[0]

	if !validCategoryScopes[categoryDeleteScope] {
		return fmt.Errorf("invalid scope '%s'. Valid values: automation, script, scene, helpers", categoryDeleteScope)
	}
	textMode := getTextMode()

	if !confirmAction(categoryDeleteForce, textMode, fmt.Sprintf("Delete category %s from scope %s?", categoryID, categoryDeleteScope)) {
		fmt.Println("Cancelled.")
		return nil
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.CategoryRegistryDelete(categoryDeleteScope, categoryID); err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Category '%s' deleted.", categoryID))
	return nil
}
