package cmd

import (
	"github.com/spf13/cobra"
)

// validCategoryScopes is the set of accepted scope values for category operations.
var validCategoryScopes = map[string]bool{
	"automation": true,
	"script":     true,
	"scene":      true,
	"helpers":    true,
}

var categoryCmd = &cobra.Command{
	Use:     "category",
	Short:   "Manage categories",
	Long:    `Create, update, delete, and assign categories for automations, scripts, scenes, and helpers.`,
	GroupID: "automation",
}

func init() {
	rootCmd.AddCommand(categoryCmd)
}
