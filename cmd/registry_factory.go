package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// Registry CRUD factory
//
// Generates list, get, create, update, and delete subcommands for registry
// resources (areas, floors, labels) from a declarative configuration.
// ---------------------------------------------------------------------------

// RegistryFlagType distinguishes string vs. integer flags.
type RegistryFlagType int

const (
	// FlagString is a string flag (default "").
	FlagString RegistryFlagType = iota
	// FlagInt is an integer flag (uses cmd.Flags().Changed to detect zero values).
	FlagInt
)

// RegistryFlagDef defines a flag for create or update subcommands.
type RegistryFlagDef struct {
	Name     string           // flag name, e.g. "floor", "icon", "level"
	ParamKey string           // API param key, e.g. "floor_id", "icon", "level"
	Usage    string           // flag help text
	Type     RegistryFlagType // string or int
}

// RegistryFilterDef defines a list filter flag.
type RegistryFilterDef struct {
	FlagName  string // flag name, e.g. "area-id", "floor"
	Shorthand string // optional one-letter shorthand, e.g. "f"
	FieldKey  string // field to match in the result, e.g. "area_id", "floor_id"
	Usage     string // flag help text
}

// RegistryCRUDConfig holds all configuration needed to generate CRUD
// subcommands for a registry resource.
type RegistryCRUDConfig struct {
	ParentCmd    *cobra.Command
	ResourceName string // "area", "floor", "label"
	IDField      string // "area_id", "floor_id", "label_id"
	SearchType   string // for SearchRelated: "area", "floor", "label" (empty = no --related)
	IDFlagName   string // flag name for get command: "area", "floor", "label"

	// List configuration
	ListFilters []RegistryFilterDef
	BriefFields []string // fields for brief mode, e.g. ["area_id", "name"]

	// Create/Update flags
	CreateFlags []RegistryFlagDef
	UpdateFlags []RegistryFlagDef // --name is always added automatically

	// WS API functions
	ListFunc   func(ws client.WebSocketAPI) ([]interface{}, error)
	CreateFunc func(ws client.WebSocketAPI, name string, params map[string]interface{}) (map[string]interface{}, error)
	UpdateFunc func(ws client.WebSocketAPI, id string, params map[string]interface{}) (map[string]interface{}, error)
	DeleteFunc func(ws client.WebSocketAPI, id string) error
}

// RegisterRegistryCRUD generates and registers list, get, create, update,
// and delete subcommands on the parent command.
func RegisterRegistryCRUD(cfg RegistryCRUDConfig) {
	registerRegistryList(cfg)
	registerRegistryGet(cfg)
	registerRegistryCreate(cfg)
	registerRegistryUpdate(cfg)
	registerRegistryDelete(cfg)
}

// ---------------------------------------------------------------------------
// list
// ---------------------------------------------------------------------------

