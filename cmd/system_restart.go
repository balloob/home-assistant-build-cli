package cmd

import (
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

	if err := confirmAction(restartForce, "This will restart Home Assistant. Continue?", "restart system"); err != nil {
		return err
	}

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	if err := restClient.Restart(); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, "Restart initiated.", output.EnvelopeContext{Operation: "restart", ResourceType: "system"})
	return nil
}
