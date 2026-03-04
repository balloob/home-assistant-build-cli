package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

// SubComponentConfig defines the parameters for generating CRUD commands
// for a sub-component of a parent entity (e.g. automation triggers,
// script actions). All five CRUD operations (list, get, create, update,
// delete) plus a parent group command are generated from this config.
type SubComponentConfig struct {
	// ParentCmd is the cobra parent under which the sub-component group
	// command is registered (e.g. automationCmd, scriptCmd).
	ParentCmd *cobra.Command

	// ParentName is the human-readable parent type used in help text and
	// error messages (e.g. "automation", "script").
	ParentName string

	// ComponentName is the singular sub-component name used in command
	// names and messages (e.g. "trigger", "condition", "action").
	ComponentName string

	// ComponentPlural is the plural form used in descriptions
	// (e.g. "triggers", "conditions", "actions").
	ComponentPlural string

	// ConfigKeys lists the JSON keys to try when extracting the
	// sub-component array from the parent config, in priority order.
	// For automations: []string{"triggers", "trigger"}.
	// For scripts: []string{"sequence"}.
	ConfigKeys []string

	// DefaultKey is the key used when creating a new array in the config
	// (e.g. "triggers", "sequence").
	DefaultKey string

	// APIBasePath is the REST API path prefix for the parent config
	// (e.g. "config/automation/config/", "config/script/config/").
	APIBasePath string

	// ResolveID converts a user-provided identifier to the internal
	// config ID. For automations this resolves entity_id -> config ID;
	// for scripts this strips the "script." prefix.
	ResolveID func(restClient client.RestAPI, rawID string) (string, error)

	// ParentFlagName is the flag name for the get command's parent ID
	// (e.g. "automation", "script").
	ParentFlagName string

	// GroupID is the command group ID under which the parent command
	// is registered (e.g. automationGroupSubcommands).
	GroupID string
}

// RegisterSubComponentCRUD creates and registers a parent group command
// plus list, get, create, update, and delete subcommands for the given
// sub-component configuration.
func RegisterSubComponentCRUD(cfg SubComponentConfig) {
	// Parent group command (e.g. "automation trigger", "script action")
	parentCmd := &cobra.Command{
		Use:     cfg.ComponentName,
		Short:   fmt.Sprintf("Manage %s %s", cfg.ParentName, cfg.ComponentPlural),
		Long:    fmt.Sprintf("Create, update, list, and delete %s in %s.", cfg.ComponentPlural, addArticle(cfg.ParentName)),
		GroupID: cfg.GroupID,
	}
	cfg.ParentCmd.AddCommand(parentCmd)

	registerSubComponentList(parentCmd, cfg)
	registerSubComponentGet(parentCmd, cfg)
	registerSubComponentCreate(parentCmd, cfg)
	registerSubComponentUpdate(parentCmd, cfg)
	registerSubComponentDelete(parentCmd, cfg)
}

// addArticle returns "a <noun>" or "an <noun>" depending on the first letter.
func addArticle(noun string) string {
	if len(noun) == 0 {
		return noun
	}
	switch noun[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an " + noun
	default:
		return "a " + noun
	}
}

// resolveItems extracts the sub-component array from a parent config map.
// It tries each key in cfg.ConfigKeys in order, returning the array and
// the key that matched. If no key matches, items is nil and key is empty.
func resolveItems(config map[string]interface{}, cfg SubComponentConfig) (items []interface{}, key string) {
	for _, k := range cfg.ConfigKeys {
		if arr, ok := config[k].([]interface{}); ok {
			return arr, k
		}
	}
	return nil, ""
}

// fetchParentConfig gets the REST client, resolves the parent ID, and
// fetches the parent config map. Returns the client, resolved config ID,
// and config map.
func fetchParentConfig(cfg SubComponentConfig, rawID string) (client.RestAPI, string, map[string]interface{}, error) {
	restClient, err := getRESTClient()
	if err != nil {
		return nil, "", nil, err
	}

	configID, err := cfg.ResolveID(restClient, rawID)
	if err != nil {
		return nil, "", nil, err
	}

	result, err := restClient.Get(cfg.APIBasePath + configID)
	if err != nil {
		return nil, "", nil, err
	}

	config, ok := result.(map[string]interface{})
	if !ok {
		return restClient, configID, nil, nil
	}
	return restClient, configID, config, nil
}

// copyItemMap shallow-copies an interface{} that is expected to be a map
// into a new map and returns it. Non-map values return an empty map.
func copyItemMap(item interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	if m, ok := item.(map[string]interface{}); ok {
		for k, v := range m {
			data[k] = v
		}
	}
	return data
}

// --- list ---

