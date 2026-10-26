package cmd

import (
	"github.com/home-assistant/hab/client"
	"testing"
)

type areaReferenceStore struct{ client.WebSocketAPI }

func (areaReferenceStore) FloorRegistryList() ([]any, error) {
	return []any{map[string]any{"floor_id": "ground"}}, nil
}
func (areaReferenceStore) LabelRegistryList() ([]any, error) {
	return []any{map[string]any{"label_id": "downstairs"}}, nil
}
func TestAreaUpdateReferences(t *testing.T) {
	for _, tt := range []struct {
		params    map[string]any
		wantError bool
	}{
		{map[string]any{"floor_id": "ground", "labels": []string{"downstairs"}}, false},
		{map[string]any{"floor_id": "ground", "labels": []string{"missing"}}, true},
		{map[string]any{"floor_id": "missing"}, true},
		{map[string]any{"floor_id": nil, "labels": []string{}}, false},
	} {
		if err := validateAreaUpdate(areaReferenceStore{}, tt.params); (err != nil) != tt.wantError {
			t.Fatalf("params=%v error=%v", tt.params, err)
		}
	}
}
func TestAreaUpdateFlags(t *testing.T) {
	cmd, err := resolveSchemaTarget([]string{"area", "update"})
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"aliases", "labels", "picture", "temperature-entity", "humidity-entity"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Fatalf("missing --%s", flag)
		}
	}
}
