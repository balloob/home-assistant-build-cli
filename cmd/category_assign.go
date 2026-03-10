package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	categoryAssignCategoryID string
	categoryAssignEntityID   string
	categoryAssignScope      string
)

// helperEntityDomains is the set of entity domains that belong to the helpers scope.
var helperEntityDomains = map[string]bool{
	"input_boolean":  true,
	"input_number":   true,
	"input_text":     true,
	"input_select":   true,
	"input_datetime": true,
	"input_button":   true,
	"counter":        true,
	"timer":          true,
	"schedule":       true,
	"derivative":     true,
	"integration":    true,
	"min_max":        true,
	"threshold":      true,
	"utility_meter":  true,
	"statistics":     true,
}

// inferCategoryScope infers the category scope from an entity_id prefix.
// Returns empty string if the scope cannot be determined automatically.
func inferCategoryScope(entityID string) string {
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) == 0 {
		return ""
	}
	domain := parts[0]
	switch domain {
	case "automation":
		return "automation"
	case "script":
		return "script"
	case "scene":
		return "scene"
	default:
		if helperEntityDomains[domain] {
			return "helpers"
		}
	}
	return ""
}

var categoryAssignCmd = &cobra.Command{
	Use:   "assign <category_id> <entity_id>",
	Short: "Assign a category to an entity",
	Long: `Assign a category to an automation, script, scene, or helper entity.

The scope is inferred from the entity_id prefix:
  - automation.* → scope "automation"
  - script.*     → scope "script"
  - scene.*      → scope "scene"
  - helper domains (input_boolean, counter, etc.) → scope "helpers"

Use --scope to override for ambiguous cases.`,
	Example: `  hab category assign abc123 automation.evening_lights
  hab category assign abc123 script.morning_routine
  hab category assign abc123 input_boolean.guest_mode --scope helpers`,
	Args: cobra.MaximumNArgs(2),
	RunE: runCategoryAssign,
}

func init() {
	categoryCmd.AddCommand(categoryAssignCmd)
	categoryAssignCmd.Flags().StringVar(&categoryAssignCategoryID, "category", "", "Category ID to assign")
	categoryAssignCmd.Flags().StringVar(&categoryAssignEntityID, "entity", "", "Entity ID to assign the category to")
	categoryAssignCmd.Flags().StringVar(&categoryAssignScope, "scope", "", "Override scope: automation, script, scene, helpers")
}

func runCategoryAssign(cmd *cobra.Command, args []string) error {
	categoryID, err := resolveArg(categoryAssignCategoryID, args, 0, "category ID")
	if err != nil {
		return err
	}
	entityID, err := resolveArg(categoryAssignEntityID, args, 1, "entity ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	scope := categoryAssignScope
	if scope == "" {
		scope = inferCategoryScope(entityID)
	}
	if scope == "" {
		return fmt.Errorf("cannot infer scope from entity_id '%s'. Use --scope to specify: automation, script, scene, helpers", entityID)
	}
	if !validCategoryScopes[scope] {
		return fmt.Errorf("invalid scope '%s'. Valid values: automation, script, scene, helpers", scope)
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	// categories is a map of scope → category_id in the entity registry
	result, err := ws.EntityRegistryUpdate(entityID, map[string]interface{}{
		"categories": map[string]interface{}{
			scope: categoryID,
		},
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Category '%s' assigned to '%s' (scope: %s).", categoryID, entityID, scope))
	return nil
}
