package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperScheduleCreateIcon string
)

var helperScheduleCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new schedule helper",
	Long: `Create a new schedule helper for time-based automation.

Schedule helpers allow you to define time blocks for each day of the week.
After creation, use the Home Assistant UI to configure the schedule blocks.`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperScheduleCreate,
}

func init() {
	helperScheduleParentCmd.AddCommand(helperScheduleCreateCmd)
	helperScheduleCreateCmd.Flags().StringVarP(&helperScheduleCreateIcon, "icon", "i", "", "Icon for the helper (e.g., mdi:calendar-clock)")
}

func runHelperScheduleCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

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
		"name": name,
	}

	if helperScheduleCreateIcon != "" {
		params["icon"] = helperScheduleCreateIcon
	}

	result, err := ws.HelperCreate("schedule", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Schedule '%s' created successfully.", name))
	return nil
}
