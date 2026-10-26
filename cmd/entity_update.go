package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/entityupdate"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

func init() { entityCmd.AddCommand(newEntityUpdateCommand()) }

func newEntityUpdateCommand() *cobra.Command {
	var plan, dryRun bool
	cmd := &cobra.Command{
		Use:   "update <entity_id>",
		Short: "Update an entity's registry settings",
		Long: `Update an entity's ID, name, icon, area, labels, aliases, categories,
device class, or hidden/disabled status, keeping its unique_id.
Provide at least one field flag; flags can be combined.
Empty name, icon, area, or device-class values clear the override.
Labels and aliases replace the full list; an empty value clears it.
Repeat --category scope=id to change category assignments; scope= removes one.
Use --hidden=false or --disabled=false to unhide or enable an entity.
Integration-specific options are not supported.

Use --plan to preview each requested field without writing.
ID changes also check collisions and preview related items.
References are not rewritten. Relationship search does not cover dashboards,
templates, or all dynamic references. Recorder history/statistics migration is
handled by Home Assistant and is not verified by this command.

The result reports saved (write acknowledged) and verified (new registry entry
read back with the requested values and unchanged unique_id) separately.`,
		Example: `  hab entity update switch.plug_2 --entity-id switch.kitchen_plug --plan
  hab entity update switch.plug_2 --entity-id switch.kitchen_plug --name "Kitchen plug" --icon mdi:power-socket-eu --json
  hab entity update switch.kitchen_plug --area kitchen --labels lighting,downstairs --hidden=false
  hab entity update input_boolean.guest_mode --category helpers=guest_settings
  hab entity update switch.kitchen_plug --icon ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := parseEntityUpdateOptions(cmd)
			if err != nil {
				return err
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
	registerEntityUpdateFlags(cmd)
	registerMutationPlanFlags(cmd, &plan, &dryRun)
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		SideEffect: "write", Capabilities: []string{"auth", "ws"}, ResourceType: "entity",
		FlagConstraints: []SchemaFlagConstraint{{Type: "at_least_one", Flags: entityUpdateFieldFlags, Description: "Provide one or more registry fields to update; omitted fields are preserved."}},
	})
	return cmd
}
