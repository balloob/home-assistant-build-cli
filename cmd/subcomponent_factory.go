package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/client"
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
	mergeSchemaAnnotation(parentCmd, SchemaAnnotation{
		SideEffect:   "meta",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ComponentPlural,
	})
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
	mergeSchemaAnnotation(listCmd, SchemaAnnotation{
		SideEffect:   "read",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ComponentPlural,
		InputSources: []string{"args"},
	})
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
			output.PrintOutputWithContext([]interface{}{}, textMode, "", output.EnvelopeContext{
				Operation:    "list",
				ResourceType: cfg.ComponentPlural,
			})
			return nil
		}

		items, _ := resolveItems(config, cfg)
		if items == nil {
			output.PrintOutputWithContext([]interface{}{}, textMode, "", output.EnvelopeContext{
				Operation:    "list",
				ResourceType: cfg.ComponentPlural,
			})
			return nil
		}

		itemList := make([]map[string]interface{}, len(items))
		for i, item := range items {
			data := copyItemMap(item)
			data["index"] = i
			itemList[i] = data
		}

		output.PrintOutputWithContext(itemList, textMode, "", output.EnvelopeContext{
			Operation:    "list",
			ResourceType: cfg.ComponentPlural,
		})
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
	mergeSchemaAnnotation(getCmd, SchemaAnnotation{
		SideEffect:   "read",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ComponentName,
		InputSources: []string{"args", "flags"},
	})
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

		output.PrintOutputWithContext(data, textMode, "", output.EnvelopeContext{
			Operation:    "get",
			ResourceType: cfg.ComponentName,
		})
		return nil
	}
}

// --- create ---

func registerSubComponentCreate(parentCmd *cobra.Command, cfg SubComponentConfig) {
	var inputFlags InputFlags
	var plan bool
	var dryRun bool

	createCmd := &cobra.Command{
		Use:   fmt.Sprintf("create <%s_id>", cfg.ParentName),
		Short: fmt.Sprintf("Create a new %s", cfg.ComponentName),
		Long:  fmt.Sprintf("Create a new %s in %s.", cfg.ComponentName, addArticle(cfg.ParentName)),
		Args:  cobra.ExactArgs(1),
		RunE:  makeSubComponentCreate(cfg, &inputFlags),
	}
	mergeSchemaAnnotation(createCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ComponentName,
		InputSources: []string{"args", "flags", "data", "file"},
	})
	inputFlags.Register(createCmd)
	registerMutationPlanFlags(createCmd, &plan, &dryRun)
	parentCmd.AddCommand(createCmd)
}

func makeSubComponentCreate(cfg SubComponentConfig, flags *InputFlags) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()
		plan, _ := cmd.Flags().GetBool("plan")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		itemConfig, err := flags.Parse()
		if err != nil {
			return err
		}

		if planRequested(plan, dryRun) {
			resolvedID := args[0]
			if restClient, getErr := getRESTClient(); getErr == nil {
				if configID, resolveErr := cfg.ResolveID(restClient, args[0]); resolveErr == nil {
					resolvedID = configID
				}
			}

			printMutationPlan(MutationPlan{
				WouldChange: true,
				Target: map[string]any{
					"resource":  cfg.ComponentName,
					"parent":    cfg.ParentName,
					"parent_id": args[0],
					"endpoint":  cfg.APIBasePath + resolvedID,
				},
				Inputs: map[string]any{"config": itemConfig},
				DerivedIDs: map[string]any{
					"resolved_parent_config_id": resolvedID,
				},
				Steps: []string{
					"Resolve parent ID to parent config ID.",
					"Fetch parent config and append new subcomponent entry.",
					fmt.Sprintf("POST %s with updated config payload.", cfg.APIBasePath+resolvedID),
				},
				VerificationCommands: []string{
					fmt.Sprintf("hab %s %s list %s --json", cfg.ParentName, cfg.ComponentName, args[0]),
				},
				RequiresConfirmation: false,
			}, textMode, cfg.ComponentName)
			return nil
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
		output.PrintSuccessWithContext(resultData, textMode, fmt.Sprintf("%s created at index %d.", capitalize(cfg.ComponentName), len(items)-1), output.EnvelopeContext{
			Operation:    "create",
			ResourceType: cfg.ComponentName,
		})
		return nil
	}
}

// --- update ---

func registerSubComponentUpdate(parentCmd *cobra.Command, cfg SubComponentConfig) {
	var inputFlags InputFlags
	var plan bool
	var dryRun bool

	updateCmd := &cobra.Command{
		Use:   fmt.Sprintf("update <%s_id> <%s_index>", cfg.ParentName, cfg.ComponentName),
		Short: fmt.Sprintf("Update %s", addArticle(cfg.ComponentName)),
		Long:  fmt.Sprintf("Update %s in %s by index.", addArticle(cfg.ComponentName), addArticle(cfg.ParentName)),
		Args:  cobra.ExactArgs(2),
		RunE:  makeSubComponentUpdate(cfg, &inputFlags),
	}
	mergeSchemaAnnotation(updateCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ComponentName,
		InputSources: []string{"args", "flags", "data", "file"},
	})
	inputFlags.Register(updateCmd)
	registerMutationPlanFlags(updateCmd, &plan, &dryRun)
	parentCmd.AddCommand(updateCmd)
}

