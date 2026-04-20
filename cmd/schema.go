package cmd

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type schemaFlag struct {
	Name       string `json:"name"`
	Shorthand  string `json:"shorthand,omitempty"`
	Type       string `json:"type"`
	Default    string `json:"default,omitempty"`
	Usage      string `json:"usage,omitempty"`
	Required   bool   `json:"required,omitempty"`
	Deprecated string `json:"deprecated,omitempty"`
	Hidden     bool   `json:"hidden,omitempty"`
	Inherited  bool   `json:"inherited,omitempty"`
}

type schemaCommand struct {
	Path            string                 `json:"path"`
	Use             string                 `json:"use"`
	Aliases         []string               `json:"aliases,omitempty"`
	Summary         string                 `json:"summary,omitempty"`
	Description     string                 `json:"description,omitempty"`
	Examples        []string               `json:"examples,omitempty"`
	GroupID         string                 `json:"group_id,omitempty"`
	Hidden          bool                   `json:"hidden,omitempty"`
	Deprecated      string                 `json:"deprecated,omitempty"`
	SideEffect      string                 `json:"side_effect,omitempty"`
	OutputMode      string                 `json:"output_mode,omitempty"`
	OutputVariants  []string               `json:"output_variants,omitempty"`
	Capabilities    []string               `json:"capabilities,omitempty"`
	InputSources    []string               `json:"input_sources,omitempty"`
	ResourceType    string                 `json:"resource_type,omitempty"`
	Args            []SchemaPositionalArg  `json:"args,omitempty"`
	FlagConstraints []SchemaFlagConstraint `json:"flag_constraints,omitempty"`
	GuideTopic      string                 `json:"guide_topic,omitempty"`
	OutputContract  *SchemaOutputContract  `json:"output_contract,omitempty"`
	Flags           []schemaFlag           `json:"flags,omitempty"`
	InheritedFlags  []schemaFlag           `json:"inherited_flags,omitempty"`
	Subcommands     []schemaCommand        `json:"subcommands,omitempty"`
}

var schemaCmd = &cobra.Command{
	Use:     "schema [command path]",
	Short:   "Show machine-readable command schema",
	Long:    "Emit machine-readable command and flag schemas for LLM/tooling integration.",
	Example: "hab schema --json\nhab schema automation create --json\nhab schema esphome logs --json",
	Args:    cobra.ArbitraryArgs,
	GroupID: "start",
	RunE:    runSchema,
}

func init() {
	rootCmd.AddCommand(schemaCmd)
	mergeSchemaAnnotation(schemaCmd, SchemaAnnotation{
		SideEffect:   "read",
		OutputMode:   "json_envelope",
		Capabilities: []string{"local"},
		ResourceType: "command_schema",
	})
}

func runSchema(cmd *cobra.Command, args []string) error {
	target, err := resolveSchemaTarget(args)
	if err != nil {
		return err
	}

	schema := buildSchemaTree(target)
	output.PrintOutputWithContext(schema, getTextMode(), "", output.EnvelopeContext{
		Operation:    "schema",
		ResourceType: "command",
	})
	return nil
}

func resolveSchemaTarget(args []string) (*cobra.Command, error) {
	if len(args) == 0 {
		return rootCmd, nil
	}

	trimmed := args
	if args[0] == "hab" || args[0] == rootCmd.Name() {
		trimmed = args[1:]
	}
	if len(trimmed) == 0 {
		return rootCmd, nil
	}

	target, _, err := rootCmd.Find(trimmed)
	if err != nil {
		return nil, fmt.Errorf("unknown command path %q", strings.Join(args, " "))
	}

	return target, nil
}

