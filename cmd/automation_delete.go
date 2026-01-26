package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	// Strip "automation." prefix if provided - API expects just the ID
	automationID = strings.TrimPrefix(automationID, "automation.")

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

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

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	_, err = restClient.Delete("config/automation/config/" + automationID)
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Automation %s deleted.", automationID))
	return nil
}
