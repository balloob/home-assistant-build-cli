package entityrename

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
	if id == "switch.new" && s.readBack != nil {
		return s.readBack, err
	}
	return map[string]any{"entity_id": id, "unique_id": "integration-id"}, err
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
	return nil, s.call("update:" + id)
}

func TestRenamePlanAndApply(t *testing.T) {
	for _, plan := range []bool{true, false} {
		t.Run(map[bool]string{true: "plan", false: "apply"}[plan], func(t *testing.T) {
			store := &fakeStore{}
			result, err := Run(store, "switch.old", "switch.new", plan)
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
				result, err := Run(&tc.store, "switch.old", "switch.new", plan)
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
			result, err := Run(&tc.store, "switch.old", "switch.new", false)
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
			if _, err := Run(store, ids[0], ids[1], false); err == nil || len(store.calls) != 0 {
				t.Fatalf("invalid IDs accepted: %v, calls=%v", err, store.calls)
			}
		})
	}
}