func buildSchemaTree(cmd *cobra.Command) schemaCommand {
	ann := getSchemaAnnotation(cmd)
	path := schemaPath(cmd)

	sideEffect := ann.SideEffect
	if sideEffect == "" {
		sideEffect = inferSideEffect(cmd)
	}

	outputMode := ann.OutputMode
	if outputMode == "" {
		outputMode = inferOutputMode(path, cmd)
	}

	capabilities := slices.Clone(ann.Capabilities)
	if len(capabilities) == 0 {
		capabilities = inferCapabilities(path)
	}

	inputSources := slices.Clone(ann.InputSources)
	if len(parseUseArgs(cmd.Use)) > 0 && !slices.Contains(inputSources, "args") {
		inputSources = append(inputSources, "args")
	}
	if cmd.Flags().HasFlags() && !slices.Contains(inputSources, "flags") {
		inputSources = append(inputSources, "flags")
	}

	args := ann.Args
	if len(args) == 0 {
		args = parseUseArgs(cmd.Use)
	}

	outputContract := ann.OutputContract
	if outputContract == nil {
		contract := inferOutputContract(path, cmd, ann, sideEffect, outputMode)
		outputContract = &contract
	}

	sc := schemaCommand{
		Path:            path,
		Use:             cmd.Use,
		Aliases:         slices.Clone(cmd.Aliases),
		Summary:         cmd.Short,
		Description:     cmd.Long,
		Examples:        parseExamples(cmd.Example),
		GroupID:         cmd.GroupID,
		Hidden:          cmd.Hidden,
		Deprecated:      cmd.Deprecated,
		SideEffect:      sideEffect,
		OutputMode:      outputMode,
		OutputVariants:  slices.Clone(ann.OutputVariants),
		Capabilities:    capabilities,
		InputSources:    inputSources,
		ResourceType:    ann.ResourceType,
		Args:            args,
		FlagConstraints: slices.Clone(ann.FlagConstraints),
		GuideTopic:      ann.GuideTopic,
		OutputContract:  outputContract,
		Flags:           extractFlags(cmd.LocalFlags(), false),
		InheritedFlags:  extractFlags(cmd.InheritedFlags(), true),
	}

	children := visibleSubcommands(cmd)
	sc.Subcommands = make([]schemaCommand, len(children))
	for i, child := range children {
		sc.Subcommands[i] = buildSchemaTree(child)
	}

	return sc
}

