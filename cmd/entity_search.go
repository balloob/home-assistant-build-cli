package cmd

import (
	"sort"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var entitySearchLimit int

var entitySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Fuzzy search for entities",
	Long:  `Search for entities by entity ID or friendly name.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEntitySearch,
}

func init() {
	entityCmd.AddCommand(entitySearchCmd)
	entitySearchCmd.Flags().IntVarP(&entitySearchLimit, "limit", "n", 10, "Maximum number of results")
}

func runEntitySearch(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(args[0])
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	states, err := ws.GetStates()
	if err != nil {
		return err
	}

	type match struct {
		entityID     string
		friendlyName string
		state        string
		score        int
	}

	var matches []match
	for _, s := range states {
		state, ok := s.(map[string]interface{})
		if !ok {
			continue
		}

		entityID, _ := state["entity_id"].(string)
		attrs, _ := state["attributes"].(map[string]interface{})
		friendlyName, _ := attrs["friendly_name"].(string)
		stateVal, _ := state["state"].(string)

		score := 0
		entityIDLower := strings.ToLower(entityID)
		friendlyNameLower := strings.ToLower(friendlyName)

		if strings.Contains(entityIDLower, query) {
			score = 100 - len(entityID) // Prefer shorter matches
		} else if strings.Contains(friendlyNameLower, query) {
			score = 50 - len(friendlyName)
		}

		if score > 0 {
			matches = append(matches, match{
				entityID:     entityID,
				friendlyName: friendlyName,
				state:        stateVal,
				score:        score,
			})
		}
	}

	// Sort by score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	// Limit results
	if len(matches) > entitySearchLimit {
		matches = matches[:entitySearchLimit]
	}

	// Convert to output format
	var results []map[string]interface{}
	for _, m := range matches {
		results = append(results, map[string]interface{}{
			"entity_id": m.entityID,
			"name":      m.friendlyName,
			"state":     m.state,
		})
	}

	output.PrintOutput(results, textMode, "")
	return nil
}
