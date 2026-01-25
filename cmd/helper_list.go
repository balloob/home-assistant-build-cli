package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperListCmd = &cobra.Command{
	Use:   "list [type]",
	Short: "List helper entities",
	Long:  `List all helper entities, optionally filtered by type.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runHelperList,
}

var (
	helperListCount bool
	helperListBrief bool
	helperListLimit int
)

func init() {
	helperCmd.AddCommand(helperListCmd)
	helperListCmd.Flags().BoolVarP(&helperListCount, "count", "c", false, "Return only the count of items")
	helperListCmd.Flags().BoolVarP(&helperListBrief, "brief", "b", false, "Return minimal fields (entity_id and name only)")
	helperListCmd.Flags().IntVarP(&helperListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperList(cmd *cobra.Command, args []string) error {
	var filterType string
	if len(args) > 0 {
		filterType = args[0]
	}

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

	entities, err := ws.EntityRegistryList()
	if err != nil {
		return err
	}

	// Helper domains
	helperDomains := map[string]bool{
		"input_boolean":  true,
		"input_number":   true,
		"input_text":     true,
		"input_select":   true,
		"input_datetime": true,
		"input_button":   true,
		"counter":        true,
		"timer":          true,
		"schedule":       true,
	}

	var result []map[string]interface{}
	for _, e := range entities {
		entity, ok := e.(map[string]interface{})
		if !ok {
			continue
		}

		entityID, _ := entity["entity_id"].(string)
		parts := strings.SplitN(entityID, ".", 2)
		if len(parts) < 2 {
			continue
		}

		domain := parts[0]
		if !helperDomains[domain] {
			continue
		}

		if filterType != "" && domain != filterType && domain != "input_"+filterType {
			continue
		}

		result = append(result, map[string]interface{}{
			"entity_id": entityID,
			"name":      entity["name"],
			"type":      domain,
		})
	}

	// Handle count mode
	if helperListCount {
		client.PrintOutput(map[string]interface{}{"count": len(result)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperListLimit > 0 && len(result) > helperListLimit {
		result = result[:helperListLimit]
	}

	// Handle brief mode
	if helperListBrief {
		var brief []map[string]interface{}
		for _, item := range result {
			brief = append(brief, map[string]interface{}{
				"entity_id": item["entity_id"],
				"name":      item["name"],
			})
		}
		client.PrintOutput(brief, textMode, "")
		return nil
	}

	client.PrintOutput(result, textMode, "")
	return nil
}
