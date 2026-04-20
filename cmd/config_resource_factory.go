package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

// ConfigResourceConfig defines a REST-backed config resource (automation, script)
// whose CRUD operations follow the same pattern:
//
//	GET    /api/config/{resource}/config/{id}
//	POST   /api/config/{resource}/config/{id}   (create / update)
//	DELETE /api/config/{resource}/config/{id}
type ConfigResourceConfig struct {
	// ParentCmd is the parent cobra command (e.g. automationCmd, scriptCmd).
	ParentCmd *cobra.Command
	// ResourceName is the human-readable name ("automation", "script").
	ResourceName string
	// APIPrefix is the REST path prefix (e.g. "config/automation/config/").
	APIPrefix string
	// IDFlagName is the flag name for the get command (e.g. "automation", "script").
	IDFlagName string
	// ResolveID converts a user-supplied ID to the internal config ID.
	ResolveID func(client.RestAPI, string) (string, error)
	// GroupID is the cobra group ID for the generated commands.
	GroupID string
	// CreateExample is the Cobra Example text for the create command.
	CreateExample string
	// RequiredCreateField is the field that must be present in the create
	// payload. Defaults to "alias" if empty (automation/script behaviour).
	// Set to "name" for resources like scenes.
	RequiredCreateField string
}

// RegisterConfigResourceCRUD generates and registers get, create, update, and
// delete subcommands for a REST config resource.
func RegisterConfigResourceCRUD(cfg ConfigResourceConfig) {
	registerConfigGet(cfg)
	registerConfigCreate(cfg)
	registerConfigUpdate(cfg)
	registerConfigDelete(cfg)
}

func registerConfigGet(cfg ConfigResourceConfig) {
	var flagID string
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("get [%s_id]", cfg.ResourceName),
		Short:   fmt.Sprintf("Get %s configuration", cfg.ResourceName),
		Long:    fmt.Sprintf("Get the full configuration of a %s.", cfg.ResourceName),
		GroupID: cfg.GroupID,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveArg(flagID, args, 0, cfg.ResourceName+" ID")
			if err != nil {
				return err
			}
			textMode := getTextMode()

			restClient, err := getRESTClient()
			if err != nil {
				return err
			}

			configID, err := cfg.ResolveID(restClient, id)
			if err != nil {
				return err
			}

			result, err := restClient.Get(cfg.APIPrefix + configID)
			if err != nil {
				return err
			}

			output.PrintOutputWithContext(result, textMode, "", output.EnvelopeContext{
				Operation:    "get",
				ResourceType: cfg.ResourceName,
			})
			return nil
		},
	}
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		SideEffect:   "read",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ResourceName,
		InputSources: []string{"args", "flags"},
	})
	cmd.Flags().StringVar(&flagID, cfg.IDFlagName, "", fmt.Sprintf("%s ID to get", capitalize(cfg.ResourceName)))
	cfg.ParentCmd.AddCommand(cmd)
}

func registerConfigCreate(cfg ConfigResourceConfig) {
	var inputFlags InputFlags
	var plan bool
	var dryRun bool
	cmd := &cobra.Command{
		Use:     "create <id>",
		Short:   fmt.Sprintf("Create a new %s", cfg.ResourceName),
		Long:    fmt.Sprintf("Create a new %s from JSON or YAML. The ID is used to identify the %s.", cfg.ResourceName, cfg.ResourceName),
		Example: cfg.CreateExample,
		GroupID: cfg.GroupID,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			textMode := getTextMode()

			config, err := inputFlags.Parse()
			if err != nil {
				return err
			}

			requiredField := cfg.RequiredCreateField
			if requiredField == "" {
				requiredField = "alias"
			}
			if _, ok := config[requiredField]; !ok {
				return fmt.Errorf("%s must have a '%s' field", cfg.ResourceName, requiredField)
			}

			if planRequested(plan, dryRun) {
				printMutationPlan(MutationPlan{
					WouldChange: true,
					Target: map[string]any{
						"resource": cfg.ResourceName,
						"id":       id,
						"endpoint": cfg.APIPrefix + id,
					},
					Inputs: map[string]any{"config": config},
					Steps: []string{
						fmt.Sprintf("Validate input payload includes '%s'.", requiredField),
						fmt.Sprintf("POST %s with provided configuration.", cfg.APIPrefix+id),
						"Return created configuration object.",
					},
					VerificationCommands: []string{
						fmt.Sprintf("hab %s get %s --json", cfg.ResourceName, id),
					},
					RequiresConfirmation: false,
				}, textMode, cfg.ResourceName)
				return nil
			}

			restClient, err := getRESTClient()
			if err != nil {
				return err
			}

			result, err := restClient.Post(cfg.APIPrefix+id, config)
			if err != nil {
				return err
			}

			output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("%s %s created successfully.", capitalize(cfg.ResourceName), id), output.EnvelopeContext{
				Operation:    "create",
				ResourceType: cfg.ResourceName,
			})
			return nil
		},
	}
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ResourceName,
		InputSources: []string{"args", "flags", "data", "file"},
	})
	inputFlags.Register(cmd)
	registerMutationPlanFlags(cmd, &plan, &dryRun)
	cfg.ParentCmd.AddCommand(cmd)
}

