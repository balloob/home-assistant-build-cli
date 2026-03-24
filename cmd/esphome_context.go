package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/config"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var esphomeContextCmd = &cobra.Command{
	Use:     "context",
	Aliases: []string{"device"},
	Short:   "Manage saved ESPHome device context",
	Long:    `Save, inspect, or clear the default ESPHome configuration used by follow-up commands.`,
}

var esphomeContextUseCmd = &cobra.Command{
	Use:   "use <configuration>",
	Short: "Save the active ESPHome configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()
		ctx := &config.ESPHomeContext{Configuration: args[0]}
		if err := config.SaveESPHomeContext(viper.GetString("config"), ctx); err != nil {
			return err
		}
		output.PrintSuccess(map[string]any{"configuration": ctx.Configuration}, textMode, fmt.Sprintf("ESPHome context set to %s.", ctx.Configuration))
		return nil
	},
}

var esphomeContextShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the saved ESPHome configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()
		ctx, err := config.LoadESPHomeContext(viper.GetString("config"))
		if err != nil {
			return err
		}
		if ctx == nil {
			output.PrintOutput(map[string]any{"configuration": nil}, textMode, "No ESPHome context selected.")
			return nil
		}
		output.PrintOutput(map[string]any{"configuration": ctx.Configuration}, textMode, "")
		return nil
	},
}

var esphomeContextClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the saved ESPHome configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()
		if err := config.ClearESPHomeContext(viper.GetString("config")); err != nil {
			return err
		}
		output.PrintSuccess(nil, textMode, "ESPHome context cleared.")
		return nil
	},
}

func init() {
	esphomeCmd.AddCommand(esphomeContextCmd)
	esphomeContextCmd.AddCommand(esphomeContextUseCmd)
	esphomeContextCmd.AddCommand(esphomeContextShowCmd)
	esphomeContextCmd.AddCommand(esphomeContextClearCmd)
}
