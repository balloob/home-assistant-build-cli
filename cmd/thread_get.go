package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var threadGetDatasetID string

var threadGetCmd = &cobra.Command{
	Use:   "get [dataset_id]",
	Short: "Get dataset details including TLV",
	Long:  `Get full details of a Thread dataset.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runThreadGet,
}

func init() {
	threadCmd.AddCommand(threadGetCmd)
	threadGetCmd.Flags().StringVar(&threadGetDatasetID, "dataset", "", "Thread dataset ID to get")
}

func runThreadGet(cmd *cobra.Command, args []string) error {
	datasetID := threadGetDatasetID
	if datasetID == "" && len(args) > 0 {
		datasetID = args[0]
	}
	if datasetID == "" {
		return fmt.Errorf("dataset ID is required (use --dataset flag or positional argument)")
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

	result, err := ws.SendCommand("thread/get_dataset_tlv", map[string]interface{}{
		"dataset_id": datasetID,
	})
	if err != nil {
		return err
	}

	client.PrintOutput(result, textMode, "")
	return nil
}
