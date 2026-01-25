package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var blueprintDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a blueprint",
	Long:  `Delete a blueprint by its path. Use --domain to specify the domain (default: automation).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBlueprintDelete,
}

func init() {
	blueprintCmd.AddCommand(blueprintDeleteCmd)
	blueprintDeleteCmd.Flags().String("domain", "automation", "Domain of the blueprint (automation/script)")
	blueprintDeleteCmd.Flags().Bool("force", false, "Skip confirmation")
}

func runBlueprintDelete(cmd *cobra.Command, args []string) error {
	path := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")
	domain, _ := cmd.Flags().GetString("domain")

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

	result, err := ws.SendCommand("blueprint/delete", map[string]interface{}{
		"domain": domain,
		"path":   path,
	})
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Blueprint %s deleted successfully.", path))
	return nil
}
