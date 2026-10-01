package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var marketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "Manage Marketplace repositories",
	Long: `Browse, install, and manage community integrations, dashboard cards,
themes, and templates from the Home Assistant Marketplace.

Requires Home Assistant 2026.11 or later.

Commands that take <repository> accept the Marketplace repository ID or the
GitHub full name (owner/repo).

Before the first install, a user must accept the Marketplace warning with
'hab marketplace accept-warning'. Adding custom repositories also needs a
connected GitHub account ('hab marketplace github-connect').`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(marketplaceCmd)
}

type marketplaceCommandSender interface {
	SendCommand(string, map[string]interface{}) (interface{}, error)
}

// resolveMarketplaceRepository returns the Marketplace ID for ref. A ref that
// contains a slash is a GitHub full name and is looked up in the repository list.
func resolveMarketplaceRepository(ws marketplaceCommandSender, ref string) (string, error) {
	if !strings.Contains(ref, "/") {
		return ref, nil
	}
	result, err := ws.SendCommand("marketplace/repositories/list", nil)
	if err != nil {
		return "", err
	}
	repositories, _ := result.([]interface{})
	for _, r := range repositories {
		repository, _ := r.(map[string]interface{})
		fullName, _ := repository["full_name"].(string)
		if strings.EqualFold(fullName, ref) {
			id, _ := repository["id"].(string)
			return id, nil
		}
	}
	return "", fmt.Errorf("repository %s not found in the Marketplace; run 'hab marketplace list --search %s'", ref, ref)
}
