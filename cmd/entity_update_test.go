package cmd

import (
	"testing"
)

func TestEntityUpdateSchema(t *testing.T) {
	cmd, err := resolveSchemaTarget([]string{"entity", "update"})
	if err != nil {
		t.Fatal(err)
	}
	schema := buildSchemaTree(cmd)
	if len(schema.Args) != 1 || schema.Args[0].Name != "entity_id" || !schema.Args[0].Required {
		t.Fatalf("incorrect positional arguments: %+v", schema.Args)
	}
	for _, flag := range schema.Flags {
		if flag.Required {
			t.Fatal("individual update flags must be optional")
		}
	}
	foundConstraint := false
	for _, constraint := range schema.FlagConstraints {
		if constraint.Type == "at_least_one" && len(constraint.Flags) == 3 {
			foundConstraint = true
		}
	}
	if !foundConstraint {
		t.Fatal("missing update field requirement")
	}
	for _, variant := range schema.OutputContract.Variants {
		fields := map[string]bool{}
		for _, field := range variant.Data.Fields {
			fields[field.Name] = true
		}
		if !fields["old_entity_id"] || !fields["new_entity_id"] || !fields["saved"] || !fields["verified"] || !fields["changes"] || fields["would_change"] {
			t.Fatalf("incorrect %s result contract: %+v", variant.Name, variant.Data.Fields)
		}
	}
}

func TestEntityUpdateRequiresAField(t *testing.T) {
	for _, flags := range [][]string{nil, {"--name", "Kitchen"}, {"--icon", ""}, {"--entity-id", "switch.new", "--dry-run"}, {"--name", "", "--icon", "mdi:power"}} {
		cmd := newEntityUpdateCommand()
		if err := cmd.ParseFlags(flags); err != nil {
			t.Fatal(err)
		}
		err := cmd.ValidateFlagGroups()
		if (err != nil) != (flags == nil) {
			t.Fatalf("flags=%v error=%v", flags, err)
		}
	}
}
