package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scriptActionDeleteForce bool

var scriptActionDeleteCmd = &cobra.Command{
	Use:   "delete <script_id> <action_index>",
	Short: "Delete an action",
	Long:  `Delete an action from a script's sequence by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runScriptActionDelete,
}

func init() {
	scriptActionCmd.AddCommand(scriptActionDeleteCmd)
	scriptActionDeleteCmd.Flags().BoolVarP(&scriptActionDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runScriptActionDelete(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	scriptID = strings.TrimPrefix(scriptID, "script.")
	actionIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid action index: %s", args[1])
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	// Get current script config
	result, err := restClient.Get("config/script/config/" + scriptID)
	if err != nil {
		return err
	}

	config, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid script config")
	}

	// Get existing sequence
	sequence, ok := config["sequence"].([]interface{})
	if !ok {
		return fmt.Errorf("no sequence in script")
	}

	if actionIndex < 0 || actionIndex >= len(sequence) {
		return fmt.Errorf("action index %d out of range (0-%d)", actionIndex, len(sequence)-1)
	}

	// Confirmation prompt
	if !scriptActionDeleteForce && !textMode {
		fmt.Printf("Are you sure you want to delete action at index %d? [y/N]: ", actionIndex)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
	}

	// Remove the action
	sequence = append(sequence[:actionIndex], sequence[actionIndex+1:]...)
	config["sequence"] = sequence

	// Save the config
	_, err = restClient.Post("config/script/config/"+scriptID, config)
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Action at index %d deleted.", actionIndex))
	return nil
}
