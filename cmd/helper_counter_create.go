package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperCounterCreateIcon    string
	helperCounterCreateInitial int
	helperCounterCreateMinimum int
	helperCounterCreateMaximum int
	helperCounterCreateStep    int
	helperCounterCreateRestore bool
)

var helperCounterCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new counter helper",
	Long:  `Create a new counter helper that can be incremented/decremented.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runHelperCounterCreate,
}

func init() {
	helperCounterParentCmd.AddCommand(helperCounterCreateCmd)
	helperCounterCreateCmd.Flags().StringVarP(&helperCounterCreateIcon, "icon", "i", "", "Icon for the helper")
	helperCounterCreateCmd.Flags().IntVar(&helperCounterCreateInitial, "initial", 0, "Initial value")
	helperCounterCreateCmd.Flags().IntVar(&helperCounterCreateMinimum, "minimum", 0, "Minimum value")
	helperCounterCreateCmd.Flags().IntVar(&helperCounterCreateMaximum, "maximum", 0, "Maximum value (0 for no limit)")
	helperCounterCreateCmd.Flags().IntVar(&helperCounterCreateStep, "step", 1, "Step value for increment/decrement")
	helperCounterCreateCmd.Flags().BoolVar(&helperCounterCreateRestore, "restore", true, "Restore value after restart")
}

func runHelperCounterCreate(cmd *cobra.Command, args []string) error {
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

	if helperCounterCreateIcon != "" {
		params["icon"] = helperCounterCreateIcon
	}

	if cmd.Flags().Changed("initial") {
		params["initial"] = helperCounterCreateInitial
	}

	if cmd.Flags().Changed("minimum") {
		params["minimum"] = helperCounterCreateMinimum
	}

	if cmd.Flags().Changed("maximum") && helperCounterCreateMaximum != 0 {
		params["maximum"] = helperCounterCreateMaximum
	}

	if cmd.Flags().Changed("step") {
		params["step"] = helperCounterCreateStep
	}

	if cmd.Flags().Changed("restore") {
		params["restore"] = helperCounterCreateRestore
	}

	result, err := ws.HelperCreate("counter", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Counter '%s' created successfully.", name))
	return nil
}
