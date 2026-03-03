package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var automationDeleteForce bool

var automationDeleteCmd = &cobra.Command{
	Use:     "delete <automation_id>",
	Short:   "Delete an automation",
	Long:    `Delete an automation from Home Assistant.`,
	GroupID: automationGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runAutomationDelete,
}

func init() {
	automationCmd.AddCommand(automationDeleteCmd)
	automationDeleteCmd.Flags().BoolVarP(&automationDeleteForce, "force", "f", false, "Skip confirmation")
}

func runAutomationDelete(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	configID, err := resolveAutomationConfigID(restClient, automationID)
	if err != nil {
		return err
	}

	if !automationDeleteForce && !textMode {
		fmt.Printf("Delete automation %s? [y/N]: ", automationID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	_, err = restClient.Delete("config/automation/config/" + configID)
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Automation %s deleted.", automationID))
	return nil
}
