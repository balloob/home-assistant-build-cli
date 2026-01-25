package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperInputNumberCreateIcon              string
	helperInputNumberCreateMin               float64
	helperInputNumberCreateMax               float64
	helperInputNumberCreateStep              float64
	helperInputNumberCreateInitial           float64
	helperInputNumberCreateMode              string
	helperInputNumberCreateUnitOfMeasurement string
)

var helperInputNumberCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new input number helper",
	Long:  `Create a new input number helper with min/max range.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runHelperInputNumberCreate,
}

func init() {
	helperInputNumberParentCmd.AddCommand(helperInputNumberCreateCmd)
	helperInputNumberCreateCmd.Flags().StringVarP(&helperInputNumberCreateIcon, "icon", "i", "", "Icon for the helper")
	helperInputNumberCreateCmd.Flags().Float64Var(&helperInputNumberCreateMin, "min", 0, "Minimum value (required)")
	helperInputNumberCreateCmd.Flags().Float64Var(&helperInputNumberCreateMax, "max", 100, "Maximum value (required)")
	helperInputNumberCreateCmd.Flags().Float64Var(&helperInputNumberCreateStep, "step", 1, "Step value")
	helperInputNumberCreateCmd.Flags().Float64Var(&helperInputNumberCreateInitial, "initial", 0, "Initial value")
	helperInputNumberCreateCmd.Flags().StringVar(&helperInputNumberCreateMode, "mode", "slider", "Display mode (box or slider)")
	helperInputNumberCreateCmd.Flags().StringVar(&helperInputNumberCreateUnitOfMeasurement, "unit", "", "Unit of measurement")
	helperInputNumberCreateCmd.MarkFlagRequired("min")
	helperInputNumberCreateCmd.MarkFlagRequired("max")
}

func runHelperInputNumberCreate(cmd *cobra.Command, args []string) error {
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
		"min":  helperInputNumberCreateMin,
		"max":  helperInputNumberCreateMax,
	}

	if helperInputNumberCreateIcon != "" {
		params["icon"] = helperInputNumberCreateIcon
	}

	if cmd.Flags().Changed("step") {
		params["step"] = helperInputNumberCreateStep
	}

	if cmd.Flags().Changed("initial") {
		params["initial"] = helperInputNumberCreateInitial
	}

	if helperInputNumberCreateMode != "" {
		params["mode"] = helperInputNumberCreateMode
	}

	if helperInputNumberCreateUnitOfMeasurement != "" {
		params["unit_of_measurement"] = helperInputNumberCreateUnitOfMeasurement
	}

	result, err := ws.HelperCreate("input_number", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input number '%s' created successfully.", name))
	return nil
}
