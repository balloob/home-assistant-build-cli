package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	calendarDeleteRecurrenceID    string
	calendarDeleteRecurrenceRange string
	calendarDeleteForce           bool
	calendarDeletePlan            bool
	calendarDeleteDryRun          bool
)

var calendarDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id> <uid>",
	Short: "Delete a calendar event",
	Long:  `Delete an event from a Home Assistant calendar by its uid.`,
	Example: `  hab calendar delete calendar.personal abc123def456
  hab calendar delete calendar.personal abc123def456 --recurrence-id 20990601T100000
  hab calendar delete calendar.personal abc123def456 --recurrence-id 20990601T100000 --recurrence-range THISANDFUTURE`,
	Args: cobra.ExactArgs(2),
	RunE: runCalendarDelete,
}

func init() {
	calendarCmd.AddCommand(calendarDeleteCmd)
	calendarDeleteCmd.Flags().StringVar(&calendarDeleteRecurrenceID, "recurrence-id", "", "For recurring events: the recurrence_id of the instance to delete")
	calendarDeleteCmd.Flags().StringVar(&calendarDeleteRecurrenceRange, "recurrence-range", "", "For recurring events: THISEVENT or THISANDFUTURE")
	calendarDeleteCmd.Flags().BoolVarP(&calendarDeleteForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(calendarDeleteCmd, &calendarDeletePlan, &calendarDeleteDryRun)
	mergeSchemaAnnotation(calendarDeleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "calendar_event",
		InputSources: []string{"args", "flags"},
	})
}

func runCalendarDelete(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "calendar")
	uid := args[1]
	textMode := getTextMode()

	if planRequested(calendarDeletePlan, calendarDeleteDryRun) {
		inputs := map[string]any{}
		if calendarDeleteRecurrenceID != "" {
			inputs["recurrence_id"] = calendarDeleteRecurrenceID
		}
		if calendarDeleteRecurrenceRange != "" {
			inputs["recurrence_range"] = calendarDeleteRecurrenceRange
		}
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":  "calendar_event",
				"entity_id": entityID,
				"uid":       uid,
			},
			Inputs: inputs,
			Steps: []string{
				"Send WebSocket command calendar/event/delete with calendar entity_id and event UID.",
				"Return deletion confirmation message.",
			},
			Risks: []string{
				"This operation permanently removes the calendar event.",
			},
			RequiresConfirmation: !calendarDeleteForce,
			VerificationCommands: []string{
				fmt.Sprintf("hab calendar list %s --json", entityID),
			},
		}, textMode, "calendar_event")
		return nil
	}

	if err := confirmAction(calendarDeleteForce, fmt.Sprintf("Delete calendar event %s from %s?", uid, entityID), "delete calendar event"); err != nil {
		return err
	}

	params := map[string]interface{}{
		"entity_id": entityID,
		"uid":       uid,
	}
	if calendarDeleteRecurrenceID != "" {
		params["recurrence_id"] = calendarDeleteRecurrenceID
	}
	if calendarDeleteRecurrenceRange != "" {
		params["recurrence_range"] = calendarDeleteRecurrenceRange
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if _, err := ws.SendCommand("calendar/event/delete", params); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, "Event deleted.", output.EnvelopeContext{Operation: "delete", ResourceType: "calendar_event"})
	return nil
}
