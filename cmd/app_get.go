package cmd

import (
	"time"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

func newAppGetCommand() *cobra.Command {
	var available bool
	cmd := &cobra.Command{
		Use:   "get <slug>",
		Short: "Get app details and configuration",
		Args:  appSlugArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := "/addons/" + args[0] + "/info"
			if available {
				endpoint = "/store/addons/" + args[0]
			}
			result, err := appRequest(cmd, "get", endpoint, nil, 30*time.Second)
			if err != nil {
				return err
			}
			output.PrintOutput(result, getTextMode(), "")
			return nil
		},
	}
	cmd.Flags().BoolVar(&available, "available", false, "Get store details for an app that is not installed")
	annotateAppCommand(cmd, "read")
	return cmd
}
