package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var (
	labelCreateIcon        string
	labelCreateColor       string
	labelCreateDescription string
)

var labelCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new label",
	Long:  `Create a new label in Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelCreate,
}

func init() {
	labelCmd.AddCommand(labelCreateCmd)
	labelCreateCmd.Flags().StringVar(&labelCreateIcon, "icon", "", "Icon for the label")
	labelCreateCmd.Flags().StringVar(&labelCreateColor, "color", "", "Color for the label")
	labelCreateCmd.Flags().StringVar(&labelCreateDescription, "description", "", "Description of the label")
}

func runLabelCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := make(map[string]interface{})
	if labelCreateIcon != "" {
		params["icon"] = labelCreateIcon
	}
	if labelCreateColor != "" {
		params["color"] = labelCreateColor
	}
	if labelCreateDescription != "" {
		params["description"] = labelCreateDescription
	}

	result, err := ws.LabelRegistryCreate(name, params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Label '%s' created.", name))
	return nil
}
