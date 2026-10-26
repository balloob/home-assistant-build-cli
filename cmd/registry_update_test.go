package cmd

import (
	"reflect"
	"testing"
)

func TestRegistryUpdateParams(t *testing.T) {
	defs := []RegistryFlagDef{
		{Name: "icon", ParamKey: "icon", Type: FlagString},
		{Name: "level", ParamKey: "level", Type: FlagInt},
		{Name: "aliases", ParamKey: "aliases", Type: FlagStringSet},
		{Name: "disabled", ParamKey: "disabled_by", Type: FlagUserBool},
	}
	for _, tt := range []struct {
		name         string
		flags        []string
		nullableName bool
		want         map[string]any
		wantError    bool
	}{
		{name: "empty request", wantError: true},
		{name: "clear icon", flags: []string{"--icon", ""}, want: map[string]any{"icon": nil}},
		{name: "zero level", flags: []string{"--level", "0"}, want: map[string]any{"level": 0}},
		{name: "clear level", flags: []string{"--level", ""}, want: map[string]any{"level": nil}},
		{name: "invalid level", flags: []string{"--level", "abc"}, wantError: true},
		{name: "empty required name", flags: []string{"--name", ""}, wantError: true},
		{name: "clear name override", flags: []string{"--name", ""}, nullableName: true, want: map[string]any{"name": nil}},
		{name: "explicit false", flags: []string{"--disabled=false"}, want: map[string]any{"disabled_by": nil}},
		{name: "disabled", flags: []string{"--disabled"}, want: map[string]any{"disabled_by": "user"}},
		{name: "aliases", flags: []string{"--aliases", "B, A,B", "--aliases", "C"}, want: map[string]any{"aliases": []string{"A", "B", "C"}}},
		{name: "clear aliases", flags: []string{"--aliases", ""}, want: map[string]any{"aliases": []string{}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRegistryUpdateCommand(RegistryCRUDConfig{ResourceName: "test", IDField: "id", UpdateFlags: defs, NullableName: tt.nullableName})
			if err := cmd.ParseFlags(tt.flags); err != nil {
				t.Fatal(err)
			}
			all := append([]RegistryFlagDef{{Name: "name", ParamKey: "name", Type: FlagString}}, defs...)
			got, err := registryUpdateParams(cmd, all, tt.nullableName)
			if (err != nil) != tt.wantError {
				t.Fatalf("error=%v", err)
			}
			if !tt.wantError && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got=%#v want=%#v", got, tt.want)
			}
		})
	}
}

func TestRegistryUpdateSchema(t *testing.T) {
	for _, resource := range []string{"area", "floor", "label"} {
		cmd, err := resolveSchemaTarget([]string{resource, "update"})
		if err != nil {
			t.Fatal(err)
		}
		schema := buildSchemaTree(cmd)
		for _, variant := range schema.OutputContract.Variants {
			found := false
			for _, field := range variant.Data.Fields {
				if field.Name == "saved" {
					found = true
				}
				if field.Name == "change_detection" {
					t.Fatal("advertises static plan")
				}
			}
			if !found {
				t.Fatalf("missing saved in %s %s", resource, variant.Name)
			}
		}
	}
}
