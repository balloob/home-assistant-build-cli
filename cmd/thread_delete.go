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
	threadDeleteForce     bool
	threadDeleteDatasetID string
)

var threadDeleteCmd = &cobra.Command{
	Use:   "delete [dataset_id]",
	Short: "Delete a Thread dataset",
	Long:  `Delete a Thread dataset.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runThreadDelete,
}

func init() {
	threadCmd.AddCommand(threadDeleteCmd)
	threadDeleteCmd.Flags().StringVar(&threadDeleteDatasetID, "dataset", "", "Thread dataset ID to delete")
	threadDeleteCmd.Flags().BoolVarP(&threadDeleteForce, "force", "f", false, "Skip confirmation")
}

func runThreadDelete(cmd *cobra.Command, args []string) error {
	datasetID := threadDeleteDatasetID
	if datasetID == "" && len(args) > 0 {
		datasetID = args[0]
	}
	if datasetID == "" {
		return fmt.Errorf("dataset ID is required (use --dataset flag or positional argument)")
	}
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	if !threadDeleteForce && !textMode {
		fmt.Printf("Delete Thread dataset %s? [y/N]: ", datasetID)
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

	_, err = ws.SendCommand("thread/delete_dataset", map[string]interface{}{
		"dataset_id": datasetID,
	})
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Thread dataset %s deleted.", datasetID))
	return nil
}
