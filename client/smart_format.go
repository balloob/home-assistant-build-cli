package client

import (
	"fmt"
	"sort"
	"strings"
)

// SmartFormat formats data in a human-readable way based on its structure.
// It detects the type of data and applies appropriate formatting.
func SmartFormat(data interface{}) string {
	if data == nil {
		return "Done."
	}

	switch v := data.(type) {
	case string:
		return v
	case bool:
		if v {
			return "Yes"
		}
		return "No"
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case []interface{}:
		return smartFormatList(v)
	case map[string]interface{}:
		return smartFormatMap(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// smartFormatList formats a list based on its contents
func smartFormatList(data []interface{}) string {
	if len(data) == 0 {
		return "No items."
	}

	// Check if it's a list of maps
	if _, ok := data[0].(map[string]interface{}); ok {
		return smartFormatMapList(data)
	}

	// Simple list
	var lines []string
	for _, item := range data {
		lines = append(lines, fmt.Sprintf("  - %v", item))
	}
	return strings.Join(lines, "\n")
}

// smartFormatMapList formats a list of maps with smart field selection
func smartFormatMapList(data []interface{}) string {
	if len(data) == 0 {
		return "No items."
	}

	// Detect the type of list based on common fields
	first, ok := data[0].(map[string]interface{})
	if !ok {
		return "No items."
	}

	// Detect list type and format accordingly
	if _, hasUrlPath := first["url_path"]; hasUrlPath {
		// Dashboard list
		return formatDashboardList(data)
	}
	if _, hasEntityId := first["entity_id"]; hasEntityId {
		// Entity list
		return formatEntityList(data)
	}
	if _, hasAlias := first["alias"]; hasAlias {
		if _, hasId := first["id"]; hasId {
			// Automation/Script list
			return formatAutomationList(data)
		}
	}
	if _, hasAreaId := first["area_id"]; hasAreaId {
		if _, hasName := first["name"]; hasName {
			// Area list
			return formatAreaList(data)
		}
	}
	if _, hasFloorId := first["floor_id"]; hasFloorId {
		// Floor list
		return formatFloorList(data)
	}
	if _, hasLabelId := first["label_id"]; hasLabelId {
		// Label list
		return formatLabelList(data)
	}
	if _, hasDeviceId := first["id"]; hasDeviceId {
		if _, hasManufacturer := first["manufacturer"]; hasManufacturer {
			// Device list
			return formatDeviceList(data)
		}
	}

	// Generic list - show key fields
	return formatGenericList(data)
}

// formatDashboardList formats a list of dashboards
func formatDashboardList(data []interface{}) string {
	var lines []string

	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		urlPath := getStringField(m, "url_path", "")
		title := getStringField(m, "title", urlPath)

		lines = append(lines, fmt.Sprintf("%s:", title))

		if urlPath != "" {
			lines = append(lines, fmt.Sprintf("  path: %s", urlPath))
		}
		if mode, ok := m["mode"].(string); ok && mode != "" {
			lines = append(lines, fmt.Sprintf("  mode: %s", mode))
		}
		if requireAdmin, ok := m["require_admin"].(bool); ok && requireAdmin {
			lines = append(lines, "  require_admin: yes")
		}
		if showInSidebar, ok := m["show_in_sidebar"].(bool); ok && !showInSidebar {
			lines = append(lines, "  show_in_sidebar: no")
		}

		lines = append(lines, "")
	}

	// Remove trailing empty line
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}

// formatEntityList formats a list of entities
func formatEntityList(data []interface{}) string {
	var lines []string

	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		entityId := getStringField(m, "entity_id", "")
		state := getStringField(m, "state", "")
		friendlyName := ""

		if attrs, ok := m["attributes"].(map[string]interface{}); ok {
			friendlyName = getStringField(attrs, "friendly_name", "")
		}

		if friendlyName != "" && friendlyName != entityId {
			lines = append(lines, fmt.Sprintf("%s (%s): %s", friendlyName, entityId, state))
		} else {
			lines = append(lines, fmt.Sprintf("%s: %s", entityId, state))
		}
	}

	return strings.Join(lines, "\n")
}

// formatAutomationList formats a list of automations or scripts
func formatAutomationList(data []interface{}) string {
	var lines []string

	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		id := getStringField(m, "id", "")
		alias := getStringField(m, "alias", id)
		state := getStringField(m, "state", "")

		line := alias
		if state != "" {
			line = fmt.Sprintf("%s [%s]", alias, state)
		}
		if id != "" && id != alias {
			line = fmt.Sprintf("%s (id: %s)", line, id)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatAreaList formats a list of areas
func formatAreaList(data []interface{}) string {
	var lines []string

	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		areaId := getStringField(m, "area_id", "")
		name := getStringField(m, "name", areaId)
		floorId := getStringField(m, "floor_id", "")

		line := name
		if floorId != "" {
			line = fmt.Sprintf("%s (floor: %s)", line, floorId)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatFloorList formats a list of floors
func formatFloorList(data []interface{}) string {
	var lines []string

	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		floorId := getStringField(m, "floor_id", "")
		name := getStringField(m, "name", floorId)
		level, hasLevel := m["level"].(float64)

		if hasLevel {
			lines = append(lines, fmt.Sprintf("%s (level: %.0f)", name, level))
		} else {
			lines = append(lines, name)
		}
	}

	return strings.Join(lines, "\n")
}

// formatLabelList formats a list of labels
func formatLabelList(data []interface{}) string {
	var lines []string

	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		labelId := getStringField(m, "label_id", "")
		name := getStringField(m, "name", labelId)
		color := getStringField(m, "color", "")

		if color != "" {
			lines = append(lines, fmt.Sprintf("%s [%s]", name, color))
		} else {
			lines = append(lines, name)
		}
	}

	return strings.Join(lines, "\n")
}

// formatDeviceList formats a list of devices
func formatDeviceList(data []interface{}) string {
	var lines []string

	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name := getStringField(m, "name", "")
		nameByUser := getStringField(m, "name_by_user", "")
		manufacturer := getStringField(m, "manufacturer", "")
		model := getStringField(m, "model", "")

		displayName := nameByUser
		if displayName == "" {
			displayName = name
		}

		details := []string{}
		if manufacturer != "" {
			details = append(details, manufacturer)
		}
		if model != "" {
			details = append(details, model)
		}

		if len(details) > 0 {
			lines = append(lines, fmt.Sprintf("%s (%s)", displayName, strings.Join(details, " ")))
		} else {
			lines = append(lines, displayName)
		}
	}

	return strings.Join(lines, "\n")
}

// formatGenericList formats a generic list of maps
func formatGenericList(data []interface{}) string {
	var lines []string

	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			lines = append(lines, fmt.Sprintf("  - %v", item))
			continue
		}

		name := getDisplayName(m)
		lines = append(lines, fmt.Sprintf("  - %s", name))
	}

	return strings.Join(lines, "\n")
}

