package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var blueprintListCmd = &cobra.Command{
	Use:   "list [domain]",
	Short: "List blueprints",
	Long:  `List all blueprints, optionally filtered by domain (automation/script).`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBlueprintList,
}

func init() {
	blueprintCmd.AddCommand(blueprintListCmd)
}

func runBlueprintList(cmd *cobra.Command, args []string) error {
	domain := "automation"
	if len(args) > 0 {
		domain = args[0]
	}

	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("blueprint/list", map[string]interface{}{
		"domain": domain,
	})
	if err != nil {
		return err
	}

	client.PrintOutput(result, textMode, "")
	return nil
}
