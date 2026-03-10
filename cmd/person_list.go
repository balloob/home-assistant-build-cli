package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var personListFlags *ListFlags

var personListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all persons",
	Long:  `List all person entries in Home Assistant.`,
	RunE:  runPersonList,
}

func init() {
	personCmd.AddCommand(personListCmd)
	personListFlags = RegisterListFlags(personListCmd, "id")
}

func runPersonList(cmd *cobra.Command, args []string) error {
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

	if personListFlags.RenderCount(len(persons), textMode) {
		return nil
	}
	persons = personListFlags.ApplyLimit(persons)
	if personListFlags.RenderBrief(persons, textMode, "id", "name") {
		return nil
	}

	if textMode {
		if len(persons) == 0 {
			fmt.Println("No persons.")
			return nil
		}
		for _, p := range persons {
			if m, ok := p.(map[string]interface{}); ok {
				name, _ := m["name"].(string)
				id, _ := m["id"].(string)
				fmt.Printf("%s (%s)\n", name, id)
			}
		}
	} else {
		output.PrintOutput(persons, false, "")
	}
	return nil
}
