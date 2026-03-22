package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	networkConfigureAdapters []string
	networkConfigureApply    bool
)

var networkConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure network adapters",
	Long: `Configure the set of network adapters Home Assistant should use.

This command is potentially disruptive and requires --apply.`,
	RunE: runNetworkConfigure,
}

func init() {
	networkCmd.AddCommand(networkConfigureCmd)
	networkConfigureCmd.Flags().StringSliceVar(&networkConfigureAdapters, "adapters", nil, "Configured adapter names (repeatable or comma-separated)")
	networkConfigureCmd.Flags().BoolVar(&networkConfigureApply, "apply", false, "Apply network configuration changes")
}

func runNetworkConfigure(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	if len(networkConfigureAdapters) == 0 {
		return fmt.Errorf("at least one adapter is required (--adapters)")
	}
	if !networkConfigureApply {
		return fmt.Errorf("refusing to configure network without --apply")
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.NetworkConfigure(networkConfigureAdapters)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, "Network configuration updated.")
	return nil
}