func registerSubComponentList(parentCmd *cobra.Command, cfg SubComponentConfig) {
	listCmd := &cobra.Command{
		Use:   fmt.Sprintf("list <%s_id>", cfg.ParentName),
		Short: fmt.Sprintf("List %s in %s", cfg.ComponentPlural, addArticle(cfg.ParentName)),
		Long:  fmt.Sprintf("List all %s in %s.", cfg.ComponentPlural, addArticle(cfg.ParentName)),
		Args:  cobra.ExactArgs(1),
		RunE:  makeSubComponentList(cfg),
	}
	parentCmd.AddCommand(listCmd)
}

func makeSubComponentList(cfg SubComponentConfig) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()

		_, _, config, err := fetchParentConfig(cfg, args[0])
		if err != nil {
			return err
		}
		if config == nil {
			output.PrintOutput([]interface{}{}, textMode, "")
			return nil
		}

		items, _ := resolveItems(config, cfg)
		if items == nil {
			output.PrintOutput([]interface{}{}, textMode, "")
			return nil
		}

		itemList := make([]map[string]interface{}, len(items))
		for i, item := range items {
			data := copyItemMap(item)
			data["index"] = i
			itemList[i] = data
		}

		output.PrintOutput(itemList, textMode, "")
		return nil
	}
}

// --- get ---

func registerSubComponentGet(parentCmd *cobra.Command, cfg SubComponentConfig) {
	var parentID string
	var itemIndex int

	getCmd := &cobra.Command{
		Use:   fmt.Sprintf("get [%s_id] [%s_index]", cfg.ParentName, cfg.ComponentName),
		Short: fmt.Sprintf("Get a specific %s", cfg.ComponentName),
		Long:  fmt.Sprintf("Get a specific %s from %s by index.", cfg.ComponentName, addArticle(cfg.ParentName)),
		Args:  cobra.MaximumNArgs(2),
		RunE:  makeSubComponentGet(cfg, &parentID, &itemIndex),
	}
	getCmd.Flags().StringVar(&parentID, cfg.ParentFlagName, "", fmt.Sprintf("%s ID", capitalize(cfg.ParentName)))
	getCmd.Flags().IntVar(&itemIndex, "index", -1, fmt.Sprintf("%s index", capitalize(cfg.ComponentName)))
	parentCmd.AddCommand(getCmd)
}

func makeSubComponentGet(cfg SubComponentConfig, parentID *string, itemIndex *int) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		id := *parentID
		if id == "" && len(args) > 0 {
			id = args[0]
		}
		if id == "" {
			return fmt.Errorf("%s ID is required (use --%s flag or first positional argument)", cfg.ParentName, cfg.ParentFlagName)
		}

		idx := *itemIndex
		if idx < 0 && len(args) > 1 {
			var err error
			idx, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid %s index: %s", cfg.ComponentName, args[1])
			}
		}
		if idx < 0 {
			return fmt.Errorf("%s index is required (use --index flag or second positional argument)", cfg.ComponentName)
		}

		textMode := getTextMode()

		_, _, config, err := fetchParentConfig(cfg, id)
		if err != nil {
			return err
		}
		if config == nil {
			return fmt.Errorf("invalid %s config", cfg.ParentName)
		}

		items, _ := resolveItems(config, cfg)
		if items == nil {
			return fmt.Errorf("no %s in %s", cfg.ComponentPlural, cfg.ParentName)
		}

		if idx < 0 || idx >= len(items) {
			return fmt.Errorf("%s index %d out of range (0-%d)", cfg.ComponentName, idx, len(items)-1)
		}

		data := copyItemMap(items[idx])
		data["index"] = idx

		output.PrintOutput(data, textMode, "")
		return nil
	}
}

// --- create ---

func registerSubComponentCreate(parentCmd *cobra.Command, cfg SubComponentConfig) {
	var data, file, format string

	createCmd := &cobra.Command{
		Use:   fmt.Sprintf("create <%s_id>", cfg.ParentName),
		Short: fmt.Sprintf("Create a new %s", cfg.ComponentName),
		Long:  fmt.Sprintf("Create a new %s in %s.", cfg.ComponentName, addArticle(cfg.ParentName)),
		Args:  cobra.ExactArgs(1),
		RunE:  makeSubComponentCreate(cfg, &data, &file, &format),
	}
	createCmd.Flags().StringVarP(&data, "data", "d", "", fmt.Sprintf("%s configuration as JSON", capitalize(cfg.ComponentName)))
	createCmd.Flags().StringVarP(&file, "file", "f", "", "Path to config file")
	createCmd.Flags().StringVar(&format, "format", "", "Input format (json, yaml)")
	parentCmd.AddCommand(createCmd)
}

func makeSubComponentCreate(cfg SubComponentConfig, dataFlag, fileFlag, formatFlag *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()

		itemConfig, err := input.ParseInput(*dataFlag, *fileFlag, *formatFlag)
		if err != nil {
			return err
		}

		restClient, configID, config, err := fetchParentConfig(cfg, args[0])
		if err != nil {
			return err
		}
		if config == nil {
			return fmt.Errorf("invalid %s config", cfg.ParentName)
		}

		// Get existing items or initialize
		items, key := resolveItems(config, cfg)
		if items == nil {
			items = []interface{}{}
			key = cfg.DefaultKey
		}

		items = append(items, itemConfig)
		config[key] = items

		_, err = restClient.Post(cfg.APIBasePath+configID, config)
		if err != nil {
			return err
		}

		resultData := map[string]interface{}{
			"index":  len(items) - 1,
			"config": itemConfig,
		}
		output.PrintSuccess(resultData, textMode, fmt.Sprintf("%s created at index %d.", capitalize(cfg.ComponentName), len(items)-1))
		return nil
	}
}

