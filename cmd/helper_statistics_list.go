package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperStatisticsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all statistics sensors",
	Long:  `List all statistics sensor helpers created via config entries.`,
	RunE:  runHelperStatisticsList,
}

var (
	helperStatisticsListCount bool
	helperStatisticsListBrief bool
	helperStatisticsListLimit int
)

func init() {
	helperStatisticsParentCmd.AddCommand(helperStatisticsListCmd)
	helperStatisticsListCmd.Flags().BoolVarP(&helperStatisticsListCount, "count", "c", false, "Return only the count of items")
	helperStatisticsListCmd.Flags().BoolVarP(&helperStatisticsListBrief, "brief", "b", false, "Return minimal fields (entry_id and title only)")
	helperStatisticsListCmd.Flags().IntVarP(&helperStatisticsListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperStatisticsList(cmd *cobra.Command, args []string) error {
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

	entries, err := ws.ConfigEntriesList("statistics")
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
	if helperStatisticsListCount {
		client.PrintOutput(map[string]interface{}{"count": len(result)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperStatisticsListLimit > 0 && len(result) > helperStatisticsListLimit {
		result = result[:helperStatisticsListLimit]
	}

	// Handle brief mode
	if helperStatisticsListBrief {
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