func registerRegistryList(cfg RegistryCRUDConfig) {
	// Filter flag values: map flag-name -> *string
	filterValues := make(map[string]*string)
	var lf *ListFlags

	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List all %ss", cfg.ResourceName),
		Long:  fmt.Sprintf("List all %ss in Home Assistant.", cfg.ResourceName),
		RunE: func(cmd *cobra.Command, args []string) error {
			textMode := getTextMode()

			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			items, err := cfg.ListFunc(ws)
			if err != nil {
				return err
			}

			// Apply filters
			for _, f := range cfg.ListFilters {
				val := *filterValues[f.FlagName]
				if val != "" {
					var filtered []interface{}
					for _, item := range items {
						if m, ok := item.(map[string]interface{}); ok {
							fieldVal, _ := m[f.FieldKey].(string)
							if fieldVal == val {
								filtered = append(filtered, item)
							}
						}
					}
					items = filtered
				}
			}

			// Count mode
			if lf.RenderCount(len(items), textMode) {
				return nil
			}

			// Apply limit
			items = lf.ApplyLimit(items)

			// Brief mode
			if lf.RenderBriefFields(items, textMode, cfg.IDField, "name", cfg.BriefFields) {
				return nil
			}

			// Full output
			if textMode {
				if len(items) == 0 {
					fmt.Printf("No %ss.\n", cfg.ResourceName)
					return nil
				}
				for _, item := range items {
					if m, ok := item.(map[string]interface{}); ok {
						name, _ := m["name"].(string)
						id, _ := m[cfg.IDField].(string)
						fmt.Printf("%s (%s)\n", name, id)
					}
				}
			} else {
				output.PrintOutputWithContext(items, false, "", output.EnvelopeContext{
					Operation:    "list",
					ResourceType: cfg.ResourceName,
				})
			}
			return nil
		},
	}
	mergeSchemaAnnotation(listCmd, SchemaAnnotation{
		SideEffect:   "read",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: cfg.ResourceName,
		GuideTopic:   "registry",
		OutputContract: func() *SchemaOutputContract {
			contract := registryListOutputContract(cfg.ResourceName, cfg.IDField)
			return &contract
		}(),
	})

	// Register standard list flags
	lf = RegisterListFlags(listCmd, cfg.IDField)

	// Register filter flags
	for i := range cfg.ListFilters {
		f := &cfg.ListFilters[i]
		val := new(string)
		filterValues[f.FlagName] = val
		listCmd.Flags().StringVarP(val, f.FlagName, f.Shorthand, "", f.Usage)
	}

	cfg.ParentCmd.AddCommand(listCmd)
}

// ---------------------------------------------------------------------------
// get
// ---------------------------------------------------------------------------

func registerRegistryGet(cfg RegistryCRUDConfig) {
	var idFlag string
	var related bool

	getCmd := &cobra.Command{
		Use:   fmt.Sprintf("get [%s]", cfg.IDField),
		Short: fmt.Sprintf("Get %s details", cfg.ResourceName),
		Long:  fmt.Sprintf("Get detailed information about a %s.", cfg.ResourceName),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveArg(idFlag, args, 0, cfg.ResourceName+" ID")
			if err != nil {
				return err
			}
			textMode := getTextMode()

			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			items, err := cfg.ListFunc(ws)
			if err != nil {
				return err
			}

			for _, item := range items {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				if m[cfg.IDField] == id {
					result := m

					if related && cfg.SearchType != "" {
						rel, err := ws.SearchRelated(cfg.SearchType, id)
						if err == nil && len(rel) > 0 {
							resultMap := make(map[string]interface{})
							for k, v := range m {
								resultMap[k] = v
							}
							resultMap["related"] = rel
							result = resultMap
						} else if err != nil {
							noteFallback(fmt.Sprintf("%s related lookup failed; returned registry data without related items", cfg.ResourceName))
							noteMissingSection("related")
						}
					}

					output.PrintOutputWithContext(result, textMode, "", output.EnvelopeContext{
						Operation:    "get",
						ResourceType: cfg.ResourceName,
					})
					return nil
				}
			}

			return fmt.Errorf("%s '%s' not found", cfg.ResourceName, id)
		},
	}
	mergeSchemaAnnotation(getCmd, SchemaAnnotation{
		SideEffect:   "read",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: cfg.ResourceName,
		GuideTopic:   "registry",
		OutputContract: func() *SchemaOutputContract {
			contract := registryGetOutputContract(cfg.ResourceName, cfg.IDField)
			return &contract
		}(),
	})

	getCmd.Flags().StringVar(&idFlag, cfg.IDFlagName, "", fmt.Sprintf("%s ID to get", capitalize(cfg.ResourceName)))
	if cfg.SearchType != "" {
		getCmd.Flags().BoolVarP(&related, "related", "r", false, "Include related items")
	}

	cfg.ParentCmd.AddCommand(getCmd)
}

// ---------------------------------------------------------------------------
// create
// ---------------------------------------------------------------------------

