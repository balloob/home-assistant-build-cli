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
	if path == "hab" || path == "hab guide" || path == "hab schema" || path == "hab version" || path == "hab update" {
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
