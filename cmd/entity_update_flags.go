package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/internal/entityupdate"
	"github.com/spf13/cobra"
)

var entityUpdateFieldFlags = []string{"entity-id", "name", "icon", "area", "labels", "aliases", "category", "device-class", "hidden", "disabled"}

func registerEntityUpdateFlags(cmd *cobra.Command) {
	for name, usage := range map[string]string{
		"entity-id":    "New entity ID (must keep the current domain)",
		"name":         "Friendly name override (empty clears the override)",
		"icon":         "Icon override, e.g. mdi:lightbulb (empty clears the override)",
		"area":         "Area ID override (empty restores the device's area)",
		"device-class": "Device class override supported by the entity's domain (empty clears it)",
	} {
		cmd.Flags().String(name, "", usage)
	}
	cmd.Flags().StringSlice("labels", nil, "Replace label IDs (comma-separated or repeated; empty clears all)")
	cmd.Flags().StringSlice("aliases", nil, "Replace ordered aliases (CSV or repeated; empty clears all)")
	cmd.Flags().StringArray("category", nil, "Set scope=category_id (repeatable; scope= removes that scope)")
	cmd.Flags().Bool("hidden", false, "Hide the entity; --hidden=false unhides it")
	cmd.Flags().Bool("disabled", false, "Disable the entity; --disabled=false enables it (may require reload/restart)")
	cmd.MarkFlagsOneRequired(entityUpdateFieldFlags...)
}

func parseEntityUpdateOptions(cmd *cobra.Command) (entityupdate.Options, error) {
	opts := entityupdate.Options{}
	for flag, target := range map[string]**string{"entity-id": &opts.EntityID, "name": &opts.Name, "icon": &opts.Icon, "area": &opts.AreaID, "device-class": &opts.DeviceClass} {
		if cmd.Flags().Changed(flag) {
			value, err := cmd.Flags().GetString(flag)
			if err != nil {
				return opts, err
			}
			*target = &value
		}
	}
	for flag, target := range map[string]**[]string{"labels": &opts.Labels, "aliases": &opts.Aliases} {
		if cmd.Flags().Changed(flag) {
			value, err := cmd.Flags().GetStringSlice(flag)
			if err != nil {
				return opts, err
			}
			*target = &value
		}
	}
	for flag, target := range map[string]**bool{"hidden": &opts.Hidden, "disabled": &opts.Disabled} {
		if cmd.Flags().Changed(flag) {
			value, err := cmd.Flags().GetBool(flag)
			if err != nil {
				return opts, err
			}
			*target = &value
		}
	}
	if cmd.Flags().Changed("category") {
		values, err := cmd.Flags().GetStringArray("category")
		if err != nil {
			return opts, err
		}
		opts.Categories, err = parseEntityCategories(values)
		if err != nil {
			return opts, err
		}
	}
	return opts, nil
}

func parseEntityCategories(values []string) (map[string]string, error) {
	result := map[string]string{}
	for _, value := range values {
		scope, id, ok := strings.Cut(value, "=")
		if !ok || !validCategoryScopes[scope] {
			return nil, fmt.Errorf("invalid category %q: use scope=category_id with scope automation, script, scene, or helpers", value)
		}
		if _, exists := result[scope]; exists {
			return nil, fmt.Errorf("category scope %q was specified more than once", scope)
		}
		result[scope] = id
	}
	return result, nil
}