func registerConfigUpdate(cfg ConfigResourceConfig) {
	var inputFlags InputFlags
	var plan bool
	var dryRun bool
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("update <%s_id>", cfg.ResourceName),
		Short:   fmt.Sprintf("Update an existing %s", cfg.ResourceName),
		Long:    fmt.Sprintf("Update a %s with new configuration.", cfg.ResourceName),
		GroupID: cfg.GroupID,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			textMode := getTextMode()

			config, err := inputFlags.Parse()
			if err != nil {
				return err
			}

			restClient, err := getRESTClient()
			if err != nil {
				return err
			}

			configID, err := cfg.ResolveID(restClient, id)
			if err != nil {
				return err
			}

			if planRequested(plan, dryRun) {
				printMutationPlan(MutationPlan{
					WouldChange: true,
					Target: map[string]any{
						"resource": cfg.ResourceName,
						"id":       id,
						"endpoint": cfg.APIPrefix + configID,
					},
					Inputs: map[string]any{"config": config},
					DerivedIDs: map[string]any{
						"resolved_id": configID,
					},
					Steps: []string{
						"Resolve user-provided identifier to internal config ID.",
						fmt.Sprintf("POST %s with replacement configuration.", cfg.APIPrefix+configID),
						"Return updated configuration object.",
					},
					VerificationCommands: []string{
						fmt.Sprintf("hab %s get %s --json", cfg.ResourceName, id),
					},
					RequiresConfirmation: false,
				}, textMode, cfg.ResourceName)
				return nil
			}

			result, err := restClient.Post(cfg.APIPrefix+configID, config)
			if err != nil {
				return err
			}

			output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("%s updated successfully.", capitalize(cfg.ResourceName)), output.EnvelopeContext{
				Operation:    "update",
				ResourceType: cfg.ResourceName,
			})
			return nil
		},
	}
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ResourceName,
		InputSources: []string{"args", "flags", "data", "file"},
	})
	inputFlags.Register(cmd)
	registerMutationPlanFlags(cmd, &plan, &dryRun)
	cfg.ParentCmd.AddCommand(cmd)
}

func registerConfigDelete(cfg ConfigResourceConfig) {
	var force bool
	var plan bool
	var dryRun bool
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("delete <%s_id>", cfg.ResourceName),
		Short:   fmt.Sprintf("Delete a %s", cfg.ResourceName),
		Long:    fmt.Sprintf("Delete a %s from Home Assistant.", cfg.ResourceName),
		GroupID: cfg.GroupID,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			textMode := getTextMode()

			restClient, err := getRESTClient()
			if err != nil {
				return err
			}

			configID, err := cfg.ResolveID(restClient, id)
			if err != nil {
				return err
			}

			if planRequested(plan, dryRun) {
				printMutationPlan(MutationPlan{
					WouldChange: true,
					Target: map[string]any{
						"resource": cfg.ResourceName,
						"id":       id,
						"endpoint": cfg.APIPrefix + configID,
					},
					DerivedIDs: map[string]any{
						"resolved_id": configID,
					},
					Steps: []string{
						"Resolve user-provided identifier to internal config ID.",
						fmt.Sprintf("DELETE %s.", cfg.APIPrefix+configID),
						"Return deletion success message.",
					},
					Risks: []string{
						"This operation permanently removes the configuration object.",
					},
					RequiresConfirmation: !force,
					VerificationCommands: []string{
						fmt.Sprintf("hab %s get %s --json", cfg.ResourceName, id),
					},
				}, textMode, cfg.ResourceName)
				return nil
			}

			if err := confirmAction(force, fmt.Sprintf("Delete %s %s?", cfg.ResourceName, id), fmt.Sprintf("delete %s", cfg.ResourceName)); err != nil {
				return err
			}

			_, err = restClient.Delete(cfg.APIPrefix + configID)
			if err != nil {
				return err
			}

			output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("%s %s deleted.", capitalize(cfg.ResourceName), id), output.EnvelopeContext{
				Operation:    "delete",
				ResourceType: cfg.ResourceName,
			})
			return nil
		},
	}
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: cfg.ResourceName,
		InputSources: []string{"args", "flags"},
	})
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(cmd, &plan, &dryRun)
	cfg.ParentCmd.AddCommand(cmd)
}
