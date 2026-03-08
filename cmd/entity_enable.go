package cmd

import (
	"github.com/spf13/cobra"
)

var entityEnableCmd = &cobra.Command{
	Use:   "enable <entity_id>",
	Short: "Enable a disabled entity",
	Long:  `Enable an entity that was previously disabled.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEntityEnable,
}

func init() {
	entityCmd.AddCommand(entityEnableCmd)
}

func runEntityEnable(cmd *cobra.Command, args []string) error {
	return entitySetDisabled(args[0], false)
}
