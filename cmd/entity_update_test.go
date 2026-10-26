package cmd

import (
	"reflect"
	"testing"

	"github.com/home-assistant/hab/internal/entityupdate"
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
		if constraint.Type == "at_least_one" && reflect.DeepEqual(constraint.Flags, entityUpdateFieldFlags) {
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

func TestEntityUpdateParsesRegistryFields(t *testing.T) {
	cmd := newEntityUpdateCommand()
	if err := cmd.ParseFlags([]string{"--area", "kitchen", "--device-class", "outlet", "--labels", "lighting,downstairs", "--aliases", `"Kitchen, plug",Socket`, "--category", "helpers=guest", "--category", "script=", "--hidden=false", "--disabled"}); err != nil {
		t.Fatal(err)
	}
	opts, err := parseEntityUpdateOptions(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if opts.EntityID != nil || opts.Name != nil || opts.Icon != nil {
		t.Fatal("omitted fields were supplied")
	}
	if *opts.AreaID != "kitchen" || *opts.DeviceClass != "outlet" || *opts.Hidden || !*opts.Disabled {
		t.Fatalf("incorrect scalar options: %+v", opts)
	}
	if !reflect.DeepEqual(*opts.Labels, []string{"lighting", "downstairs"}) || !reflect.DeepEqual(*opts.Aliases, []string{"Kitchen, plug", "Socket"}) || !reflect.DeepEqual(opts.Categories, map[string]string{"helpers": "guest", "script": ""}) {
		t.Fatalf("incorrect collection options: %+v", opts)
	}
}

func TestEntityUpdateClearsAndIndividualFields(t *testing.T) {
	for _, flags := range [][]string{{"--area", ""}, {"--device-class", ""}, {"--labels", ""}, {"--aliases", ""}, {"--category", "helpers="}, {"--hidden=false"}, {"--disabled=false"}} {
		t.Run(flags[0], func(t *testing.T) {
			cmd := newEntityUpdateCommand()
			if err := cmd.ParseFlags(flags); err != nil {
				t.Fatal(err)
			}
			if err := cmd.ValidateFlagGroups(); err != nil {
				t.Fatal(err)
			}
			opts, err := parseEntityUpdateOptions(cmd)
			if err != nil {
				t.Fatal(err)
			}
			if err := entityupdate.Validate("switch.old", opts); err != nil {
				t.Fatal(err)
			}
			if opts.Labels != nil && len(*opts.Labels) != 0 {
				t.Fatalf("clear labels = %v", *opts.Labels)
			}
			if opts.Aliases != nil && len(*opts.Aliases) != 0 {
				t.Fatalf("clear aliases = %v", *opts.Aliases)
			}
		})
	}
}

func TestEntityUpdateRejectsInvalidCategoriesAndOptions(t *testing.T) {
	for _, flags := range [][]string{{"--category", "helpers"}, {"--category", "=id"}, {"--category", "bad_scope=id"}, {"--category", "helpers=one", "--category", "helpers=two"}, {"--options", "{}"}} {
		cmd := newEntityUpdateCommand()
		if err := cmd.ParseFlags(flags); err != nil {
			continue
		}
		if _, err := parseEntityUpdateOptions(cmd); err == nil {
			t.Fatalf("accepted %v", flags)
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
