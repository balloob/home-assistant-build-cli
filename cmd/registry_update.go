package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/registryupdate"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

type registryUpdateStore struct {
	ws  client.WebSocketAPI
	cfg RegistryCRUDConfig
}

func (s registryUpdateStore) Get(id string) (map[string]any, error) {
	entries, err := s.cfg.ListFunc(s.ws)
	if err != nil {
		return nil, err
	}
	for _, item := range entries {
		if entry, ok := item.(map[string]any); ok && entry[s.cfg.IDField] == id {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("%s %q not found", s.cfg.ResourceName, id)
}

func (s registryUpdateStore) Update(id string, params map[string]any) (map[string]any, error) {
	return s.cfg.UpdateFunc(s.ws, id, params)
}

func registerRegistryUpdate(cfg RegistryCRUDConfig) {
	cfg.ParentCmd.AddCommand(newRegistryUpdateCommand(cfg))
}

func newRegistryUpdateCommand(cfg RegistryCRUDConfig) *cobra.Command {
	var plan, dryRun bool
	nameKey := cfg.NameField
	if nameKey == "" {
		nameKey = "name"
	}
	defs := append([]RegistryFlagDef{{Name: "name", ParamKey: nameKey, Usage: "New name (required names cannot be cleared)", Type: FlagString}}, cfg.UpdateFlags...)
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("update <%s>", cfg.IDField),
		Short: fmt.Sprintf("Update a %s", cfg.ResourceName),
		Long:  "Update registry metadata. Omitted fields are preserved. Empty optional values clear them.\n--plan and --dry-run read current values without writing. Apply reads back requested fields.\nThe result reports saved (write acknowledged) and verified separately; verification covers\nregistry metadata, not runtime entity availability. No-op updates do not write.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := registryUpdateParams(cmd, defs, cfg.NullableName)
			if err != nil {
				return err
			}
			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer func() {
				if err := ws.Close(); err != nil {
					noteWarning("Registry connection close failed.")
				}
			}()
			if cfg.ValidateUpdate != nil {
				if err := cfg.ValidateUpdate(ws, params); err != nil {
					return err
				}
			}
			result, err := registryupdate.Run(registryUpdateStore{ws, cfg}, args[0], params, planRequested(plan, dryRun))
			if err != nil {
				return &client.APIError{Code: "REGISTRY_UPDATE_FAILED", Message: err.Error(), Details: map[string]any{"result": result.Output(), "retryable": false}}
			}
			output.PrintOutputWithContext(result.Output(), getTextMode(), "", output.EnvelopeContext{Operation: "update", ResourceType: cfg.ResourceName})
			return nil
		},
	}
	names := make([]string, 0, len(defs))
	for _, def := range defs {
		names = append(names, def.Name)
		switch def.Type {
		case FlagStringSet:
			cmd.Flags().StringSlice(def.Name, nil, def.Usage+" (replaces the set; empty clears)")
		case FlagUserBool:
			cmd.Flags().Bool(def.Name, false, def.Usage+" (use =false to clear)")
		default:
			cmd.Flags().String(def.Name, "", def.Usage+" (empty clears optional values)")
		}
	}
	cmd.MarkFlagsOneRequired(names...)
	registerMutationPlanFlags(cmd, &plan, &dryRun)
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		SideEffect: "write", OutputMode: "json_envelope", Capabilities: []string{"auth", "ws"},
		ResourceType: cfg.ResourceName, InputSources: []string{"args", "flags"}, GuideTopic: "registry",
		FlagConstraints: []SchemaFlagConstraint{{Type: "at_least_one", Flags: names, Description: "Provide one or more fields to update."}},
	})
	return cmd
}

func registryUpdateParams(cmd *cobra.Command, defs []RegistryFlagDef, nullableName bool) (map[string]any, error) {
	params := map[string]any{}
	for _, def := range defs {
		if !cmd.Flags().Changed(def.Name) {
			continue
		}
		value, err := registryUpdateValue(cmd, def)
		if err != nil {
			return nil, err
		}
		if def.Name == "name" && value == nil && !nullableName {
			return nil, fmt.Errorf("--name cannot be empty")
		}
		params[def.ParamKey] = value
	}
	if len(params) == 0 {
		return nil, fmt.Errorf("no update parameters provided")
	}
	return params, nil
}

func registryUpdateValue(cmd *cobra.Command, def RegistryFlagDef) (any, error) {
	switch def.Type {
	case FlagStringSet:
		values, err := cmd.Flags().GetStringSlice(def.Name)
		return registryupdate.NormalizeSet(values, def.ParamKey == "aliases"), err
	case FlagUserBool:
		value, err := cmd.Flags().GetBool(def.Name)
		if value {
			return "user", err
		}
		return nil, err
	default:
		value, err := cmd.Flags().GetString(def.Name)
		if err != nil || value == "" {
			return nil, err
		}
		if def.Type == FlagInt {
			number, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("--%s must be an integer or empty: %w", def.Name, err)
			}
			return number, nil
		}
		return value, nil
	}
}
