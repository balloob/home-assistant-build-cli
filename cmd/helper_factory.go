package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

// HelperCategory represents the API pattern used by a helper type.
type HelperCategory int

const (
	// HelperCategoryWS uses WebSocket for list/create/delete.
	HelperCategoryWS HelperCategory = iota
	// HelperCategoryConfigFlow uses WS for list and REST config flow for create/delete.
	HelperCategoryConfigFlow
)

// HelperDef defines a helper type for factory registration.
type HelperDef struct {
	// TypeName is the HA internal type name (e.g., "input_boolean", "derivative").
	TypeName string
	// CommandName is the cobra command name (e.g., "input-boolean", "derivative").
	CommandName string
	// DisplayName is the human-readable name (e.g., "input boolean", "derivative sensor").
	DisplayName string

	// Short and Long descriptions for the parent command.
	Short string
	Long  string

	// Category selects the API pattern for list/delete.
	Category HelperCategory

	// Create command (all optional — if RunCreate is nil, no create subcommand is registered).
	CreateShort    string
	CreateLong     string
	SetupFlags     func(cmd *cobra.Command)
	RunCreate      func(cmd *cobra.Command, args []string) error
	RequiredFlags  []string // Flag names to mark as required
}

// registerHelperType creates and registers parent, list, delete, and optionally create
// subcommands for a helper type under the helperCmd parent.
func registerHelperType(def HelperDef) {
	parentCmd := &cobra.Command{
		Use:     def.CommandName,
		Short:   def.Short,
		Long:    def.Long,
		GroupID: helperGroupSubcommands,
	}
	helperCmd.AddCommand(parentCmd)

	// List subcommand
	registerHelperList(parentCmd, def)

	// Delete subcommand
	registerHelperDelete(parentCmd, def)

	// Create subcommand (optional)
	if def.RunCreate != nil {
		createCmd := &cobra.Command{
			Use:   "create <name>",
			Short: def.CreateShort,
			Long:  def.CreateLong,
			Args:  cobra.ExactArgs(1),
			RunE:  def.RunCreate,
		}
		if def.SetupFlags != nil {
			def.SetupFlags(createCmd)
		}
		for _, flag := range def.RequiredFlags {
			createCmd.MarkFlagRequired(flag)
		}
		parentCmd.AddCommand(createCmd)
	}
}

// registerHelperList creates the "list" subcommand for a helper type.
func registerHelperList(parentCmd *cobra.Command, def HelperDef) {
	var lf *ListFlags

	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List all %s helpers", def.DisplayName),
		Long:  fmt.Sprintf("List all %s helpers.", def.DisplayName),
		RunE: func(cmd *cobra.Command, args []string) error {
			textMode := getTextMode()

			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			if def.Category == HelperCategoryWS {
				return runWSList(ws, def, textMode, lf)
			}
			return runConfigFlowList(ws, def, textMode, lf)
		},
	}

	lf = RegisterListFlags(listCmd, "id")
	parentCmd.AddCommand(listCmd)
}

// runWSList handles the list command for WS-based helpers.
func runWSList(ws client.WebSocketAPI, def HelperDef, textMode bool, lf *ListFlags) error {
	helpers, err := ws.HelperList(def.TypeName)
	if err != nil {
		return err
	}

	if lf.RenderCount(len(helpers), textMode) {
		return nil
	}

	helpers = lf.ApplyLimit(helpers)

	if lf.RenderBrief(helpers, textMode, "id", "name") {
		return nil
	}

	output.PrintOutput(helpers, textMode, "")
	return nil
}

// runConfigFlowList handles the list command for config-flow-based helpers.
func runConfigFlowList(ws client.WebSocketAPI, def HelperDef, textMode bool, lf *ListFlags) error {
	entries, err := ws.ConfigEntriesList(def.TypeName)
	if err != nil {
		return err
	}

	var result []map[string]interface{}
	for _, e := range entries {
		entry, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		item := map[string]interface{}{
			"entry_id": entry["entry_id"],
			"title":    entry["title"],
		}
		if domain, ok := entry["domain"].(string); ok {
			item["domain"] = domain
		}
		result = append(result, item)
	}

	if lf.RenderCount(len(result), textMode) {
		return nil
	}

	result = lf.ApplyLimitMap(result)

	if lf.RenderBriefMap(result, textMode, "entry_id", "title") {
		return nil
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

// registerHelperDelete creates the "delete" subcommand for a helper type.
func registerHelperDelete(parentCmd *cobra.Command, def HelperDef) {
	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: fmt.Sprintf("Delete a %s helper", def.DisplayName),
		Long:  fmt.Sprintf("Delete a %s helper by entity ID or ID.", def.DisplayName),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			textMode := getTextMode()

			if def.Category == HelperCategoryWS {
				return runWSDelete(id, def, textMode)
			}
			return runConfigFlowDelete(id, def, textMode)
		},
	}
	parentCmd.AddCommand(deleteCmd)
}

