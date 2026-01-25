package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperMinMaxListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all min/max sensors",
	Long:  `List all min/max sensor helpers created via config entries.`,
	RunE:  runHelperMinMaxList,
}

var (
	helperMinMaxListCount bool
	helperMinMaxListBrief bool
	helperMinMaxListLimit int
)

func init() {
	helperMinMaxParentCmd.AddCommand(helperMinMaxListCmd)
	helperMinMaxListCmd.Flags().BoolVarP(&helperMinMaxListCount, "count", "c", false, "Return only the count of items")
	helperMinMaxListCmd.Flags().BoolVarP(&helperMinMaxListBrief, "brief", "b", false, "Return minimal fields (entry_id and title only)")
	helperMinMaxListCmd.Flags().IntVarP(&helperMinMaxListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperMinMaxList(cmd *cobra.Command, args []string) error {
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

	entries, err := ws.ConfigEntriesList("min_max")
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

		if domain, ok := entry["domain"].(string); ok {
			item["domain"] = domain
		}

		result = append(result, item)
	}

	// Handle count mode
	if helperMinMaxListCount {
		client.PrintOutput(map[string]interface{}{"count": len(result)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperMinMaxListLimit > 0 && len(result) > helperMinMaxListLimit {
		result = result[:helperMinMaxListLimit]
	}

	// Handle brief mode
	if helperMinMaxListBrief {
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
