package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	applySchemaAutonomyDefaults()
}

func applySchemaAutonomyDefaults() {
	overrides := outputContractOverrides()
	walkCommandTree(rootCmd, func(cmd *cobra.Command) {
		if cmd == nil || cmd.Hidden {
			return
		}

		path := schemaPath(cmd)
		ann := getSchemaAnnotation(cmd)
		changed := false

		if len(ann.Capabilities) == 0 {
			if caps := defaultCapabilitiesForCommandPath(path); len(caps) > 0 {
				ann.Capabilities = caps
				changed = true
			}
		}

		if ann.GuideTopic == "" {
			if topic := defaultGuideTopicForCommandPath(path); topic != "" {
				ann.GuideTopic = topic
				changed = true
			}
		}

		if ann.OutputContract == nil {
			if contract, ok := overrides[path]; ok {
				contractCopy := contract
				ann.OutputContract = &contractCopy
				changed = true
			}
		}

		if changed {
			setSchemaAnnotation(cmd, ann)
		}

		wireGenericPlanSupport(cmd)
	})
}

func walkCommandTree(root *cobra.Command, fn func(*cobra.Command)) {
	if root == nil {
		return
	}
	fn(root)
	for _, child := range root.Commands() {
		walkCommandTree(child, fn)
	}
}

func defaultCapabilitiesForCommandPath(path string) []string {
	parts := strings.Fields(path)
	if len(parts) < 2 {
		return []string{"local"}
	}
	top := parts[1]

	if slices.Contains([]string{"guide", "schema", "version", "update", "help", "capability", "auth"}, top) {
		return []string{"local"}
	}
	if top == "overview" {
		return []string{"auth", "rest", "ws"}
	}
	if top == "esphome" {
		return []string{"auth", "esphome"}
	}
	if slices.Contains([]string{"area", "backup", "category", "dashboard", "device", "diagnostics", "energy", "entity", "floor", "helper", "label", "network", "person", "repairs", "search", "thread", "zone"}, top) {
		return []string{"auth", "ws"}
	}
	if slices.Contains([]string{"action", "automation", "blueprint", "calendar", "event", "integration", "notification", "scene", "script", "system", "template", "todo"}, top) {
		return []string{"auth", "rest"}
	}

	return []string{"auth"}
}

func defaultGuideTopicForCommandPath(path string) string {
	parts := strings.Fields(path)
	if len(parts) < 2 {
		return "index"
	}
	top := parts[1]

	switch {
	case top == "guide":
		return "index"
	case top == "auth":
		return "auth"
	case top == "schema" || top == "version":
		return "input-output"
	case top == "overview" || top == "search" || top == "capability":
		return "discovery"
	case slices.Contains([]string{"area", "category", "device", "entity", "floor", "label", "person", "thread", "zone"}, top):
		return "registry"
	case slices.Contains([]string{"action", "automation", "blueprint", "event", "integration", "notification", "scene", "script", "template"}, top):
		return "automation"
	case top == "dashboard":
		return "dashboard"
	case top == "helper":
		return "helpers"
	case top == "calendar" || top == "todo":
		return "calendar-todo"
	case top == "esphome":
		return "esphome"
	case slices.Contains([]string{"backup", "diagnostics", "energy", "network", "repairs", "system", "update"}, top):
		return "operations"
	default:
		return "index"
	}
}

func wireGenericPlanSupport(cmd *cobra.Command) {
	if cmd == nil || cmd.Hidden {
		return
	}
	if cmd.Run == nil && cmd.RunE == nil {
		return
	}
	if cmd.Flags().Lookup("plan") != nil || cmd.Flags().Lookup("dry-run") != nil {
		return
	}

	sideEffect := getSchemaAnnotation(cmd).SideEffect
	if sideEffect == "" {
		sideEffect = inferSideEffect(cmd)
	}
	if sideEffect != "write" && sideEffect != "destructive" {
		return
	}

	var planFlag bool
	var dryRunFlag bool
	cmd.Flags().BoolVar(&planFlag, "plan", false, "Show execution plan without applying changes")
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Alias for --plan")
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		InputSources:   []string{"flags"},
		OutputVariants: []string{"plan", "full"},
		FlagConstraints: []SchemaFlagConstraint{{
			Type:        "equivalent",
			Flags:       []string{"plan", "dry-run"},
			Description: "--dry-run is an alias of --plan.",
		}},
	})

	originalRunE := cmd.RunE
	originalRun := cmd.Run
	resourceType := defaultResourceTypeForCommand(cmd)
	commandPath := schemaPath(cmd)

	cmd.Run = nil
	cmd.RunE = func(c *cobra.Command, args []string) error {
		plan, _ := c.Flags().GetBool("plan")
		dryRun, _ := c.Flags().GetBool("dry-run")
		if plan || dryRun {
			printMutationPlan(MutationPlan{
				WouldChange: true,
				Target: map[string]any{
					"command": commandPath,
					"args":    append([]string(nil), args...),
				},
				Inputs:               changedFlagValues(c),
				Steps:                genericPlanSteps(commandPath),
				Risks:                genericPlanRisks(sideEffect),
				RequiresConfirmation: genericPlanRequiresConfirmation(c, sideEffect),
				VerificationCommands: genericVerificationCommands(commandPath, args),
			}, getTextMode(), resourceType)
			return nil
		}

		if originalRunE != nil {
			return originalRunE(c, args)
		}
		if originalRun != nil {
			originalRun(c, args)
		}
		return nil
	}
}

func changedFlagValues(cmd *cobra.Command) map[string]any {
	values := map[string]any{}
	if cmd == nil {
		return values
	}
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case "plan", "dry-run", "json", "text", "verbose", "skip-update-check", "config":
			return
		}
		values[flag.Name] = flag.Value.String()
	})
	if len(values) == 0 {
		return nil
	}
	return values
}

func genericPlanSteps(commandPath string) []string {
	return []string{
		"Validate command arguments and flags.",
		fmt.Sprintf("Execute `%s` against Home Assistant.", commandPath),
		"Return a mutation response envelope.",
	}
}

func genericPlanRisks(sideEffect string) []string {
	if sideEffect != "destructive" {
		return nil
	}
	return []string{"This operation may permanently modify or remove resources."}
}

func genericPlanRequiresConfirmation(cmd *cobra.Command, sideEffect string) bool {
	if sideEffect != "destructive" {
		return false
	}
	if cmd == nil {
		return true
	}
	if cmd.Flags().Lookup("force") == nil {
		return true
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return true
	}
	return !force
}

func genericVerificationCommands(commandPath string, args []string) []string {
	parts := strings.Fields(commandPath)
	if len(parts) < 2 {
		return nil
	}
	top := parts[1]

	switch top {
	case "area", "backup", "category", "dashboard", "device", "entity", "floor", "helper", "label", "person", "repairs", "thread", "zone", "automation", "blueprint", "integration", "scene", "script":
		return []string{fmt.Sprintf("hab %s list --json", top)}
	case "calendar":
		if len(args) > 0 {
			return []string{fmt.Sprintf("hab calendar list %s --json", args[0])}
		}
		return []string{"hab calendar list <entity_id> --json"}
	case "todo":
		if len(args) > 0 {
			return []string{fmt.Sprintf("hab todo items %s --json", args[0])}
		}
		return []string{"hab todo items <entity_id> --json"}
	case "energy":
		return []string{"hab energy prefs get --json"}
	case "network":
		return []string{"hab network get --json"}
	default:
		return nil
	}
}
