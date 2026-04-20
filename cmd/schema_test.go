package cmd

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestParseUseArgs(t *testing.T) {
	args := parseUseArgs("create <dashboard_url_path> [view_index]")
	if len(args) != 2 {
		t.Fatalf("parseUseArgs length = %d, want 2", len(args))
	}
	if !args[0].Required || args[0].Name != "dashboard_url_path" {
		t.Fatalf("first arg = %#v, want required dashboard_url_path", args[0])
	}
	if args[1].Required || args[1].Name != "view_index" {
		t.Fatalf("second arg = %#v, want optional view_index", args[1])
	}
}

func TestSchemaTreeContainsSchemaCommand(t *testing.T) {
	tree := buildSchemaTree(rootCmd)
	node := findSchemaNode(tree, "hab schema")
	if node == nil {
		t.Fatal("schema tree missing hab schema command")
	}
	if node.SideEffect != "read" {
		t.Fatalf("schema side_effect = %q, want read", node.SideEffect)
	}
}

func TestFactoryAnnotationCoverageInSchema(t *testing.T) {
	areaCreateCmd, _, err := rootCmd.Find([]string{"area", "create"})
	if err != nil {
		t.Fatalf("find area create: %v", err)
	}

	s := buildSchemaTree(areaCreateCmd)
	if s.SideEffect != "write" {
		t.Fatalf("side_effect = %q, want write", s.SideEffect)
	}
	if !slices.Contains(s.Capabilities, "ws") {
		t.Fatalf("capabilities = %#v, want ws", s.Capabilities)
	}
	if s.ResourceType != "area" {
		t.Fatalf("resource_type = %q, want area", s.ResourceType)
	}
}

func TestInputFlagConstraintIncludedInSchema(t *testing.T) {
	automationCreateCmd, _, err := rootCmd.Find([]string{"automation", "create"})
	if err != nil {
		t.Fatalf("find automation create: %v", err)
	}

	s := buildSchemaTree(automationCreateCmd)
	found := false
	for _, c := range s.FlagConstraints {
		if c.Type == "one_of" && slices.Contains(c.Flags, "data") && slices.Contains(c.Flags, "file") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing one_of data/file constraint: %#v", s.FlagConstraints)
	}
}

func TestSchemaTreeCoversVisibleCommands(t *testing.T) {
	tree := buildSchemaTree(rootCmd)
	got := countSchemaNodes(tree)
	want := countVisibleCommands(rootCmd)
	if got != want {
		t.Fatalf("schema node count = %d, want %d", got, want)
	}
}

func TestSchemaIncludesOutputContractVariants(t *testing.T) {
	scriptListCmd, _, err := rootCmd.Find([]string{"script", "list"})
	if err != nil {
		t.Fatalf("find script list: %v", err)
	}
	s := buildSchemaTree(scriptListCmd)
	if s.OutputContract == nil {
		t.Fatal("expected output contract for script list")
	}
	variantNames := make([]string, 0, len(s.OutputContract.Variants))
	for _, variant := range s.OutputContract.Variants {
		variantNames = append(variantNames, variant.Name)
	}
	for _, expected := range []string{"full", "brief", "count"} {
		if !slices.Contains(variantNames, expected) {
			t.Fatalf("variants = %#v, want %q", variantNames, expected)
		}
	}
}

func TestSchemaIncludesPlanVariantForMutation(t *testing.T) {
	areaCreateCmd, _, err := rootCmd.Find([]string{"area", "create"})
	if err != nil {
		t.Fatalf("find area create: %v", err)
	}
	s := buildSchemaTree(areaCreateCmd)
	if s.OutputContract == nil {
		t.Fatal("expected output contract for area create")
	}
	variantNames := make([]string, 0, len(s.OutputContract.Variants))
	for _, variant := range s.OutputContract.Variants {
		variantNames = append(variantNames, variant.Name)
	}
	if !slices.Contains(variantNames, "plan") {
		t.Fatalf("variants = %#v, want plan", variantNames)
	}
}

func TestSchemaIncludesStreamContractForESPHome(t *testing.T) {
	logsCmd, _, err := rootCmd.Find([]string{"esphome", "logs"})
	if err != nil {
		t.Fatalf("find esphome logs: %v", err)
	}
	s := buildSchemaTree(logsCmd)
	if s.OutputContract == nil {
		t.Fatal("expected output contract for esphome logs")
	}
	if s.OutputContract.OutputMode != "ndjson_stream" {
		t.Fatalf("output_mode = %q, want ndjson_stream", s.OutputContract.OutputMode)
	}
	if len(s.OutputContract.StreamEvents) == 0 {
		t.Fatal("expected stream event contracts")
	}
}

func TestSchemaOutputContractsPresentForVisibleCommands(t *testing.T) {
	tree := buildSchemaTree(rootCmd)
	assertOutputContracts(t, tree)
}

