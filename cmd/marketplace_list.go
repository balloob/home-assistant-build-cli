package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	marketplaceListCategories []string
	marketplaceListInstalled  bool
	marketplaceListSearch     string
	marketplaceListFlags      *ListFlags
)

var marketplaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Marketplace repositories",
	Long: `List repositories in the Marketplace catalog and custom repositories.

Categories are integration, plugin (dashboard cards and other frontend
resources), theme, and template.`,
	Example: `  hab marketplace list --installed
  hab marketplace list --category plugin --search mushroom --brief
  hab marketplace list --category integration --count`,
	RunE: runMarketplaceList,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceListCmd)
	marketplaceListCmd.Flags().StringSliceVar(&marketplaceListCategories, "category", nil, "Filter by category: integration, plugin, theme, or template (repeatable)")
	marketplaceListCmd.Flags().BoolVar(&marketplaceListInstalled, "installed", false, "Only list installed repositories")
	marketplaceListCmd.Flags().StringVarP(&marketplaceListSearch, "search", "s", "", "Filter by text in the name, full name, description, or topics")
	marketplaceListFlags = RegisterListFlags(marketplaceListCmd, "full_name")
}

func runMarketplaceList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{}
	if len(marketplaceListCategories) > 0 {
		params["categories"] = marketplaceListCategories
	}
	result, err := ws.SendCommand("marketplace/repositories/list", params)
	if err != nil {
		return err
	}

	repositories, _ := result.([]interface{})
	search := strings.ToLower(marketplaceListSearch)
	filtered := make([]interface{}, 0, len(repositories))
	for _, r := range repositories {
		repository, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		if installed, _ := repository["installed"].(bool); marketplaceListInstalled && !installed {
			continue
		}
		if search != "" && !marketplaceRepositoryMatches(repository, search) {
			continue
		}
		filtered = append(filtered, repository)
	}

	if marketplaceListFlags.RenderCount(len(filtered), textMode) {
		return nil
	}
	filtered = marketplaceListFlags.ApplyLimit(filtered)
	if marketplaceListFlags.RenderBrief(filtered, textMode, "full_name", "name") {
		return nil
	}

	if textMode {
		filtered = marketplaceTextColumns(filtered)
	}
	output.PrintOutput(filtered, textMode, "")
	return nil
}

// marketplaceTextColumns keeps the columns that identify a repository and its
// state. The text table shows at most six columns in alphabetical order, which
// would otherwise hide the name.
func marketplaceTextColumns(repositories []interface{}) []interface{} {
	rows := make([]interface{}, 0, len(repositories))
	for _, r := range repositories {
		repository, _ := r.(map[string]interface{})
		row := map[string]interface{}{}
		for _, key := range []string{"full_name", "name", "category", "installed_version", "available_version", "status"} {
			row[key] = repository[key]
		}
		rows = append(rows, row)
	}
	return rows
}

// marketplaceRepositoryMatches reports whether search, in lower case, occurs in
// the name, full name, description, or a topic of the repository.
func marketplaceRepositoryMatches(repository map[string]interface{}, search string) bool {
	for _, key := range []string{"name", "full_name", "description"} {
		if value, _ := repository[key].(string); strings.Contains(strings.ToLower(value), search) {
			return true
		}
	}
	topics, _ := repository["topics"].([]interface{})
	for _, t := range topics {
		if topic, _ := t.(string); strings.Contains(strings.ToLower(topic), search) {
			return true
		}
	}
	return false
}
