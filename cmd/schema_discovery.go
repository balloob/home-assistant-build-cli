package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const commandSchemaVersion = "hab.command.v1.1"
const envelopeSchemaVersion = "hab.envelope.v1"

type commandIndexEntry struct {
	Path       string   `json:"path"`
	Summary    string   `json:"summary"`
	SideEffect string   `json:"side_effect"`
	Transports []string `json:"transports"`
	Preview    string   `json:"preview"`
}

type commandIndex struct {
	SchemaVersion string              `json:"schema_version"`
	CLIVersion    string              `json:"cli_version"`
	EnvelopeRef   string              `json:"envelope_ref"`
	Commands      []commandIndexEntry `json:"commands"`
	Total         int                 `json:"total"`
	Offset        int                 `json:"offset"`
	Complete      bool                `json:"complete"`
	NextOffset    *int                `json:"next_offset,omitempty"`
}

func schemaResult(cmd, target *cobra.Command) (any, error) {
	index, err := cmd.Flags().GetBool("index")
	if err != nil {
		return nil, err
	}
	compact, err := cmd.Flags().GetBool("compact")
	if err != nil {
		return nil, err
	}
	if index {
		if compact {
			return nil, fmt.Errorf("--compact and --index are mutually exclusive")
		}
		query, err := cmd.Flags().GetString("search")
		if err != nil {
			return nil, err
		}
		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return nil, err
		}
		offset, err := cmd.Flags().GetInt("offset")
		if err != nil {
			return nil, err
		}
		return buildCommandIndex(target, query, offset, limit)
	}
	for _, flag := range []string{"search", "limit", "offset"} {
		if cmd.Flags().Changed(flag) {
			return nil, fmt.Errorf("--%s requires --index", flag)
		}
	}
	schema := buildCommandSchema(target, !compact)
	if compact {
		return compactCommandSchema(schema)
	}
	return schema, nil
}

func buildCommandIndex(target *cobra.Command, query string, offset, limit int) (commandIndex, error) {
	result := commandIndex{SchemaVersion: commandSchemaVersion, CLIVersion: Version, EnvelopeRef: envelopeSchemaVersion, Offset: offset, Commands: []commandIndexEntry{}}
	if offset < 0 || limit < 1 || limit > 500 {
		return result, fmt.Errorf("invalid index bounds: offset >= 0 and limit 1-500 required")
	}
	var visit func(*cobra.Command)
	visit = func(cmd *cobra.Command) {
		if cmd.Hidden {
			return
		}
		text := strings.ToLower(schemaPath(cmd) + " " + cmd.Short + " " + strings.Join(cmd.Aliases, " "))
		matches := true
		for _, word := range strings.Fields(strings.ToLower(query)) {
			matches = matches && strings.Contains(text, word)
		}
		if cmd.Runnable() && matches {
			effect := getSchemaAnnotation(cmd).SideEffect
			if effect == "" {
				effect = inferSideEffect(cmd)
			}
			result.Commands = append(result.Commands, commandIndexEntry{schemaPath(cmd), cmd.Short, effect, commandTransports(cmd), commandPreview(cmd)})
		}
		for _, child := range visibleSubcommands(cmd) {
			visit(child)
		}
	}
	visit(target)
	result.Total = len(result.Commands)
	start := min(offset, result.Total)
	end := min(start+limit, result.Total)
	result.Commands = result.Commands[start:end]
	result.Complete = end == result.Total
	if !result.Complete {
		result.NextOffset = &end
	}
	return result, nil
}

// Compact schemas retain invocation and payload contracts, but reference the
// shared envelope rather than repeating it for every output variant.
func compactCommandSchema(schema schemaCommand) (map[string]any, error) {
	children := make([]string, len(schema.Subcommands))
	for i, child := range schema.Subcommands {
		children[i] = child.Path
	}
	schema.Subcommands = nil
	schema.Description = ""
	schema.Examples = nil
	contract := *schema.OutputContract
	contract.Variants = slices.Clone(contract.Variants)
	for i := range contract.Variants {
		contract.Variants[i].Envelope = nil
	}
	schema.OutputContract = nil
	encoded, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		return nil, err
	}
	result["envelope_ref"] = envelopeSchemaVersion
	result["output_contract"] = map[string]any{"output_mode": contract.OutputMode, "variants": contract.Variants, "stream_events": contract.StreamEvents}
	if len(children) > 0 {
		result["subcommands"] = children
	}
	encoded, err = json.Marshal(result)
	if err != nil {
		return nil, err
	}
	result["schema_id"] = fmt.Sprintf("sha256:%x", sha256.Sum256(encoded))
	return result, nil
}
