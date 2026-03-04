package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var restartForce bool

var systemRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Home Assistant",
	Long:  `Restart the Home Assistant core.`,
	RunE:  runSystemRestart,
}

func init() {
	systemCmd.AddCommand(systemRestartCmd)
	systemRestartCmd.Flags().BoolVarP(&restartForce, "force", "f", false, "Skip confirmation")
}

func runSystemRestart(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	if !confirmAction(restartForce, textMode, "This will restart Home Assistant. Continue?") {
		fmt.Println("Cancelled.")
		return nil
	}

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	if err := restClient.Restart(); err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, "Restart initiated.")
	return nil
}
