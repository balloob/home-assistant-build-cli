package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
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
  hab calendar delete calendar.personal abc123def456 --recurrence-range THISEVENT
  hab calendar delete calendar.personal abc123def456 --recurrence-range THISANDFUTURE`,
	Args: cobra.ExactArgs(2),
	RunE: runCalendarDelete,
}

func init() {
	calendarCmd.AddCommand(calendarDeleteCmd)
	calendarDeleteCmd.Flags().StringVar(&calendarDeleteRecurrenceRange, "recurrence-range", "", "For recurring events: THISEVENT or THISANDFUTURE")
	calendarDeleteCmd.Flags().BoolVarP(&calendarDeleteForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(calendarDeleteCmd, &calendarDeletePlan, &calendarDeleteDryRun)
	annotateServiceMutationCommand(calendarDeleteCmd, "destructive", "calendar_event", []string{"args", "flags"})
}

func runCalendarDelete(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "calendar")
	uid := args[1]
	textMode := getTextMode()

	if planRequested(calendarDeletePlan, calendarDeleteDryRun) {
		inputs := map[string]any{}
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
				"Call Home Assistant service calendar.delete_event with calendar entity_id and event UID.",
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

	data := map[string]interface{}{
		"entity_id": entityID,
		"uid":       uid,
	}
	if calendarDeleteRecurrenceRange != "" {
		data["recurrence_range"] = calendarDeleteRecurrenceRange
	}

	return callServiceAction("delete", "calendar_event", "calendar", "delete_event", "Event deleted.", data)
}
