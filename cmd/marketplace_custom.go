package cmd

import "github.com/spf13/cobra"

var marketplaceCustomCmd = &cobra.Command{
	Use:   "custom",
	Short: "Manage custom Marketplace repositories",
	Long: `Add and remove GitHub repositories that are not in the Marketplace catalog.

Adding a repository needs a connected GitHub account and the accepted
Marketplace warning.`,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceCustomCmd)
}
