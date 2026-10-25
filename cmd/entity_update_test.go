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
		if flag.Name == "entity-id" && !flag.Required {
			t.Fatal("new entity ID flag must be required")
		}
	}
	for _, variant := range schema.OutputContract.Variants {
		fields := map[string]bool{}
		for _, field := range variant.Data.Fields {
			fields[field.Name] = true
		}
		if !fields["old_entity_id"] || !fields["new_entity_id"] || !fields["saved"] || !fields["verified"] || fields["would_change"] {
			t.Fatalf("incorrect %s result contract: %+v", variant.Name, variant.Data.Fields)
		}
	}
}

func TestEntityUpdateRequiresNewID(t *testing.T) {
	cmd := newEntityUpdateCommand()
	if err := cmd.ValidateRequiredFlags(); err == nil {
		t.Fatal("missing new entity ID accepted")
	}
	if err := cmd.ParseFlags([]string{"--entity-id", "switch.new", "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	if err := cmd.ValidateRequiredFlags(); err != nil {
		t.Fatal(err)
	}
}
