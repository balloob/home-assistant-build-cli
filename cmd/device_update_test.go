package cmd

import (
	"reflect"
	"testing"

	"github.com/home-assistant/hab/client"
)

type deviceReferenceStore struct {
	client.WebSocketAPI
}

func (deviceReferenceStore) AreaRegistryList() ([]any, error) {
	return []any{map[string]any{"area_id": "kitchen"}}, nil
}
func (deviceReferenceStore) LabelRegistryList() ([]any, error) {
	return []any{map[string]any{"label_id": "lighting"}}, nil
}

func TestDeviceUpdateReferences(t *testing.T) {
	for _, tt := range []struct {
		params    map[string]any
		wantError bool
	}{
		{map[string]any{"area_id": "kitchen", "labels": []string{"lighting"}}, false},
		{map[string]any{"area_id": "missing"}, true},
		{map[string]any{"labels": []string{"missing"}}, true},
		{map[string]any{"area_id": nil, "labels": []string{}}, false},
	} {
		if err := validateDeviceUpdate(deviceReferenceStore{}, tt.params); (err != nil) != tt.wantError {
			t.Fatalf("params=%v error=%v", tt.params, err)
		}
	}
}

func TestDeviceUpdateCommand(t *testing.T) {
	cmd := newDeviceUpdateCommand()
	if err := cmd.ParseFlags([]string{"--name", "", "--area", "", "--labels", "", "--disabled=false"}); err != nil {
		t.Fatal(err)
	}
	defs := []RegistryFlagDef{
		{Name: "name", ParamKey: "name_by_user", Type: FlagString},
		{Name: "area", ParamKey: "area_id", Type: FlagString},
		{Name: "labels", ParamKey: "labels", Type: FlagStringSet},
		{Name: "disabled", ParamKey: "disabled_by", Type: FlagUserBool},
	}
	got, err := registryUpdateParams(cmd, defs, true)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"name_by_user": nil, "area_id": nil, "labels": []string{}, "disabled_by": nil}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v", got)
	}
	schemaCmd, err := resolveSchemaTarget([]string{"device", "update"})
	if err != nil {
		t.Fatal(err)
	}
	schema := buildSchemaTree(schemaCmd)
	if len(schema.Args) != 1 || schema.Args[0].Name != "device_id" {
		t.Fatalf("wrong arguments: %v", schema.Args)
	}
}
