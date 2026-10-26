package entityupdate

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fakeStore struct {
	calls    []string
	fail     string
	registry []any
	states   []any
	readBack map[string]any
	update   map[string]any
	ack      map[string]any
}

func (s *fakeStore) call(name string) error {
	s.calls = append(s.calls, name)
	if name == s.fail {
		return errors.New("test failure")
	}
	return nil
}

func (s *fakeStore) EntityRegistryGet(id string) (map[string]any, error) {
	err := s.call("get:" + id)
	if s.update != nil && s.readBack != nil {
		return s.readBack, err
	}
	entry := map[string]any{
		"entity_id": id, "unique_id": "integration-id", "name": "Old name", "icon": "mdi:old",
		"area_id": nil, "device_class": nil, "hidden_by": nil, "disabled_by": nil,
		"labels": []any{"old_label"}, "aliases": []any{"Old alias"},
		"categories": map[string]any{"script": "keep", "helpers": "old_category"},
	}
	for key, value := range s.update {
		if key == "categories" {
			for scope, category := range value.(map[string]any) {
				if category == nil {
					delete(entry[key].(map[string]any), scope)
				} else {
					entry[key].(map[string]any)[scope] = category
				}
			}
			continue
		}
		if key != "new_entity_id" {
			entry[key] = value
		}
	}
	return entry, err
}

func (s *fakeStore) EntityRegistryList() ([]any, error) {
	return s.registry, s.call("registry")
}

func (s *fakeStore) GetStates() ([]any, error) {
	return s.states, s.call("states")
}

func (s *fakeStore) SearchRelated(kind, id string) (map[string][]string, error) {
	return map[string][]string{"automation": {"automation.z", "automation.a"}}, s.call("related:" + kind + ":" + id)
}

func (s *fakeStore) EntityRegistryUpdate(id string, params map[string]any) (map[string]any, error) {
	s.update = params
	return s.ack, s.call("update:" + id)
}

func (s *fakeStore) AreaRegistryList() ([]any, error) {
	return []any{map[string]any{"area_id": "kitchen"}}, s.call("areas")
}

func (s *fakeStore) LabelRegistryList() ([]any, error) {
	return []any{map[string]any{"label_id": "lighting"}, map[string]any{"label_id": "downstairs"}}, s.call("labels")
}

func (s *fakeStore) CategoryRegistryList(scope string) ([]any, error) {
	return []any{map[string]any{"category_id": scope + "_category"}}, s.call("categories:" + scope)
}

func TestRenamePlanAndApply(t *testing.T) {
	for _, plan := range []bool{true, false} {
		t.Run(map[bool]string{true: "plan", false: "apply"}[plan], func(t *testing.T) {
			store := &fakeStore{}
			result, err := Run(store, "switch.old", Options{EntityID: stringPtr("switch.new")}, plan)
			if err != nil {
				t.Fatal(err)
			}
			if result.Saved == nil || *result.Saved != !plan || result.Verified != !plan {
				t.Fatalf("incorrect save/verification status: %+v", result)
			}
			if result.UniqueID != "integration-id" || len(result.Warnings) == 0 || !reflect.DeepEqual(result.Related["automation"], []string{"automation.a", "automation.z"}) {
				t.Fatalf("incomplete preview: %+v", result)
			}
			calls := []string{"get:switch.old", "registry", "states", "related:entity:switch.old"}
			if !plan {
				calls = append(calls, "update:switch.old", "get:switch.new")
				if !reflect.DeepEqual(store.update, map[string]any{"new_entity_id": "switch.new"}) {
					t.Fatalf("unexpected registry changes: %#v", store.update)
				}
			}
			if !reflect.DeepEqual(store.calls, calls) {
				t.Fatalf("calls = %v, want %v", store.calls, calls)
			}
		})
	}
}

func TestRenamePreflightFailuresDoNotWrite(t *testing.T) {
	for _, tc := range []struct {
		name  string
		store fakeStore
	}{
		{"registry collision including disabled entity", fakeStore{registry: []any{map[string]any{"entity_id": "switch.new", "disabled_by": "user"}}}},
		{"unregistered live entity collision", fakeStore{states: []any{map[string]any{"entity_id": "switch.new"}}}},
		{"source unavailable", fakeStore{fail: "get:switch.old"}},
		{"registry unavailable", fakeStore{fail: "registry"}},
		{"states unavailable", fakeStore{fail: "states"}},
		{"relationships unavailable", fakeStore{fail: "related:entity:switch.old"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, plan := range []bool{true, false} {
				result, err := Run(&tc.store, "switch.old", Options{EntityID: stringPtr("switch.new")}, plan)
				if err == nil || tc.store.update != nil || result.Saved == nil || *result.Saved || result.Verified {
					t.Fatalf("preflight failure allowed write: result=%+v update=%v err=%v", result, tc.store.update, err)
				}
			}
		})
	}
}

