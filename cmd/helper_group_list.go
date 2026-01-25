package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperGroupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	Long:  `List all group helpers created via config entries.`,
	RunE:  runHelperGroupList,
}

var (
	helperGroupListCount bool
	helperGroupListBrief bool
	helperGroupListLimit int
)

func init() {
	helperGroupParentCmd.AddCommand(helperGroupListCmd)
	helperGroupListCmd.Flags().BoolVarP(&helperGroupListCount, "count", "c", false, "Return only the count of items")
	helperGroupListCmd.Flags().BoolVarP(&helperGroupListBrief, "brief", "b", false, "Return minimal fields (entry_id and title only)")
	helperGroupListCmd.Flags().IntVarP(&helperGroupListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperGroupList(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	// Get config entries for the "group" domain
	entries, err := ws.ConfigEntriesList("group")
	if err != nil {
		return err
	}

	var result []map[string]interface{}
	for _, e := range entries {
		entry, ok := e.(map[string]interface{})
		if !ok {
			continue
		}

		item := map[string]interface{}{
			"entry_id": entry["entry_id"],
			"title":    entry["title"],
		}

		// Extract domain from entry if available
		if domain, ok := entry["domain"].(string); ok {
			item["domain"] = domain
		}

		result = append(result, item)
	}

	// Handle count mode
	if helperGroupListCount {
		client.PrintOutput(map[string]interface{}{"count": len(result)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperGroupListLimit > 0 && len(result) > helperGroupListLimit {
		result = result[:helperGroupListLimit]
	}

	// Handle brief mode
	if helperGroupListBrief {
		var brief []map[string]interface{}
		for _, item := range result {
			brief = append(brief, map[string]interface{}{
				"entry_id": item["entry_id"],
				"title":    item["title"],
			})
		}
		client.PrintOutput(brief, textMode, "")
		return nil
	}

	client.PrintOutput(result, textMode, "")
	return nil
}
