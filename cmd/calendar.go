package cmd

import (
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Manage calendar events",
	Long: `List, create, update, and delete calendar events.

For workflow guidance, run 'hab guide calendar-todo'.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(calendarCmd)
}
