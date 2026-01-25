package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperThresholdListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all threshold sensors",
	Long:  `List all threshold binary sensor helpers created via config entries.`,
	RunE:  runHelperThresholdList,
}

var (
	helperThresholdListCount bool
	helperThresholdListBrief bool
	helperThresholdListLimit int
)

func init() {
	helperThresholdParentCmd.AddCommand(helperThresholdListCmd)
	helperThresholdListCmd.Flags().BoolVarP(&helperThresholdListCount, "count", "c", false, "Return only the count of items")
	helperThresholdListCmd.Flags().BoolVarP(&helperThresholdListBrief, "brief", "b", false, "Return minimal fields (entry_id and title only)")
	helperThresholdListCmd.Flags().IntVarP(&helperThresholdListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperThresholdList(cmd *cobra.Command, args []string) error {
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

	entries, err := ws.ConfigEntriesList("threshold")
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
	if helperThresholdListCount {
		client.PrintOutput(map[string]interface{}{"count": len(result)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperThresholdListLimit > 0 && len(result) > helperThresholdListLimit {
		result = result[:helperThresholdListLimit]
	}

	// Handle brief mode
	if helperThresholdListBrief {
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
