package cmd

import (
	"encoding/json"
	"slices"

	"github.com/spf13/cobra"
)

const schemaAnnotationKey = "hab/schema"

// SchemaPositionalArg defines a positional argument in machine-readable form.
type SchemaPositionalArg struct {
	Name     string   `json:"name"`
	Required bool     `json:"required"`
	Raw      string   `json:"raw,omitempty"`
	Choices  []string `json:"choices,omitempty"`
}

// SchemaFlagConstraint describes relationships between flags.
type SchemaFlagConstraint struct {
	Type        string   `json:"type"`
	Flags       []string `json:"flags"`
	Description string   `json:"description,omitempty"`
}

// SchemaAnnotation stores supplemental command metadata not inferable from Cobra.
type SchemaAnnotation struct {
	SideEffect      string                 `json:"side_effect,omitempty"`
	OutputMode      string                 `json:"output_mode,omitempty"`
	OutputVariants  []string               `json:"output_variants,omitempty"`
	Capabilities    []string               `json:"capabilities,omitempty"`
	InputSources    []string               `json:"input_sources,omitempty"`
	ResourceType    string                 `json:"resource_type,omitempty"`
	Args            []SchemaPositionalArg  `json:"args,omitempty"`
	FlagConstraints []SchemaFlagConstraint `json:"flag_constraints,omitempty"`
	GuideTopic      string                 `json:"guide_topic,omitempty"`
}

func getSchemaAnnotation(cmd *cobra.Command) SchemaAnnotation {
	if cmd == nil || cmd.Annotations == nil {
		return SchemaAnnotation{}
	}
	raw, ok := cmd.Annotations[schemaAnnotationKey]
	if !ok || raw == "" {
		return SchemaAnnotation{}
	}

	var ann SchemaAnnotation
	if err := json.Unmarshal([]byte(raw), &ann); err != nil {
		return SchemaAnnotation{}
	}
	return ann
}

func setSchemaAnnotation(cmd *cobra.Command, ann SchemaAnnotation) {
	if cmd == nil {
		return
	}
	b, err := json.Marshal(ann)
	if err != nil {
		return
	}
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations[schemaAnnotationKey] = string(b)
}

func mergeSchemaAnnotation(cmd *cobra.Command, incoming SchemaAnnotation) {
	current := getSchemaAnnotation(cmd)

	if incoming.SideEffect != "" {
		current.SideEffect = incoming.SideEffect
	}
	if incoming.OutputMode != "" {
		current.OutputMode = incoming.OutputMode
	}
	if incoming.ResourceType != "" {
		current.ResourceType = incoming.ResourceType
	}
	if incoming.GuideTopic != "" {
		current.GuideTopic = incoming.GuideTopic
	}
	if len(incoming.Args) > 0 {
		current.Args = incoming.Args
	}
	if len(incoming.FlagConstraints) > 0 {
		current.FlagConstraints = append(current.FlagConstraints, incoming.FlagConstraints...)
	}
	for _, variant := range incoming.OutputVariants {
		if !slices.Contains(current.OutputVariants, variant) {
			current.OutputVariants = append(current.OutputVariants, variant)
		}
	}
	for _, capability := range incoming.Capabilities {
		if !slices.Contains(current.Capabilities, capability) {
			current.Capabilities = append(current.Capabilities, capability)
		}
	}
	for _, source := range incoming.InputSources {
		if !slices.Contains(current.InputSources, source) {
			current.InputSources = append(current.InputSources, source)
		}
	}

	setSchemaAnnotation(cmd, current)
}