func TestRenameReportsSaveAndVerificationSeparately(t *testing.T) {
	for _, tc := range []struct {
		name      string
		store     fakeStore
		uncertain bool
	}{
		{"write not acknowledged", fakeStore{fail: "update:switch.old"}, true},
		{"read-back unavailable", fakeStore{fail: "get:switch.new"}, false},
		{"wrong entity ID", fakeStore{readBack: map[string]any{"entity_id": "switch.other", "unique_id": "integration-id"}}, false},
		{"changed unique ID", fakeStore{readBack: map[string]any{"entity_id": "switch.new", "unique_id": "different"}}, false},
		{"missing unique ID", fakeStore{readBack: map[string]any{"entity_id": "switch.new"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Run(&tc.store, "switch.old", Options{EntityID: stringPtr("switch.new")}, false)
			if err == nil || result.Verified || (result.Saved == nil) != tc.uncertain {
				t.Fatalf("incorrect failure status: %+v err=%v", result, err)
			}
			if !tc.uncertain && !*result.Saved {
				t.Fatal("acknowledged write not reported as saved")
			}
		})
	}
}

func TestRenameRejectsInvalidIDsBeforeIO(t *testing.T) {
	for _, ids := range [][2]string{
		{"switch.old", ""}, {"switch.old", "switch.old"}, {"switch.old", "light.new"},
		{"switch.old", "switch.New"}, {"switch.old", "switch.new-id"}, {"switch.old", "switch.new\n"},
		{"invalid", "switch.new"}, {"switch.old", "switch."},
	} {
		t.Run(strings.Join(ids[:], "->"), func(t *testing.T) {
			store := &fakeStore{}
			if _, err := Run(store, ids[0], Options{EntityID: &ids[1]}, false); err == nil || len(store.calls) != 0 {
				t.Fatalf("invalid IDs accepted: %v, calls=%v", err, store.calls)
			}
		})
	}
}

func stringPtr(value string) *string { return &value }

func TestUpdateNameAndIcon(t *testing.T) {
	for _, tc := range []struct {
		name   string
		opts   Options
		params map[string]any
	}{
		{"name only", Options{Name: stringPtr("Kitchen")}, map[string]any{"name": "Kitchen"}},
		{"icon only", Options{Icon: stringPtr("mdi:power")}, map[string]any{"icon": "mdi:power"}},
		{"clear name", Options{Name: stringPtr("")}, map[string]any{"name": nil}},
		{"clear icon", Options{Icon: stringPtr("")}, map[string]any{"icon": nil}},
		{"clear both", Options{Name: stringPtr(""), Icon: stringPtr("")}, map[string]any{"name": nil, "icon": nil}},
		{"same ID with name", Options{EntityID: stringPtr("switch.old"), Name: stringPtr("Kitchen")}, map[string]any{"name": "Kitchen"}},
		{"all fields", Options{EntityID: stringPtr("switch.new"), Name: stringPtr("Kitchen"), Icon: stringPtr("mdi:power")}, map[string]any{"new_entity_id": "switch.new", "name": "Kitchen", "icon": "mdi:power"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, plan := range []bool{true, false} {
				store := &fakeStore{}
				result, err := Run(store, "switch.old", tc.opts, plan)
				if err != nil {
					t.Fatal(err)
				}
				if result.Saved == nil || *result.Saved != !plan || result.Verified != !plan {
					t.Fatalf("incorrect result: %+v", result)
				}
				if plan {
					if store.update != nil {
						t.Fatal("preview wrote changes")
					}
				} else if !reflect.DeepEqual(store.update, tc.params) {
					t.Fatalf("update = %#v, want %#v", store.update, tc.params)
				}
				for key, want := range tc.params {
					if key == "new_entity_id" {
						key = "entity_id"
					}
					change, exists := result.Changes[key]
					if !exists || !reflect.DeepEqual(change.After, want) {
						t.Fatalf("incorrect preview: %+v", result.Changes)
					}
					before := map[string]any{"entity_id": "switch.old", "name": "Old name", "icon": "mdi:old"}[key]
					if !reflect.DeepEqual(change.Before, before) {
						t.Fatalf("incorrect before: %+v", change)
					}
				}
				if result.NewEntityID == result.OldEntityID {
					wantCalls := []string{"get:switch.old"}
					if !plan {
						wantCalls = append(wantCalls, "update:switch.old", "get:switch.old")
					}
					if !reflect.DeepEqual(store.calls, wantCalls) || len(result.Related) != 0 {
						t.Fatalf("name/icon-only update performed rename checks: %v", store.calls)
					}
				}
			}
		})
	}
}

func TestUpdateVerifiesEachRequestedField(t *testing.T) {
	for _, tc := range []struct {
		name     string
		opts     Options
		readBack map[string]any
	}{
		{"name mismatch", Options{Name: stringPtr("Kitchen")}, map[string]any{"name": "Old name"}},
		{"icon mismatch", Options{Icon: stringPtr("mdi:power")}, map[string]any{"icon": "mdi:old"}},
		{"clear not applied", Options{Name: stringPtr("")}, map[string]any{"name": "Old name"}},
		{"missing field is not cleared", Options{Icon: stringPtr("")}, map[string]any{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.readBack["entity_id"] = "switch.old"
			tc.readBack["unique_id"] = "integration-id"
			store := &fakeStore{readBack: tc.readBack}
			result, err := Run(store, "switch.old", tc.opts, false)
			if err == nil || result.Saved == nil || !*result.Saved || result.Verified {
				t.Fatalf("unverified fields reported as applied: %+v err=%v", result, err)
			}
		})
	}
}

func TestUpdateRequiresFieldsBeforeIO(t *testing.T) {
	store := &fakeStore{}
	if _, err := Run(store, "switch.old", Options{}, false); err == nil || len(store.calls) != 0 {
		t.Fatalf("empty update accepted: err=%v calls=%v", err, store.calls)
	}
}