func registerRegistryCreate(cfg RegistryCRUDConfig) {
	// Allocate flag storage
	stringFlags := make(map[string]*string)
	intFlags := make(map[string]*int)
	var plan bool
	var dryRun bool

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: fmt.Sprintf("Create a new %s", cfg.ResourceName),
		Long:  fmt.Sprintf("Create a new %s in Home Assistant.", cfg.ResourceName),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			textMode := getTextMode()
			params := buildRegistryParams(cmd, cfg.CreateFlags, stringFlags, intFlags)

			if planRequested(plan, dryRun) {
				printMutationPlan(MutationPlan{
					WouldChange: true,
					Target: map[string]any{
						"resource": cfg.ResourceName,
						"name":     name,
					},
					Inputs: map[string]any{"params": params},
					Steps: []string{
						"Connect to Home Assistant WebSocket API.",
						fmt.Sprintf("Send %s command with merged name/flag payload.", registryCommandName(cfg.ResourceName, "create")),
						"Return created resource metadata.",
					},
					VerificationCommands: []string{
						fmt.Sprintf("hab %s list --json", cfg.ResourceName),
					},
					RequiresConfirmation: false,
				}, textMode, cfg.ResourceName)
				return nil
			}

			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			result, err := cfg.CreateFunc(ws, name, params)
			if err != nil {
				return err
			}

			output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("%s '%s' created.", capitalize(cfg.ResourceName), name), output.EnvelopeContext{
				Operation:    "create",
				ResourceType: cfg.ResourceName,
			})
			return nil
		},
	}
	mergeSchemaAnnotation(createCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: cfg.ResourceName,
		InputSources: []string{"args", "flags"},
		GuideTopic:   "registry",
	})

	registerRegistryFlags(createCmd, cfg.CreateFlags, stringFlags, intFlags)
	registerMutationPlanFlags(createCmd, &plan, &dryRun)
	cfg.ParentCmd.AddCommand(createCmd)
}

// ---------------------------------------------------------------------------
// update
// ---------------------------------------------------------------------------

func registerRegistryUpdate(cfg RegistryCRUDConfig) {
	// Allocate flag storage — includes --name plus custom flags
	stringFlags := make(map[string]*string)
	intFlags := make(map[string]*int)
	var plan bool
	var dryRun bool

	updateCmd := &cobra.Command{
		Use:   fmt.Sprintf("update <%s>", cfg.IDField),
		Short: fmt.Sprintf("Update a %s", cfg.ResourceName),
		Long:  fmt.Sprintf("Update an existing %s.", cfg.ResourceName),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			textMode := getTextMode()

			params := buildRegistryParams(cmd, cfg.UpdateFlags, stringFlags, intFlags)

			// Check --name flag
			if nameVal := *stringFlags["name"]; nameVal != "" {
				params["name"] = nameVal
			}

			if len(params) == 0 {
				return fmt.Errorf("no update parameters provided")
			}

			if planRequested(plan, dryRun) {
				printMutationPlan(MutationPlan{
					WouldChange: true,
					Target: map[string]any{
						"resource": cfg.ResourceName,
						"id":       id,
					},
					Inputs: map[string]any{"params": params},
					Steps: []string{
						"Connect to Home Assistant WebSocket API.",
						fmt.Sprintf("Send %s command with provided update payload.", registryCommandName(cfg.ResourceName, "update")),
						"Return updated resource metadata.",
					},
					VerificationCommands: []string{
						fmt.Sprintf("hab %s get %s --json", cfg.ResourceName, id),
					},
					RequiresConfirmation: false,
				}, textMode, cfg.ResourceName)
				return nil
			}

			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			result, err := cfg.UpdateFunc(ws, id, params)
			if err != nil {
				return err
			}

			output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("%s '%s' updated.", capitalize(cfg.ResourceName), id), output.EnvelopeContext{
				Operation:    "update",
				ResourceType: cfg.ResourceName,
			})
			return nil
		},
	}
	mergeSchemaAnnotation(updateCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: cfg.ResourceName,
		InputSources: []string{"args", "flags"},
		GuideTopic:   "registry",
	})

	// Always add --name for update
	nameVal := new(string)
	stringFlags["name"] = nameVal
	updateCmd.Flags().StringVar(nameVal, "name", "", fmt.Sprintf("New name for the %s", cfg.ResourceName))

	registerRegistryFlags(updateCmd, cfg.UpdateFlags, stringFlags, intFlags)
	registerMutationPlanFlags(updateCmd, &plan, &dryRun)
	cfg.ParentCmd.AddCommand(updateCmd)
}

