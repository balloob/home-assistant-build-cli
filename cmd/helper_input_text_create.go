package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperInputTextCreateIcon    string
	helperInputTextCreateInitial string
	helperInputTextCreateMin     int
	helperInputTextCreateMax     int
	helperInputTextCreatePattern string
	helperInputTextCreateMode    string
)

var helperInputTextCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new input text helper",
	Long:  `Create a new input text helper.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runHelperInputTextCreate,
}

func init() {
	helperInputTextParentCmd.AddCommand(helperInputTextCreateCmd)
	helperInputTextCreateCmd.Flags().StringVarP(&helperInputTextCreateIcon, "icon", "i", "", "Icon for the helper")
	helperInputTextCreateCmd.Flags().StringVar(&helperInputTextCreateInitial, "initial", "", "Initial value")
	helperInputTextCreateCmd.Flags().IntVar(&helperInputTextCreateMin, "min", 0, "Minimum length")
	helperInputTextCreateCmd.Flags().IntVar(&helperInputTextCreateMax, "max", 100, "Maximum length")
	helperInputTextCreateCmd.Flags().StringVar(&helperInputTextCreatePattern, "pattern", "", "Regex pattern for validation")
	helperInputTextCreateCmd.Flags().StringVar(&helperInputTextCreateMode, "mode", "text", "Display mode (text or password)")
}

func runHelperInputTextCreate(cmd *cobra.Command, args []string) error {
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

	if helperInputTextCreateIcon != "" {
		params["icon"] = helperInputTextCreateIcon
	}

	if helperInputTextCreateInitial != "" {
		params["initial"] = helperInputTextCreateInitial
	}

	if cmd.Flags().Changed("min") {
		params["min"] = helperInputTextCreateMin
	}

	if cmd.Flags().Changed("max") {
		params["max"] = helperInputTextCreateMax
	}

	if helperInputTextCreatePattern != "" {
		params["pattern"] = helperInputTextCreatePattern
	}

	if helperInputTextCreateMode != "" {
		params["mode"] = helperInputTextCreateMode
	}

	result, err := ws.HelperCreate("input_text", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input text '%s' created successfully.", name))
	return nil
}
