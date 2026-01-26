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

var (
	zoneDeleteForce bool
	zoneDeleteID    string
)

var zoneDeleteCmd = &cobra.Command{
	Use:   "delete [zone_id]",
	Short: "Delete a zone",
	Long:  `Delete a zone from Home Assistant.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runZoneDelete,
}

func init() {
	zoneCmd.AddCommand(zoneDeleteCmd)
	zoneDeleteCmd.Flags().StringVar(&zoneDeleteID, "zone", "", "Zone ID to delete")
	zoneDeleteCmd.Flags().BoolVarP(&zoneDeleteForce, "force", "f", false, "Skip confirmation")
}

func runZoneDelete(cmd *cobra.Command, args []string) error {
	zoneID := zoneDeleteID
	if zoneID == "" && len(args) > 0 {
		zoneID = args[0]
	}
	if zoneID == "" {
		return fmt.Errorf("zone ID is required (use --zone flag or positional argument)")
	}
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	if !zoneDeleteForce && !textMode {
		fmt.Printf("Delete zone %s? [y/N]: ", zoneID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

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

	if err := ws.ZoneDelete(zoneID); err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Zone '%s' deleted.", zoneID))
	return nil
}
