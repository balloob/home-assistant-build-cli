package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var blueprintDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a blueprint",
	Long:  `Delete a blueprint by its path. Use --domain to specify the domain (default: automation).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBlueprintDelete,
}

func init() {
	blueprintCmd.AddCommand(blueprintDeleteCmd)
	blueprintDeleteCmd.Flags().String("domain", "automation", "Domain of the blueprint (automation/script)")
	blueprintDeleteCmd.Flags().Bool("force", false, "Skip confirmation")
}

func runBlueprintDelete(cmd *cobra.Command, args []string) error {
	path := args[0]
	textMode := getTextMode()
	domain, _ := cmd.Flags().GetString("domain")

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("blueprint/delete", map[string]interface{}{
		"domain": domain,
		"path":   path,
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Blueprint %s deleted successfully.", path))
	return nil
}
