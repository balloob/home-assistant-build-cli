package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var (
	threadAddSource string
	threadAddTLV    string
)

var threadAddCmd = &cobra.Command{
	Use:   "add [tlv]",
	Short: "Add a new Thread dataset from TLV",
	Long:  `Add a new Thread dataset from an operational dataset TLV.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runThreadAdd,
}

func init() {
	threadCmd.AddCommand(threadAddCmd)
	threadAddCmd.Flags().StringVar(&threadAddTLV, "tlv", "", "Thread operational dataset TLV")
	threadAddCmd.Flags().StringVar(&threadAddSource, "source", "CLI", "Source identifier for the dataset")
}

func runThreadAdd(cmd *cobra.Command, args []string) error {
	tlv := threadAddTLV
	if tlv == "" && len(args) > 0 {
		tlv = args[0]
	}
	if tlv == "" {
		return fmt.Errorf("TLV is required (use --tlv flag or positional argument)")
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("thread/add_dataset_tlv", map[string]interface{}{
		"source": threadAddSource,
		"tlv":    tlv,
	})
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Thread dataset added."))
	return nil
}
