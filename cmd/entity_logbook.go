package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	entityLogbookStart string
	entityLogbookEnd   string
	entityLogbookID    string
)

var entityLogbookCmd = &cobra.Command{
	Use:   "logbook [entity_id]",
	Short: "Get logbook entries",
	Long: `Get logbook entries for an entity.

The logbook shows human-readable event entries (what happened and when),
as opposed to raw state history which shows individual state values.`,
	Example: `  hab entity logbook light.kitchen
  hab entity logbook sensor.temperature -s "2025-01-01T00:00:00Z" -e "2025-01-02T00:00:00Z"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEntityLogbook,
}

func init() {
	entityCmd.AddCommand(entityLogbookCmd)
	entityLogbookCmd.Flags().StringVar(&entityLogbookID, "entity", "", "Entity ID to get logbook entries for")
	entityLogbookCmd.Flags().StringVarP(&entityLogbookStart, "start", "s", "", "Start time (ISO format, e.g. 2025-01-01T00:00:00Z)")
	entityLogbookCmd.Flags().StringVarP(&entityLogbookEnd, "end", "e", "", "End time (ISO format)")
}

func runEntityLogbook(cmd *cobra.Command, args []string) error {
	entityID, err := resolveArg(entityLogbookID, args, 0, "entity ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	entries, err := restClient.GetLogbook(entityID, entityLogbookStart, entityLogbookEnd)
	if err != nil {
		return err
	}

	output.PrintOutput(entries, textMode, "")
	return nil
}
