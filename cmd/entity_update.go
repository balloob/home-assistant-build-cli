package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/entityupdate"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

func init() { entityCmd.AddCommand(newEntityUpdateCommand()) }

func newEntityUpdateCommand() *cobra.Command {
	var newID, name, icon string
	var plan, dryRun bool
	cmd := &cobra.Command{
		Use:   "update <entity_id>",
		Short: "Update an entity's ID, friendly name, or icon",
		Long: `Update an entity's registry ID, friendly name, or icon, keeping its unique_id.
Provide at least one of --entity-id, --name, or --icon; flags can be combined.
Use --name "" or --icon "" to clear an override and restore the default.

Use --plan to preview each requested field without writing.
ID changes also check collisions and preview related items.
References are not rewritten. Relationship search does not cover dashboards,
templates, or all dynamic references. Recorder history/statistics migration is
handled by Home Assistant and is not verified by this command.

The result reports saved (write acknowledged) and verified (new registry entry
read back with the requested values and unchanged unique_id) separately.`,
		Example: `  hab entity update switch.plug_2 --entity-id switch.kitchen_plug --plan
  hab entity update switch.plug_2 --entity-id switch.kitchen_plug --name "Kitchen plug" --icon mdi:power-socket-eu --json
  hab entity update switch.kitchen_plug --icon ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := entityupdate.Options{}
			if cmd.Flags().Changed("entity-id") {
				opts.EntityID = &newID
			}
			if cmd.Flags().Changed("name") {
				opts.Name = &name
			}
			if cmd.Flags().Changed("icon") {
				opts.Icon = &icon
			}
			if err := entityupdate.Validate(args[0], opts); err != nil {
				return err
			}
			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer func() {
				if err := ws.Close(); err != nil {
					noteWarning("Entity registry connection close failed.")
				}
			}()
			result, err := entityupdate.Run(ws, args[0], opts, planRequested(plan, dryRun))
			if err != nil {
				return &client.APIError{Code: "ENTITY_UPDATE_FAILED", Message: err.Error(), Details: map[string]any{"result": result, "retryable": false}}
			}
			output.PrintOutputWithContext(result, getTextMode(), "", output.EnvelopeContext{Operation: "update", ResourceType: "entity"})
			return nil
		},
	}
	cmd.Flags().StringVar(&newID, "entity-id", "", "New entity ID (must keep the current domain)")
	cmd.Flags().StringVar(&name, "name", "", "Friendly name override (empty clears the override)")
	cmd.Flags().StringVar(&icon, "icon", "", "Icon override, e.g. mdi:lightbulb (empty clears the override)")
	cmd.MarkFlagsOneRequired("entity-id", "name", "icon")
	registerMutationPlanFlags(cmd, &plan, &dryRun)
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		SideEffect: "write", Capabilities: []string{"auth", "ws"}, ResourceType: "entity",
		FlagConstraints: []SchemaFlagConstraint{{Type: "at_least_one", Flags: []string{"entity-id", "name", "icon"}, Description: "Provide one or more fields to update; empty name/icon values clear overrides."}},
	})
	return cmd
}
