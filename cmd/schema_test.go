package cmd

import (
	"slices"
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
