package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var deviceDeleteForce bool

var deviceDeleteCmd = &cobra.Command{
	Use:   "delete <device_id>",
	Short: "Delete a device",
	Long:  `Delete a device from Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDeviceDelete,
}

func init() {
	deviceCmd.AddCommand(deviceDeleteCmd)
	deviceDeleteCmd.Flags().BoolVarP(&deviceDeleteForce, "force", "f", false, "Skip confirmation")
}

func runDeviceDelete(cmd *cobra.Command, args []string) error {
	deviceID := args[0]
	textMode := getTextMode()

	if !deviceDeleteForce && !textMode {
		fmt.Printf("Delete device %s? This will also remove all its entities. [y/N]: ", deviceID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	_, err = ws.SendCommand("config/device_registry/remove_config_entry", map[string]interface{}{
		"device_id": deviceID,
	})
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Device '%s' deleted.", deviceID))
	return nil
}