// runWSDelete handles the delete command for WS-based helpers.
func runWSDelete(id string, def HelperDef, textMode bool) error {
	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.DeleteHelperByEntityOrEntryID(id, def.TypeName); err != nil {
		return fmt.Errorf("failed to delete %s: %w", def.DisplayName, err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}
	output.PrintSuccess(result, textMode, fmt.Sprintf("%s '%s' deleted successfully.", capitalize(def.DisplayName), id))
	return nil
}

// runConfigFlowDelete handles the delete command for config-flow-based helpers.
func runConfigFlowDelete(id string, def HelperDef, textMode bool) error {
	rest, err := getRESTClient()
	if err != nil {
		return err
	}

	entryID := id
	if strings.Contains(id, ".") {
		// Entity ID — resolve to config entry ID via WebSocket
		ws, err := getWSClient()
		if err != nil {
			return err
		}
		defer ws.Close()

		resolved, err := ws.ResolveEntityToConfigEntry(id)
		if err != nil {
			return fmt.Errorf("failed to resolve entity_id: %w", err)
		}
		if resolved == "" {
			return fmt.Errorf("entity %s does not have a config entry", id)
		}
		entryID = resolved
	}

	if err := rest.ConfigEntryDelete(entryID); err != nil {
		return fmt.Errorf("failed to delete %s: %w", def.DisplayName, err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}
	output.PrintSuccess(result, textMode, fmt.Sprintf("%s '%s' deleted successfully.", capitalize(def.DisplayName), id))
	return nil
}

// helperWSCreate returns a RunE function for simple WS-based helper create commands.
// The buildParams callback builds the WS params map from command flags.
func helperWSCreate(typeName, displayName string, buildParams func(cmd *cobra.Command, name string) (map[string]interface{}, error)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		name := args[0]
		textMode := getTextMode()

		ws, err := getWSClient()
		if err != nil {
			return err
		}
		defer ws.Close()

		params, err := buildParams(cmd, name)
		if err != nil {
			return err
		}

		result, err := ws.HelperCreate(typeName, params)
		if err != nil {
			return err
		}

		output.PrintSuccess(result, textMode, fmt.Sprintf("%s '%s' created successfully.", capitalize(displayName), name))
		return nil
	}
}

// helperConfigFlowCreate returns a RunE function for simple single-step config flow
// helper create commands. The buildFormData callback builds the form data map.
func helperConfigFlowCreate(typeName, displayName string, buildFormData func(cmd *cobra.Command, name string) (map[string]interface{}, error)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		name := args[0]
		textMode := getTextMode()

		rest, err := getRESTClient()
		if err != nil {
			return err
		}

		flowResult, err := rest.ConfigFlowCreate(typeName)
		if err != nil {
			return fmt.Errorf("failed to start config flow: %w", err)
		}

		flowID, ok := flowResult["flow_id"].(string)
		if !ok {
			return fmt.Errorf("no flow_id in response")
		}

		formData, err := buildFormData(cmd, name)
		if err != nil {
			return err
		}

		finalResult, err := rest.ConfigFlowStep(flowID, formData)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", displayName, err)
		}

		resultType, _ := finalResult["type"].(string)
		if resultType == "abort" {
			reason, _ := finalResult["reason"].(string)
			return fmt.Errorf("config flow aborted: %s", reason)
		}
		if resultType == "form" {
			if errors, ok := finalResult["errors"].(map[string]interface{}); ok && len(errors) > 0 {
				return fmt.Errorf("validation error: %v", errors)
			}
			return fmt.Errorf("unexpected form step required: %v", finalResult)
		}
		if resultType != "create_entry" {
			return fmt.Errorf("unexpected flow result type: %s", resultType)
		}

		result := map[string]interface{}{
			"title": finalResult["title"],
		}
		if entryResult, ok := finalResult["result"].(map[string]interface{}); ok {
			if entryID, ok := entryResult["entry_id"]; ok {
				result["entry_id"] = entryID
			}
		}

		output.PrintSuccess(result, textMode, fmt.Sprintf("%s '%s' created successfully.", capitalize(displayName), name))
		return nil
	}
}

// parseDuration converts HH:MM:SS format to a duration map for config flows.
func parseDuration(s string) map[string]interface{} {
	hours, minutes, seconds := 0, 0, 0
	fmt.Sscanf(s, "%d:%d:%d", &hours, &minutes, &seconds)
	return map[string]interface{}{
		"hours":   hours,
		"minutes": minutes,
		"seconds": seconds,
	}
}

// capitalize returns the string with the first letter uppercased.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