// ---------------------------------------------------------------------------
// delete
// ---------------------------------------------------------------------------

func registerRegistryDelete(cfg RegistryCRUDConfig) {
	var force bool
	var plan bool
	var dryRun bool

	deleteCmd := &cobra.Command{
		Use:   fmt.Sprintf("delete <%s>", cfg.IDField),
		Short: fmt.Sprintf("Delete a %s", cfg.ResourceName),
		Long:  fmt.Sprintf("Delete a %s from Home Assistant.", cfg.ResourceName),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			textMode := getTextMode()

			if planRequested(plan, dryRun) {
				printMutationPlan(MutationPlan{
					WouldChange: true,
					Target: map[string]any{
						"resource": cfg.ResourceName,
						"id":       id,
					},
					Steps: []string{
						"Connect to Home Assistant WebSocket API.",
						fmt.Sprintf("Send %s delete command for the target ID.", registryCommandName(cfg.ResourceName, "delete")),
						"Return deletion success message.",
					},
					Risks: []string{
						"This operation permanently removes the registry resource.",
					},
					RequiresConfirmation: !force,
					VerificationCommands: []string{
						fmt.Sprintf("hab %s list --json", cfg.ResourceName),
					},
				}, textMode, cfg.ResourceName)
				return nil
			}

			if err := confirmAction(force, fmt.Sprintf("Delete %s %s?", cfg.ResourceName, id), fmt.Sprintf("delete %s", cfg.ResourceName)); err != nil {
				return err
			}

			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			if err := cfg.DeleteFunc(ws, id); err != nil {
				return err
			}

			output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("%s '%s' deleted.", capitalize(cfg.ResourceName), id), output.EnvelopeContext{
				Operation:    "delete",
				ResourceType: cfg.ResourceName,
			})
			return nil
		},
	}
	mergeSchemaAnnotation(deleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: cfg.ResourceName,
		InputSources: []string{"args", "flags"},
		GuideTopic:   "registry",
	})

	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(deleteCmd, &plan, &dryRun)
	cfg.ParentCmd.AddCommand(deleteCmd)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// registerRegistryFlags registers string and int flags on a command from flag definitions.
func registerRegistryFlags(cmd *cobra.Command, defs []RegistryFlagDef, stringFlags map[string]*string, intFlags map[string]*int) {
	for _, def := range defs {
		switch def.Type {
		case FlagString:
			val := new(string)
			stringFlags[def.Name] = val
			cmd.Flags().StringVar(val, def.Name, "", def.Usage)
		case FlagInt:
			val := new(int)
			intFlags[def.Name] = val
			cmd.Flags().IntVar(val, def.Name, 0, def.Usage)
		}
	}
}

// buildRegistryParams builds a params map from flag definitions and their current values.
func buildRegistryParams(cmd *cobra.Command, defs []RegistryFlagDef, stringFlags map[string]*string, intFlags map[string]*int) map[string]interface{} {
	params := make(map[string]interface{})
	for _, def := range defs {
		switch def.Type {
		case FlagString:
			if val, ok := stringFlags[def.Name]; ok && *val != "" {
				params[def.ParamKey] = *val
			}
		case FlagInt:
			if _, ok := intFlags[def.Name]; ok && cmd.Flags().Changed(def.Name) {
				params[def.ParamKey] = *intFlags[def.Name]
			}
		}
	}
	return params
}

func registryCommandName(resource, action string) string {
	switch resource {
	case "area":
		return "config/area_registry/" + action
	case "floor":
		return "config/floor_registry/" + action
	case "label":
		return "config/label_registry/" + action
	default:
		return resource + "/" + action
	}
}