func TestSchemaResourceTypeMatchesCommandFamilyDefaults(t *testing.T) {
	cases := []struct {
		path         []string
		resourceType string
	}{
		{path: []string{"backup", "delete"}, resourceType: "backup"},
		{path: []string{"system", "restart"}, resourceType: "system"},
		{path: []string{"thread", "delete"}, resourceType: "thread_dataset"},
		{path: []string{"notification", "list"}, resourceType: "notification"},
	}

	for _, tc := range cases {
		t.Run(joinCommandPath(tc.path), func(t *testing.T) {
			cmd, _, err := rootCmd.Find(tc.path)
			if err != nil {
				t.Fatalf("find %v: %v", tc.path, err)
			}
			s := buildSchemaTree(cmd)
			if s.ResourceType != tc.resourceType {
				t.Fatalf("resource_type = %q, want %q", s.ResourceType, tc.resourceType)
			}
		})
	}
}

func TestDestructiveCommandsExposeForceAndPlanFlags(t *testing.T) {
	paths := [][]string{
		{"backup", "delete"},
		{"system", "restart"},
		{"thread", "delete"},
		{"blueprint", "delete"},
		{"dashboard", "delete"},
		{"zone", "delete"},
		{"device", "delete"},
		{"person", "delete"},
		{"category", "delete"},
		{"esphome", "serial", "erase-flash"},
	}

	for _, path := range paths {
		t.Run(joinCommandPath(path), func(t *testing.T) {
			cmd, _, err := rootCmd.Find(path)
			if err != nil {
				t.Fatalf("find %v: %v", path, err)
			}
			s := buildSchemaTree(cmd)
			if s.SideEffect != "destructive" {
				t.Fatalf("side_effect = %q, want destructive", s.SideEffect)
			}
			for _, required := range []string{"force", "plan", "dry-run"} {
				if !schemaHasFlag(s.Flags, required) {
					t.Fatalf("missing %q flag; flags = %#v", required, s.Flags)
				}
			}
			if s.OutputContract == nil {
				t.Fatal("missing output contract")
			}
			if !schemaHasVariant(s.OutputContract.Variants, "plan") {
				t.Fatalf("missing plan output variant: %#v", s.OutputContract.Variants)
			}
		})
	}
}

func TestVisibleRunnableCommandsExposeExplicitCapabilities(t *testing.T) {
	var missing []string
	walkCommandTree(rootCmd, func(command *cobra.Command) {
		if command == nil || command.Hidden {
			return
		}
		if command.Run == nil && command.RunE == nil {
			return
		}
		if len(getSchemaAnnotation(command).Capabilities) == 0 {
			missing = append(missing, schemaPath(command))
		}
	})

	if len(missing) > 0 {
		t.Fatalf("commands missing explicit capabilities: %v", missing)
	}
}

func TestVisibleRunnableCommandsExposeGuideTopic(t *testing.T) {
	var missing []string
	walkCommandTree(rootCmd, func(command *cobra.Command) {
		if command == nil || command.Hidden {
			return
		}
		if command.Run == nil && command.RunE == nil {
			return
		}
		path := schemaPath(command)
		if path == "hab help" {
			return
		}
		if getSchemaAnnotation(command).GuideTopic == "" {
			missing = append(missing, path)
		}
	})

	if len(missing) > 0 {
		t.Fatalf("commands missing guide_topic: %v", missing)
	}
}

func TestMutatingCommandsExposePlanFlags(t *testing.T) {
	tree := buildSchemaTree(rootCmd)
	var missing []string
	collectSchemaNodes(tree, func(node schemaCommand) {
		if node.SideEffect != "write" && node.SideEffect != "destructive" {
			return
		}
		if !schemaHasFlag(node.Flags, "plan") || !schemaHasFlag(node.Flags, "dry-run") {
			missing = append(missing, node.Path)
			return
		}
		if node.OutputContract == nil || !schemaHasVariant(node.OutputContract.Variants, "plan") {
			missing = append(missing, node.Path+"(missing plan variant)")
		}
	})

	if len(missing) > 0 {
		t.Fatalf("mutating commands missing plan support: %v", missing)
	}
}

func TestGuideTopicFamilyMappings(t *testing.T) {
	cases := map[string]string{
		"hab area list":        "registry",
		"hab action docs":      "automation",
		"hab dashboard list":   "dashboard",
		"hab helper list":      "helpers",
		"hab calendar list":    "calendar-todo",
		"hab esphome update":   "esphome",
		"hab backup list":      "operations",
		"hab overview":         "discovery",
		"hab auth status":      "auth",
		"hab schema":           "input-output",
		"hab guide":            "index",
		"hab capability probe": "discovery",
	}

	tree := buildSchemaTree(rootCmd)
	for path, expected := range cases {
		t.Run(path, func(t *testing.T) {
			node := findSchemaNode(tree, path)
			if node == nil {
				t.Fatalf("missing schema node for %s", path)
			}
			if node.GuideTopic != expected {
				t.Fatalf("guide_topic for %s = %q, want %q", path, node.GuideTopic, expected)
			}
		})
	}
}