func makeSubComponentUpdate(cfg SubComponentConfig, flags *InputFlags) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid %s index: %s", cfg.ComponentName, args[1])
		}

		textMode := getTextMode()
		plan, _ := cmd.Flags().GetBool("plan")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		newItem, err := flags.Parse()
		if err != nil {
			return err
		}

		if planRequested(plan, dryRun) {
			resolvedID := args[0]
			if restClient, getErr := getRESTClient(); getErr == nil {
				if configID, resolveErr := cfg.ResolveID(restClient, args[0]); resolveErr == nil {
					resolvedID = configID
				}
			}

			printMutationPlan(MutationPlan{
				WouldChange: true,
				Target: map[string]any{
					"resource":  cfg.ComponentName,
					"parent":    cfg.ParentName,
					"parent_id": args[0],
					"index":     idx,
					"endpoint":  cfg.APIBasePath + resolvedID,
				},
				Inputs: map[string]any{"config": newItem},
				DerivedIDs: map[string]any{
					"resolved_parent_config_id": resolvedID,
				},
				Steps: []string{
					"Resolve parent ID to parent config ID.",
					"Fetch parent config and replace subcomponent at target index.",
					fmt.Sprintf("POST %s with updated config payload.", cfg.APIBasePath+resolvedID),
				},
				VerificationCommands: []string{
					fmt.Sprintf("hab %s %s get %s %d --json", cfg.ParentName, cfg.ComponentName, args[0], idx),
				},
				RequiresConfirmation: false,
			}, textMode, cfg.ComponentName)
			return nil
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
		output.PrintSuccessWithContext(resultData, textMode, fmt.Sprintf("%s at index %d updated.", capitalize(cfg.ComponentName), idx), output.EnvelopeContext{
			Operation:    "update",
			ResourceType: cfg.ComponentName,
		})
		return nil
	}
}

// --- delete ---

func registerSubComponentDelete(parentCmd *cobra.Command, cfg SubComponentConfig) {
	var force bool
	var plan bool
	var dryRun bool

	deleteCmd := &cobra.Command{
		Use:   fmt.Sprintf("delete <%s_id> <%s_index>", cfg.ParentName, cfg.ComponentName),
		Short: fmt.Sprintf("Delete %s", addArticle(cfg.ComponentName)),
		Long:  fmt.Sprintf("Delete %s from %s by index.", addArticle(cfg.ComponentName), addArticle(cfg.ParentName)),
		Args:  cobra.ExactArgs(2),
		RunE:  makeSubComponentDelete(cfg, &force),
	}
	mergeSchemaAnnotation(deleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ComponentName,
		InputSources: []string{"args", "flags"},
	})
	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	registerMutationPlanFlags(deleteCmd, &plan, &dryRun)
	parentCmd.AddCommand(deleteCmd)
}

func makeSubComponentDelete(cfg SubComponentConfig, force *bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid %s index: %s", cfg.ComponentName, args[1])
		}

		textMode := getTextMode()
		plan, _ := cmd.Flags().GetBool("plan")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if planRequested(plan, dryRun) {
			resolvedID := args[0]
			if restClient, getErr := getRESTClient(); getErr == nil {
				if configID, resolveErr := cfg.ResolveID(restClient, args[0]); resolveErr == nil {
					resolvedID = configID
				}
			}

			printMutationPlan(MutationPlan{
				WouldChange: true,
				Target: map[string]any{
					"resource":  cfg.ComponentName,
					"parent":    cfg.ParentName,
					"parent_id": args[0],
					"index":     idx,
					"endpoint":  cfg.APIBasePath + resolvedID,
				},
				DerivedIDs: map[string]any{
					"resolved_parent_config_id": resolvedID,
				},
				Steps: []string{
					"Resolve parent ID to parent config ID.",
					"Fetch parent config and remove subcomponent at target index.",
					fmt.Sprintf("POST %s with updated config payload.", cfg.APIBasePath+resolvedID),
				},
				Risks: []string{
					"This operation permanently removes the selected subcomponent entry.",
				},
				RequiresConfirmation: !*force,
				VerificationCommands: []string{
					fmt.Sprintf("hab %s %s list %s --json", cfg.ParentName, cfg.ComponentName, args[0]),
				},
			}, textMode, cfg.ComponentName)
			return nil
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

		if err := confirmAction(*force, fmt.Sprintf("Are you sure you want to delete %s at index %d?", cfg.ComponentName, idx), fmt.Sprintf("delete %s", cfg.ComponentName)); err != nil {
			return err
		}

		items = append(items[:idx], items[idx+1:]...)
		config[key] = items

		_, err = restClient.Post(cfg.APIBasePath+configID, config)
		if err != nil {
			return err
		}

		output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("%s at index %d deleted.", capitalize(cfg.ComponentName), idx), output.EnvelopeContext{
			Operation:    "delete",
			ResourceType: cfg.ComponentName,
		})
		return nil
	}
}
