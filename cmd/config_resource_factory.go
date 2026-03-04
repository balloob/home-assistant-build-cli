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

			output.PrintOutput(result, textMode, "")
			return nil
		},
	}
	cmd.Flags().StringVar(&flagID, cfg.IDFlagName, "", fmt.Sprintf("%s ID to get", capitalize(cfg.ResourceName)))
	cfg.ParentCmd.AddCommand(cmd)
}

func registerConfigCreate(cfg ConfigResourceConfig) {
	var inputFlags InputFlags
	cmd := &cobra.Command{
		Use:     "create <id>",
		Short:   fmt.Sprintf("Create a new %s", cfg.ResourceName),
		Long:    fmt.Sprintf("Create a new %s from JSON or YAML. The ID is used to identify the %s.", cfg.ResourceName, cfg.ResourceName),
		GroupID: cfg.GroupID,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			textMode := getTextMode()

			config, err := inputFlags.Parse()
			if err != nil {
				return err
			}

			if _, ok := config["alias"]; !ok {
				return fmt.Errorf("%s must have an 'alias' field", cfg.ResourceName)
			}

			restClient, err := getRESTClient()
			if err != nil {
				return err
			}

			result, err := restClient.Post(cfg.APIPrefix+id, config)
			if err != nil {
				return err
			}

			output.PrintSuccess(result, textMode, fmt.Sprintf("%s %s created successfully.", capitalize(cfg.ResourceName), id))
			return nil
		},
	}
	inputFlags.Register(cmd)
	cfg.ParentCmd.AddCommand(cmd)
}

func registerConfigUpdate(cfg ConfigResourceConfig) {
	var inputFlags InputFlags
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

			result, err := restClient.Post(cfg.APIPrefix+configID, config)
			if err != nil {
				return err
			}

			output.PrintSuccess(result, textMode, fmt.Sprintf("%s updated successfully.", capitalize(cfg.ResourceName)))
			return nil
		},
	}
	inputFlags.Register(cmd)
	cfg.ParentCmd.AddCommand(cmd)
}

func registerConfigDelete(cfg ConfigResourceConfig) {
	var force bool
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

			if !confirmAction(force, textMode, fmt.Sprintf("Delete %s %s?", cfg.ResourceName, id)) {
				fmt.Println("Cancelled.")
				return nil
			}

			_, err = restClient.Delete(cfg.APIPrefix + configID)
			if err != nil {
				return err
			}

			output.PrintSuccess(nil, textMode, fmt.Sprintf("%s %s deleted.", capitalize(cfg.ResourceName), id))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	cfg.ParentCmd.AddCommand(cmd)
}
