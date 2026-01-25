package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperInputSelectCreateIcon    string
	helperInputSelectCreateOptions []string
	helperInputSelectCreateInitial string
)

var helperInputSelectCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new input select helper",
	Long:  `Create a new input select (dropdown) helper.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runHelperInputSelectCreate,
}

func init() {
	helperInputSelectParentCmd.AddCommand(helperInputSelectCreateCmd)
	helperInputSelectCreateCmd.Flags().StringVarP(&helperInputSelectCreateIcon, "icon", "i", "", "Icon for the helper")
	helperInputSelectCreateCmd.Flags().StringSliceVarP(&helperInputSelectCreateOptions, "options", "o", nil, "Options for the dropdown (required)")
	helperInputSelectCreateCmd.Flags().StringVar(&helperInputSelectCreateInitial, "initial", "", "Initial selected option")
	helperInputSelectCreateCmd.MarkFlagRequired("options")
}

func runHelperInputSelectCreate(cmd *cobra.Command, args []string) error {
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
		"name":    name,
		"options": helperInputSelectCreateOptions,
	}

	if helperInputSelectCreateIcon != "" {
		params["icon"] = helperInputSelectCreateIcon
	}

	if helperInputSelectCreateInitial != "" {
		params["initial"] = helperInputSelectCreateInitial
	}

	result, err := ws.HelperCreate("input_select", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input select '%s' created successfully.", name))
	return nil
}
