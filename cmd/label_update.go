package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	labelUpdateName        string
	labelUpdateIcon        string
	labelUpdateColor       string
	labelUpdateDescription string
)

var labelUpdateCmd = &cobra.Command{
	Use:   "update <label_id>",
	Short: "Update a label",
	Long:  `Update an existing label.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelUpdate,
}

func init() {
	labelCmd.AddCommand(labelUpdateCmd)
	labelUpdateCmd.Flags().StringVar(&labelUpdateName, "name", "", "New name for the label")
	labelUpdateCmd.Flags().StringVar(&labelUpdateIcon, "icon", "", "Icon for the label")
	labelUpdateCmd.Flags().StringVar(&labelUpdateColor, "color", "", "Color for the label")
	labelUpdateCmd.Flags().StringVar(&labelUpdateDescription, "description", "", "Description of the label")
}

func runLabelUpdate(cmd *cobra.Command, args []string) error {
	labelID := args[0]
	textMode := getTextMode()

	params := make(map[string]interface{})
	if labelUpdateName != "" {
		params["name"] = labelUpdateName
	}
	if labelUpdateIcon != "" {
		params["icon"] = labelUpdateIcon
	}
	if labelUpdateColor != "" {
		params["color"] = labelUpdateColor
	}
	if labelUpdateDescription != "" {
		params["description"] = labelUpdateDescription
	}

	if len(params) == 0 {
		return fmt.Errorf("no update parameters provided")
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.LabelRegistryUpdate(labelID, params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Label '%s' updated.", labelID))
	return nil
}
