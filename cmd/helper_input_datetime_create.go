package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperInputDatetimeCreateIcon    string
	helperInputDatetimeCreateHasDate bool
	helperInputDatetimeCreateHasTime bool
	helperInputDatetimeCreateInitial string
)

var helperInputDatetimeCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new input datetime helper",
	Long: `Create a new input datetime helper.

At least one of --has-date or --has-time must be true.

Examples:
  hab helper-input-datetime create "My Date" --has-date
  hab helper-input-datetime create "My Time" --has-time
  hab helper-input-datetime create "My DateTime" --has-date --has-time`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperInputDatetimeCreate,
}

func init() {
	helperInputDatetimeParentCmd.AddCommand(helperInputDatetimeCreateCmd)
	helperInputDatetimeCreateCmd.Flags().StringVarP(&helperInputDatetimeCreateIcon, "icon", "i", "", "Icon for the helper (e.g., mdi:calendar)")
	helperInputDatetimeCreateCmd.Flags().BoolVar(&helperInputDatetimeCreateHasDate, "has-date", false, "Include date component")
	helperInputDatetimeCreateCmd.Flags().BoolVar(&helperInputDatetimeCreateHasTime, "has-time", false, "Include time component")
	helperInputDatetimeCreateCmd.Flags().StringVar(&helperInputDatetimeCreateInitial, "initial", "", "Initial value (format depends on has_date/has_time)")
}

func runHelperInputDatetimeCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Validate that at least one of has_date or has_time is true
	if !helperInputDatetimeCreateHasDate && !helperInputDatetimeCreateHasTime {
		return fmt.Errorf("at least one of --has-date or --has-time must be specified")
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
		"name":     name,
		"has_date": helperInputDatetimeCreateHasDate,
		"has_time": helperInputDatetimeCreateHasTime,
	}

	if helperInputDatetimeCreateIcon != "" {
		params["icon"] = helperInputDatetimeCreateIcon
	}

	if helperInputDatetimeCreateInitial != "" {
		params["initial"] = helperInputDatetimeCreateInitial
	}

	result, err := ws.HelperCreate("input_datetime", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input datetime '%s' created successfully.", name))
	return nil
}
