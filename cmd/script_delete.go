package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scriptDeleteForce bool

var scriptDeleteCmd = &cobra.Command{
	Use:     "delete <script_id>",
	Short:   "Delete a script",
	Long:    `Delete a script from Home Assistant.`,
	GroupID: scriptGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runScriptDelete,
}

func init() {
	scriptCmd.AddCommand(scriptDeleteCmd)
	scriptDeleteCmd.Flags().BoolVarP(&scriptDeleteForce, "force", "f", false, "Skip confirmation")
}

func runScriptDelete(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	// Strip "script." prefix if provided - API expects just the ID
	scriptID = strings.TrimPrefix(scriptID, "script.")

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	if !scriptDeleteForce && !textMode {
		fmt.Printf("Delete script %s? [y/N]: ", scriptID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	_, err = restClient.Delete("config/script/config/" + scriptID)
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Script %s deleted.", scriptID))
	return nil
}
