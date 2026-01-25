package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperTimerCreateIcon     string
	helperTimerCreateDuration string
	helperTimerCreateRestore  bool
)

var helperTimerCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new timer helper",
	Long: `Create a new timer helper that counts down.

Duration can be specified in format: HH:MM:SS or just seconds (e.g., "01:30:00" or "5400").`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperTimerCreate,
}

func init() {
	helperTimerParentCmd.AddCommand(helperTimerCreateCmd)
	helperTimerCreateCmd.Flags().StringVarP(&helperTimerCreateIcon, "icon", "i", "", "Icon for the helper")
	helperTimerCreateCmd.Flags().StringVarP(&helperTimerCreateDuration, "duration", "d", "", "Default duration (e.g., 00:05:00 for 5 minutes)")
	helperTimerCreateCmd.Flags().BoolVar(&helperTimerCreateRestore, "restore", true, "Restore timer state after restart")
}

func runHelperTimerCreate(cmd *cobra.Command, args []string) error {
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

	if helperTimerCreateIcon != "" {
		params["icon"] = helperTimerCreateIcon
	}

	if helperTimerCreateDuration != "" {
		params["duration"] = helperTimerCreateDuration
	}

	if cmd.Flags().Changed("restore") {
		params["restore"] = helperTimerCreateRestore
	}

	result, err := ws.HelperCreate("timer", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Timer '%s' created successfully.", name))
	return nil
}
