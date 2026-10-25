package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/entityrename"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

func init() { entityCmd.AddCommand(newEntityUpdateCommand()) }

func newEntityUpdateCommand() *cobra.Command {
	var newID string
	var plan, dryRun bool
	cmd := &cobra.Command{
		Use:   "update <entity_id>",
		Short: "Change an entity's registry ID without changing its unique_id",
		Long: `Change an entity's registry ID, keeping its domain and unique_id.

Use --plan to check collisions and preview related items without writing.
References are not rewritten. Relationship search does not cover dashboards,
templates, or all dynamic references. Recorder history/statistics migration is
handled by Home Assistant and is not verified by this command.

The result reports saved (write acknowledged) and verified (new registry entry
read back with unchanged unique_id) separately. Use entity rename for names.`,
		Example: `  hab entity update switch.plug_2 --entity-id switch.kitchen_plug --plan
  hab entity update switch.plug_2 --entity-id switch.kitchen_plug --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := entityrename.Validate(args[0], newID); err != nil {
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
			result, err := entityrename.Run(ws, args[0], newID, planRequested(plan, dryRun))
			if err != nil {
				return &client.APIError{Code: "ENTITY_UPDATE_FAILED", Message: err.Error(), Details: map[string]any{"result": result, "retryable": false}}
			}
			output.PrintOutputWithContext(result, getTextMode(), "", output.EnvelopeContext{Operation: "update", ResourceType: "entity"})
			return nil
		},
	}
	cmd.Flags().StringVar(&newID, "entity-id", "", "New entity ID (must keep the current domain)")
	if err := cmd.MarkFlagRequired("entity-id"); err != nil {
		panic(err) // A missing flag here is a command registration bug.
	}
	registerMutationPlanFlags(cmd, &plan, &dryRun)
	mergeSchemaAnnotation(cmd, SchemaAnnotation{SideEffect: "write", Capabilities: []string{"auth", "ws"}, ResourceType: "entity"})
	return cmd
}