// smartFormatMap formats a single map with smart field ordering
func smartFormatMap(data map[string]interface{}) string {
	if len(data) == 0 {
		return "Done."
	}

	// Check for specific types
	if _, hasEntityId := data["entity_id"]; hasEntityId {
		return formatEntityDetail(data)
	}
	if _, hasUrlPath := data["url_path"]; hasUrlPath {
		if _, hasViews := data["views"]; hasViews {
			return formatDashboardConfig(data)
		}
		return formatDashboardDetail(data)
	}
	if _, hasAlias := data["alias"]; hasAlias {
		return formatAutomationDetail(data)
	}
	if _, hasCount := data["count"]; hasCount {
		if len(data) == 1 {
			return fmt.Sprintf("Count: %v", data["count"])
		}
	}

	// Generic map formatting
	return formatGenericMap(data)
}

// formatEntityDetail formats a single entity
func formatEntityDetail(data map[string]interface{}) string {
	var lines []string

	entityId := getStringField(data, "entity_id", "")
	state := getStringField(data, "state", "")

	lines = append(lines, fmt.Sprintf("Entity: %s", entityId))
	lines = append(lines, fmt.Sprintf("State: %s", state))

	if attrs, ok := data["attributes"].(map[string]interface{}); ok && len(attrs) > 0 {
		lines = append(lines, "Attributes:")
		for key, value := range attrs {
			lines = append(lines, fmt.Sprintf("  %s: %s", formatKey(key), formatSmartValue(value)))
		}
	}

	return strings.Join(lines, "\n")
}

// formatDashboardDetail formats a single dashboard item
func formatDashboardDetail(data map[string]interface{}) string {
	var lines []string

	title := getStringField(data, "title", "")
	urlPath := getStringField(data, "url_path", "")

	if title != "" {
		lines = append(lines, fmt.Sprintf("%s:", title))
	} else {
		lines = append(lines, fmt.Sprintf("%s:", urlPath))
	}

	if urlPath != "" {
		lines = append(lines, fmt.Sprintf("  path: %s", urlPath))
	}
	if mode, ok := data["mode"].(string); ok && mode != "" {
		lines = append(lines, fmt.Sprintf("  mode: %s", mode))
	}
	if requireAdmin, ok := data["require_admin"].(bool); ok {
		lines = append(lines, fmt.Sprintf("  require_admin: %s", formatBool(requireAdmin)))
	}
	if showInSidebar, ok := data["show_in_sidebar"].(bool); ok {
		lines = append(lines, fmt.Sprintf("  show_in_sidebar: %s", formatBool(showInSidebar)))
	}

	return strings.Join(lines, "\n")
}

