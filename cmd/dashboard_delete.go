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

var dashboardDeleteForce bool

var dashboardDeleteCmd = &cobra.Command{
	Use:     "delete <dashboard_id>",
	Short:   "Delete a dashboard",
	Long:    `Delete a dashboard from Home Assistant.`,
	GroupID: dashboardGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runDashboardDelete,
}

func init() {
	dashboardCmd.AddCommand(dashboardDeleteCmd)
	dashboardDeleteCmd.Flags().BoolVarP(&dashboardDeleteForce, "force", "f", false, "Skip confirmation")
}

func runDashboardDelete(cmd *cobra.Command, args []string) error {
	dashboardID := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	if !dashboardDeleteForce && !textMode {
		fmt.Printf("Delete dashboard %s? [y/N]: ", dashboardID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{
		"dashboard_id": dashboardID,
	}

	_, err = ws.SendCommand("lovelace/dashboards/delete", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Dashboard '%s' deleted.", dashboardID))
	return nil
}