func parseExamples(example string) []string {
	if example == "" {
		return nil
	}
	lines := strings.Split(example, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func extractFlags(flags *pflag.FlagSet, inherited bool) []schemaFlag {
	result := make([]schemaFlag, 0)
	flags.VisitAll(func(flag *pflag.Flag) {
		result = append(result, schemaFlag{
			Name:       flag.Name,
			Shorthand:  flag.Shorthand,
			Type:       flag.Value.Type(),
			Default:    flag.DefValue,
			Usage:      flag.Usage,
			Required:   isRequiredFlag(flag),
			Deprecated: flag.Deprecated,
			Hidden:     flag.Hidden,
			Inherited:  inherited,
		})
	})
	slices.SortFunc(result, func(a, b schemaFlag) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return result
}

func isRequiredFlag(flag *pflag.Flag) bool {
	if flag == nil || flag.Annotations == nil {
		return false
	}
	_, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]
	return ok
}

func visibleSubcommands(cmd *cobra.Command) []*cobra.Command {
	commands := cmd.Commands()
	visible := make([]*cobra.Command, 0, len(commands))
	for _, child := range commands {
		if child.Hidden {
			continue
		}
		visible = append(visible, child)
	}
	slices.SortFunc(visible, func(a, b *cobra.Command) int {
		return cmp.Compare(a.Name(), b.Name())
	})
	return visible
}

func parseUseArgs(use string) []SchemaPositionalArg {
	parts := strings.Fields(use)
	if len(parts) <= 1 {
		return nil
	}

	args := make([]SchemaPositionalArg, 0, len(parts)-1)
	for _, token := range parts[1:] {
		if strings.HasPrefix(token, "-") {
			continue
		}

		required := strings.HasPrefix(token, "<") && strings.HasSuffix(token, ">")
		optional := strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]")
		if !required && !optional {
			continue
		}

		name := strings.Trim(token, "<>")
		name = strings.Trim(name, "[]")
		if name == "" {
			continue
		}

		arg := SchemaPositionalArg{
			Name:     name,
			Required: required,
			Raw:      token,
		}
		if strings.Contains(name, "|") {
			arg.Choices = strings.Split(name, "|")
		}

		args = append(args, arg)
	}

	return args
}

func schemaPath(cmd *cobra.Command) string {
	path := strings.TrimSpace(cmd.CommandPath())
	if path == "" {
		return "hab"
	}
	parts := strings.Fields(path)
	parts[0] = "hab"
	return strings.Join(parts, " ")
}

func inferSideEffect(cmd *cobra.Command) string {
	if cmd.RunE == nil && cmd.Run == nil {
		return "meta"
	}

	name := cmd.Name()
	switch name {
	case "delete", "remove", "dismiss", "erase", "restore", "restart":
		return "destructive"
	case "create", "update", "set", "configure", "enable", "disable", "assign", "unassign", "fire", "call", "import", "reload", "run", "trigger", "rename", "add", "patch", "write":
		return "write"
	default:
		return "read"
	}
}

func inferOutputMode(path string, cmd *cobra.Command) string {
	if strings.HasPrefix(path, "hab esphome ") {
		switch cmd.Name() {
		case "build", "upload", "run", "logs":
			return "ndjson_stream"
		}
	}
	return "json_envelope"
}

func inferCapabilities(path string) []string {
	if path == "hab" || path == "hab guide" || path == "hab schema" || path == "hab version" || path == "hab update" || strings.HasPrefix(path, "hab capability") {
		return []string{"local"}
	}
	if strings.HasPrefix(path, "hab auth") {
		return []string{"local"}
	}
	if strings.HasPrefix(path, "hab esphome") {
		return []string{"auth", "esphome"}
	}
	return []string{"auth"}
}

func inferOutputContract(path string, cmd *cobra.Command, ann SchemaAnnotation, sideEffect, outputMode string) SchemaOutputContract {
	resourceType := ann.ResourceType
	if resourceType == "" {
		resourceType = cmd.Name()
	}
	baseSuccess := baseEnvelopeContract(resourceType, outputMode, true)
	baseError := baseEnvelopeContract(resourceType, outputMode, false)
	partial := partialEnvelopeContract(resourceType, outputMode)

	variants := make([]SchemaOutputVariant, 0)
	variantNames := slices.Clone(ann.OutputVariants)
	if len(variantNames) == 0 {
		variantNames = defaultVariantNames(path, cmd, sideEffect, outputMode)
	}
	for _, variant := range variantNames {
		variants = append(variants, SchemaOutputVariant{
			Name:        variant,
			Description: variantDescription(variant, resourceType),
			OutputMode:  outputMode,
			Envelope:    &baseSuccess,
			Data:        dataContractForVariant(variant, resourceType, sideEffect, cmd.Name()),
		})
	}

	contract := SchemaOutputContract{
		OutputMode:      outputMode,
		SuccessEnvelope: baseSuccess,
		ErrorEnvelope:   baseError,
		PartialEnvelope: &partial,
		Variants:        variants,
	}
	if outputMode == "ndjson_stream" {
		contract.StreamEvents = []SchemaObjectContract{
			{
				Type:        "object",
				Description: fmt.Sprintf("stream event records for %s", resourceType),
				Fields:      []SchemaField{{Name: "event", Type: "string", Required: true, Description: "event type"}, {Name: "message", Type: "string", Description: "human-readable event detail"}, {Name: "data", Type: "object", Description: "event-specific payload", AdditionalProps: true}},
			},
		}
	}
	return contract
}

func defaultVariantNames(path string, cmd *cobra.Command, sideEffect, outputMode string) []string {
	if outputMode == "ndjson_stream" {
		return []string{"stream"}
	}
	if cmd.Flags().Lookup("plan") != nil || cmd.Flags().Lookup("dry-run") != nil {
		return []string{"full", "plan"}
	}
	switch cmd.Name() {
	case "list":
		return []string{"full", "brief", "count"}
	case "get", "show", "info", "trace", "history", "logbook", "docs", "probe", "overview", "schema":
		return []string{"full"}
	case "create", "update", "patch", "rename", "assign", "remove", "set", "run", "call", "fire", "delete", "restart", "restore":
		return []string{"full"}
	default:
		if sideEffect == "meta" {
			return []string{"full"}
		}
		return []string{"full"}
	}
}

func variantDescription(variant, resourceType string) string {
	switch variant {
	case "brief":
		return fmt.Sprintf("minimal %s list entries", resourceType)
	case "count":
		return fmt.Sprintf("count summary for %s results", resourceType)
	case "plan":
		return fmt.Sprintf("dry-run execution plan for %s mutations", resourceType)
	case "stream":
		return fmt.Sprintf("streaming event records for %s", resourceType)
	default:
		return fmt.Sprintf("default %s response payload", resourceType)
	}
}

func baseEnvelopeContract(resourceType, outputMode string, success bool) SchemaObjectContract {
	fields := []SchemaField{
		{Name: "success", Type: "boolean", Required: true, Description: "indicates whether the command succeeded"},
		{Name: "operation", Type: "string", Description: "normalized operation name"},
		{Name: "resource_type", Type: "string", Description: fmt.Sprintf("resource family for %s", resourceType)},
		{Name: "data", Type: "object", Description: fmt.Sprintf("payload for %s responses", resourceType), AdditionalProps: true},
		{Name: "message", Type: "string", Description: "human-readable status message"},
		{Name: "partial_result", Type: "boolean", Description: "whether the response omitted or degraded part of the requested data"},
		{Name: "warnings", Type: "array", ItemType: "string", Description: "non-fatal warnings about the operation or payload"},
		{Name: "fallbacks_applied", Type: "array", ItemType: "string", Description: "fallback behaviors used to satisfy the request"},
		{Name: "missing_sections", Type: "array", ItemType: "string", Description: "requested sections that could not be returned"},
		{Name: "verification_commands", Type: "array", ItemType: "string", Description: "commands recommended to verify the result"},
		{Name: "next_suggested_commands", Type: "array", ItemType: "string", Description: "commands commonly executed after this result"},
		{Name: "metadata", Type: "object", Description: "execution metadata such as timestamp, transport, auth source, and resolution decisions", AdditionalProps: true},
	}
	if !success {
		fields = append(fields, SchemaField{Name: "error", Type: "object", Required: true, Description: "structured error detail", Fields: []SchemaField{{Name: "code", Type: "string", Required: true}, {Name: "message", Type: "string", Required: true}, {Name: "details", Type: "object", Description: "error-specific machine-readable context", AdditionalProps: true}}})
	}
	return SchemaObjectContract{Type: "object", Description: fmt.Sprintf("%s %s envelope", outputMode, map[bool]string{true: "success", false: "error"}[success]), Fields: fields}
}

func partialEnvelopeContract(resourceType, outputMode string) SchemaObjectContract {
	contract := baseEnvelopeContract(resourceType, outputMode, true)
	contract.Description = fmt.Sprintf("partial %s envelope for %s", outputMode, resourceType)
	return contract
}

func dataContractForVariant(variant, resourceType, sideEffect, commandName string) *SchemaObjectContract {
	switch variant {
	case "count":
		return &SchemaObjectContract{Type: "object", Description: fmt.Sprintf("count result for %s", resourceType), Fields: []SchemaField{{Name: "count", Type: "number", Required: true, Description: "number of matching items"}}}
	case "brief":
		return &SchemaObjectContract{Type: "array", Description: fmt.Sprintf("minimal list rows for %s", resourceType), Fields: []SchemaField{{Name: "items", Type: "object", Description: "brief resource rows", Fields: []SchemaField{{Name: "id", Type: "string", Description: "resource identifier when available"}, {Name: "name", Type: "string", Description: "display name when available"}}, AdditionalProps: true}}}
	case "plan":
		return &SchemaObjectContract{Type: "object", Description: fmt.Sprintf("execution plan for %s", resourceType), Fields: []SchemaField{{Name: "mode", Type: "string", Required: true, Enum: []string{"plan"}}, {Name: "would_change", Type: "boolean", Required: true}, {Name: "target", Type: "object", Required: true, AdditionalProps: true}, {Name: "inputs", Type: "object", AdditionalProps: true}, {Name: "derived_ids", Type: "object", AdditionalProps: true}, {Name: "steps", Type: "array", ItemType: "string", Required: true}, {Name: "risks", Type: "array", ItemType: "string"}, {Name: "requires_confirmation", Type: "boolean", Required: true}, {Name: "verification_commands", Type: "array", ItemType: "string"}}}
	case "stream":
		return &SchemaObjectContract{Type: "array", Description: fmt.Sprintf("stream records for %s", resourceType)}
	default:
		shapeType := "object"
		if commandName == "list" {
			shapeType = "array"
		}
		if sideEffect == "read" {
			if shapeType == "array" {
				return &SchemaObjectContract{Type: "array", Description: fmt.Sprintf("resource-specific list payload for %s", resourceType)}
			}
			return &SchemaObjectContract{Type: "object", Description: fmt.Sprintf("resource-specific payload for %s", resourceType), Fields: []SchemaField{{Name: "payload", Type: "object", Description: fmt.Sprintf("fields vary by %s command", resourceType), AdditionalProps: true}}}
		}
		return &SchemaObjectContract{Type: "object", Description: fmt.Sprintf("mutation result payload for %s", resourceType), Fields: []SchemaField{{Name: "payload", Type: "object", Description: fmt.Sprintf("fields vary by %s command", resourceType), AdditionalProps: true}}}
	}
}