// formatDashboardConfig formats a full dashboard configuration
func formatDashboardConfig(data map[string]interface{}) string {
	var lines []string

	title := getStringField(data, "title", "Dashboard")
	lines = append(lines, fmt.Sprintf("Dashboard: %s", title))

	if views, ok := data["views"].([]interface{}); ok {
		lines = append(lines, fmt.Sprintf("Views: %d", len(views)))
		for i, v := range views {
			if view, ok := v.(map[string]interface{}); ok {
				viewTitle := getStringField(view, "title", fmt.Sprintf("View %d", i+1))
				lines = append(lines, fmt.Sprintf("  - %s", viewTitle))

				// Count cards in this view
				cardCount := 0
				if cards, ok := view["cards"].([]interface{}); ok {
					cardCount = len(cards)
				}
				if sections, ok := view["sections"].([]interface{}); ok {
					for _, s := range sections {
						if section, ok := s.(map[string]interface{}); ok {
							if sCards, ok := section["cards"].([]interface{}); ok {
								cardCount += len(sCards)
							}
						}
					}
				}
				if cardCount > 0 {
					lines = append(lines, fmt.Sprintf("    cards: %d", cardCount))
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// formatAutomationDetail formats a single automation or script
func formatAutomationDetail(data map[string]interface{}) string {
	var lines []string

	alias := getStringField(data, "alias", "")
	id := getStringField(data, "id", "")
	description := getStringField(data, "description", "")

	if alias != "" {
		lines = append(lines, fmt.Sprintf("Name: %s", alias))
	}
	if id != "" {
		lines = append(lines, fmt.Sprintf("ID: %s", id))
	}
	if description != "" {
		lines = append(lines, fmt.Sprintf("Description: %s", description))
	}

	if triggers, ok := data["trigger"].([]interface{}); ok && len(triggers) > 0 {
		lines = append(lines, fmt.Sprintf("Triggers: %d", len(triggers)))
	}
	if conditions, ok := data["condition"].([]interface{}); ok && len(conditions) > 0 {
		lines = append(lines, fmt.Sprintf("Conditions: %d", len(conditions)))
	}
	if actions, ok := data["action"].([]interface{}); ok && len(actions) > 0 {
		lines = append(lines, fmt.Sprintf("Actions: %d", len(actions)))
	}

	// For scripts, show sequence
	if sequence, ok := data["sequence"].([]interface{}); ok && len(sequence) > 0 {
		lines = append(lines, fmt.Sprintf("Steps: %d", len(sequence)))
	}

	return strings.Join(lines, "\n")
}

// formatGenericMap formats a generic map
func formatGenericMap(data map[string]interface{}) string {
	var lines []string

	// Sort keys for consistent output
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := data[key]
		keyLabel := formatKey(key)

		switch v := value.(type) {
		case map[string]interface{}:
			lines = append(lines, fmt.Sprintf("%s:", keyLabel))
			for k, val := range v {
				lines = append(lines, fmt.Sprintf("  %s: %s", k, formatSmartValue(val)))
			}
		case []interface{}:
			if len(v) == 0 {
				lines = append(lines, fmt.Sprintf("%s: (none)", keyLabel))
			} else if len(v) <= 5 {
				lines = append(lines, fmt.Sprintf("%s:", keyLabel))
				for _, item := range v {
					if m, ok := item.(map[string]interface{}); ok {
						lines = append(lines, fmt.Sprintf("  - %s", getDisplayName(m)))
					} else {
						lines = append(lines, fmt.Sprintf("  - %v", item))
					}
				}
			} else {
				lines = append(lines, fmt.Sprintf("%s: %d items", keyLabel, len(v)))
			}
		default:
			lines = append(lines, fmt.Sprintf("%s: %s", keyLabel, formatSmartValue(value)))
		}
	}

	return strings.Join(lines, "\n")
}

// Helper functions

func getStringField(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}

func formatBool(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func formatSmartValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return "(none)"
	case string:
		if val == "" {
			return "(empty)"
		}
		return val
	case bool:
		return formatBool(val)
	case map[string]interface{}, []interface{}:
		return "..."
	default:
		return fmt.Sprintf("%v", val)
	}
}
