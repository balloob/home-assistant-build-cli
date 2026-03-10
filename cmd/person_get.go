package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var personGetID string

var personGetCmd = &cobra.Command{
	Use:     "get [person_id]",
	Short:   "Get person details",
	Long:    `Get detailed information about a person entry.`,
	Example: `  hab person get ada
  hab person get --person ada6789`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPersonGet,
}

func init() {
	personCmd.AddCommand(personGetCmd)
	personGetCmd.Flags().StringVar(&personGetID, "person", "", "Person ID to get")
}

func runPersonGet(cmd *cobra.Command, args []string) error {
	personID, err := resolveArg(personGetID, args, 0, "person ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	persons, err := ws.PersonRegistryList()
	if err != nil {
		return err
	}

	for _, p := range persons {
		m, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if m["id"] == personID {
			output.PrintOutput(m, textMode, "")
			return nil
		}
	}

	return fmt.Errorf("person '%s' not found", personID)
}
