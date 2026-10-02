package cmd

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var commandDataShapeInventory = auditedDataShapes()

func auditedOutputContract(cmd *cobra.Command, original *SchemaOutputContract) *SchemaOutputContract {
	contract := *original
	contract.Variants = slices.Clone(original.Variants)
	shape, reviewed := commandDataShapeInventory[schemaPath(cmd)]
	for i := range contract.Variants {
		variant := &contract.Variants[i]
		if variant.Data == nil {
			continue
		}
		data := *variant.Data
		variant.Data = &data
		if variant.Name == "full" && reviewed && shape != "stream" {
			data.Type = shape
			data.Open = shape == "object" || shape == "array" || shape == "any"
			if shape == "null" || shape == "string" || shape == "any" || strings.HasPrefix(data.Description, "resource-specific") || strings.HasPrefix(data.Description, "mutation result") {
				data.Fields = nil
				data.Description = "Audited data shape; open payload fields depend on resource, flags, and HA version."
			}
		}
		if variant.Name == "plan" {
			variant.OutputMode = "json_envelope"
			data.Fields = append(slices.Clone(data.Fields), SchemaField{Name: "change_detection", Type: "string", Description: "not_performed: would_change is an assumption, not an observed diff"}, SchemaField{Name: "validation", Type: "string", Description: "scope of local checks; no implicit server validation"})
		}
		if variant.Name == "stream" {
			data.Type = "object"
			data.Description = "One NDJSON record per event (not an enclosing array). Errors may end with a JSON error envelope."
			data.Fields = []SchemaField{{Name: "event", Type: "string", Required: true}, {Name: "data", Type: "string"}, {Name: "code", Type: "number"}}
			variant.Envelope = nil
		}
	}
	// Only advertise variants actually selected by this executable's flags.
	contract.Variants = slices.DeleteFunc(contract.Variants, func(v SchemaOutputVariant) bool {
		return (v.Name == "brief" && cmd.Flags().Lookup("brief") == nil) || (v.Name == "count" && cmd.Flags().Lookup("count") == nil)
	})
	if cmd.Flags().Lookup("plan") != nil && !hasOutputVariant(contract.Variants, "plan") {
		contract.Variants = append(contract.Variants, schemaVariant("plan", inferredResourceType(cmd), "json_envelope", *dataContractForVariant("plan", inferredResourceType(cmd), "write", cmd.Name())))
	}
	if schemaPath(cmd) == "hab esphome validate" {
		contract.OutputMode = "ndjson_stream"
	}
	if schemaPath(cmd) == "hab dashboard patch" {
		for i := range contract.Variants {
			contract.Variants[i].Data = dashboardPatchDataContract()
		}
	}
	return &contract
}

func hasOutputVariant(variants []SchemaOutputVariant, name string) bool {
	return slices.ContainsFunc(variants, func(v SchemaOutputVariant) bool { return v.Name == name })
}

func dashboardPatchDataContract() *SchemaObjectContract {
	return &SchemaObjectContract{Type: "object", Open: true, Description: "Optimistic dashboard patch result; failures include this object at error.details.result.", Fields: []SchemaField{
		{Name: "status", Type: "string", Required: true, Enum: []string{"planned", "noop", "saved", "verified", "failed", "uncertain"}},
		{Name: "dashboard", Type: "string", Description: "Dashboard URL path"},
		{Name: "target", Type: "string", Required: true, Description: "Resolved exact JSON pointer; empty means root"},
		{Name: "base_revision", Type: "string", Description: "Full-config SHA-256 precondition for apply"},
		{Name: "result_revision", Type: "string", Description: "Expected post-patch revision"},
		{Name: "observed_revision", Type: "string", Description: "Last read revision"},
		{Name: "observed_at", Type: "string", Required: true},
		{Name: "would_change", Type: "boolean", Required: true},
		{Name: "saved", Type: "boolean|null", Required: true, Description: "true: save acknowledged; false: not saved; null: acknowledgement uncertain"},
		{Name: "reloaded", Type: "string", Required: true, Enum: []string{"not_applicable"}},
		{Name: "verified", Type: "boolean", Required: true},
		{Name: "changes", Type: "array", ItemType: "object", Required: true, Fields: []SchemaField{{Name: "path", Type: "string"}, {Name: "before", Type: "any"}, {Name: "after", Type: "any"}, {Name: "before_exists", Type: "boolean"}, {Name: "after_exists", Type: "boolean"}}},
		{Name: "change_count", Type: "number", Required: true},
		{Name: "diff_complete", Type: "boolean", Required: true},
		{Name: "requests", Type: "number", Required: true},
		{Name: "validation", Type: "string", Required: true},
		{Name: "limits", Type: "array", ItemType: "string", Required: true},
	}}
}
