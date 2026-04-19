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
