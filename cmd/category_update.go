package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	categoryUpdateName string
	categoryUpdateIcon string
)

var categoryUpdateCmd = &cobra.Command{
	Use:   "update <category_id>",
	Short: "Update a category",
	Long:  `Update an existing category's name or icon.`,
	Example: `  hab category update abc123 --name "Critical Alerts"
  hab category update abc123 --icon mdi:alert`,
	Args: cobra.ExactArgs(1),
	RunE: runCategoryUpdate,
}

func init() {
	categoryCmd.AddCommand(categoryUpdateCmd)
	categoryUpdateCmd.Flags().StringVar(&categoryUpdateName, "name", "", "New name for the category")
	categoryUpdateCmd.Flags().StringVar(&categoryUpdateIcon, "icon", "", "New icon for the category")
}

func runCategoryUpdate(cmd *cobra.Command, args []string) error {
	categoryID := args[0]
	textMode := getTextMode()

	params := map[string]interface{}{}
	if categoryUpdateName != "" {
		params["name"] = categoryUpdateName
	}
	if cmd.Flags().Changed("icon") {
		params["icon"] = categoryUpdateIcon
	}

	if len(params) == 0 {
		return fmt.Errorf("no update parameters provided")
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.CategoryRegistryUpdate(categoryID, params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Category '%s' updated.", categoryID))
	return nil
}