// --- update ---

func registerSubComponentUpdate(parentCmd *cobra.Command, cfg SubComponentConfig) {
	var data, file, format string

	updateCmd := &cobra.Command{
		Use:   fmt.Sprintf("update <%s_id> <%s_index>", cfg.ParentName, cfg.ComponentName),
		Short: fmt.Sprintf("Update %s", addArticle(cfg.ComponentName)),
		Long:  fmt.Sprintf("Update %s in %s by index.", addArticle(cfg.ComponentName), addArticle(cfg.ParentName)),
		Args:  cobra.ExactArgs(2),
		RunE:  makeSubComponentUpdate(cfg, &data, &file, &format),
	}
	updateCmd.Flags().StringVarP(&data, "data", "d", "", fmt.Sprintf("%s configuration as JSON (replaces entire %s)", capitalize(cfg.ComponentName), cfg.ComponentName))
	updateCmd.Flags().StringVarP(&file, "file", "f", "", "Path to config file")
	updateCmd.Flags().StringVar(&format, "format", "", "Input format (json, yaml)")
	parentCmd.AddCommand(updateCmd)
}

func makeSubComponentUpdate(cfg SubComponentConfig, dataFlag, fileFlag, formatFlag *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid %s index: %s", cfg.ComponentName, args[1])
		}

		textMode := getTextMode()

		newItem, err := input.ParseInput(*dataFlag, *fileFlag, *formatFlag)
		if err != nil {
			return err
		}

		restClient, configID, config, err := fetchParentConfig(cfg, args[0])
		if err != nil {
			return err
		}
		if config == nil {
			return fmt.Errorf("invalid %s config", cfg.ParentName)
		}

		items, key := resolveItems(config, cfg)
		if items == nil {
			return fmt.Errorf("no %s in %s", cfg.ComponentPlural, cfg.ParentName)
		}

		if idx < 0 || idx >= len(items) {
			return fmt.Errorf("%s index %d out of range (0-%d)", cfg.ComponentName, idx, len(items)-1)
		}

		items[idx] = newItem
		config[key] = items

		_, err = restClient.Post(cfg.APIBasePath+configID, config)
		if err != nil {
			return err
		}

		resultData := map[string]interface{}{
			"index":  idx,
			"config": newItem,
		}
		output.PrintSuccess(resultData, textMode, fmt.Sprintf("%s at index %d updated.", capitalize(cfg.ComponentName), idx))
		return nil
	}
}

// --- delete ---

func registerSubComponentDelete(parentCmd *cobra.Command, cfg SubComponentConfig) {
	var force bool

	deleteCmd := &cobra.Command{
		Use:   fmt.Sprintf("delete <%s_id> <%s_index>", cfg.ParentName, cfg.ComponentName),
		Short: fmt.Sprintf("Delete %s", addArticle(cfg.ComponentName)),
		Long:  fmt.Sprintf("Delete %s from %s by index.", addArticle(cfg.ComponentName), addArticle(cfg.ParentName)),
		Args:  cobra.ExactArgs(2),
		RunE:  makeSubComponentDelete(cfg, &force),
	}
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	parentCmd.AddCommand(deleteCmd)
}

func makeSubComponentDelete(cfg SubComponentConfig, force *bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid %s index: %s", cfg.ComponentName, args[1])
		}

		textMode := getTextMode()

		restClient, configID, config, err := fetchParentConfig(cfg, args[0])
		if err != nil {
			return err
		}
		if config == nil {
			return fmt.Errorf("invalid %s config", cfg.ParentName)
		}

		items, key := resolveItems(config, cfg)
		if items == nil {
			return fmt.Errorf("no %s in %s", cfg.ComponentPlural, cfg.ParentName)
		}

		if idx < 0 || idx >= len(items) {
			return fmt.Errorf("%s index %d out of range (0-%d)", cfg.ComponentName, idx, len(items)-1)
		}

		if !*force && !textMode {
			fmt.Printf("Are you sure you want to delete %s at index %d? [y/N]: ", cfg.ComponentName, idx)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				return fmt.Errorf("deletion cancelled")
			}
		}

		items = append(items[:idx], items[idx+1:]...)
		config[key] = items

		_, err = restClient.Post(cfg.APIBasePath+configID, config)
		if err != nil {
			return err
		}

		output.PrintSuccess(nil, textMode, fmt.Sprintf("%s at index %d deleted.", capitalize(cfg.ComponentName), idx))
		return nil
	}
}
