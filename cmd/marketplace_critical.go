package cmd

import "github.com/spf13/cobra"

var marketplaceCriticalCmd = &cobra.Command{
	Use:   "critical",
	Short: "Manage critical Marketplace repositories",
	Long: `List repositories that the Marketplace catalog marks as critical, and
acknowledge them. The Marketplace removes a critical repository that is
installed and asks for a restart.`,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceCriticalCmd)
}