func TestHighValueCommandsUseSpecificOutputContracts(t *testing.T) {
	paths := []string{
		"hab overview",
		"hab auth status",
		"hab capability probe",
		"hab area list",
		"hab area get",
		"hab entity search",
		"hab action docs",
		"hab dashboard list",
		"hab dashboard get",
		"hab calendar list",
		"hab todo items",
		"hab esphome update",
		"hab esphome validate",
		"hab esphome logs",
	}

	tree := buildSchemaTree(rootCmd)
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			node := findSchemaNode(tree, path)
			if node == nil {
				t.Fatalf("missing schema node for %s", path)
			}
			if node.OutputContract == nil {
				t.Fatalf("missing output contract for %s", path)
			}
			for _, variant := range node.OutputContract.Variants {
				if variant.Name != "full" {
					continue
				}
				if variant.Data == nil {
					t.Fatalf("full variant missing data contract for %s", path)
				}
				desc := strings.ToLower(variant.Data.Description)
				if strings.Contains(desc, "fields vary by") {
					t.Fatalf("full variant for %s remains generic: %#v", path, variant.Data)
				}
			}
		})
	}
}

func collectSchemaNodes(node schemaCommand, fn func(schemaCommand)) {
	fn(node)
	for _, child := range node.Subcommands {
		collectSchemaNodes(child, fn)
	}
}

func TestSchemaCapabilityAlignmentSmokeCheck(t *testing.T) {
	cases := []struct {
		path []string
		want []string
	}{
		{path: []string{"zone", "create"}, want: []string{"auth", "ws"}},
		{path: []string{"auth", "status"}, want: []string{"local"}},
		{path: []string{"esphome", "update"}, want: []string{"auth", "esphome"}},
		{path: []string{"overview"}, want: []string{"auth", "rest", "ws"}},
	}

	for _, tc := range cases {
		t.Run(joinCommandPath(tc.path), func(t *testing.T) {
			command, _, err := rootCmd.Find(tc.path)
			if err != nil {
				t.Fatalf("find command %v: %v", tc.path, err)
			}
			caps := getSchemaAnnotation(command).Capabilities
			if len(caps) == 0 {
				t.Fatalf("no explicit capabilities for %s", schemaPath(command))
			}
			for _, capability := range tc.want {
				if !slices.Contains(caps, capability) {
					t.Fatalf("capabilities for %s = %v, missing %q", schemaPath(command), caps, capability)
				}
			}
		})
	}
}

func assertOutputContracts(t *testing.T, node schemaCommand) {
	t.Helper()
	if node.OutputContract == nil {
		t.Fatalf("command %q missing output contract", node.Path)
	}
	if node.OutputContract.SuccessEnvelope.Type == "" {
		t.Fatalf("command %q missing success envelope contract", node.Path)
	}
	if node.OutputContract.ErrorEnvelope.Type == "" {
		t.Fatalf("command %q missing error envelope contract", node.Path)
	}
	if node.OutputContract.PartialEnvelope == nil {
		t.Fatalf("command %q missing partial envelope contract", node.Path)
	}
	for _, child := range node.Subcommands {
		assertOutputContracts(t, child)
	}
}

func findSchemaNode(node schemaCommand, path string) *schemaCommand {
	if node.Path == path {
		result := node
		return &result
	}
	for _, child := range node.Subcommands {
		if found := findSchemaNode(child, path); found != nil {
			return found
		}
	}
	return nil
}

func countSchemaNodes(node schemaCommand) int {
	total := 1
	for _, child := range node.Subcommands {
		total += countSchemaNodes(child)
	}
	return total
}

func countVisibleCommands(cmd *cobra.Command) int {
	total := 1
	for _, child := range cmd.Commands() {
		if child.Hidden {
			continue
		}
		total += countVisibleCommands(child)
	}
	return total
}

func joinCommandPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := ""
	for i, part := range parts {
		if i > 0 {
			out += " "
		}
		out += part
	}
	return out
}

func schemaHasFlag(flags []schemaFlag, name string) bool {
	for _, flag := range flags {
		if flag.Name == name {
			return true
		}
	}
	return false
}

func schemaHasVariant(variants []SchemaOutputVariant, name string) bool {
	for _, variant := range variants {
		if variant.Name == name {
			return true
		}
	}
	return false
}

func TestSchemaOutputContractFallbackList(t *testing.T) {
	allowedGeneric := map[string]struct{}{
		"hab":            {},
		"hab help":       {},
		"hab guide":      {},
		"hab schema":     {},
		"hab version":    {},
		"hab update":     {},
		"hab capability": {},
	}

	tree := buildSchemaTree(rootCmd)
	var generic []string
	collectSchemaNodes(tree, func(node schemaCommand) {
		if node.OutputContract == nil {
			return
		}
		if _, ok := allowedGeneric[node.Path]; ok {
			return
		}
		for _, variant := range node.OutputContract.Variants {
			if variant.Name != "full" || variant.Data == nil {
				continue
			}
			desc := strings.ToLower(variant.Data.Description)
			if strings.Contains(desc, "fields vary by") {
				generic = append(generic, fmt.Sprintf("%s (%s)", node.Path, desc))
			}
		}
	})

	if len(generic) > 30 {
		t.Fatalf("too many generic full contracts remain (%d): %v", len(generic), generic)
	}
}
