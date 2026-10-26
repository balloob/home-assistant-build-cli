package registryupdate

import (
	"errors"
	"maps"
	"testing"
)

type fakeStore struct {
	entry             map[string]any
	reads, writes     int
	writeErr, readErr error
	readback          map[string]any
}

func (s *fakeStore) Get(string) (map[string]any, error) {
	s.reads++
	if s.reads > 1 {
		if s.readErr != nil {
			return nil, s.readErr
		}
		if s.readback != nil {
			return s.readback, nil
		}
	}
	return maps.Clone(s.entry), nil
}
func (s *fakeStore) Update(_ string, p map[string]any) (map[string]any, error) {
	s.writes++
	maps.Copy(s.entry, p)
	return nil, s.writeErr
}

func TestRun(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		plan                    bool
		params                  map[string]any
		readback                map[string]any
		writeErr, readErr       error
		wantWrites              int
		wantSaved               *bool
		wantVerified, wantError bool
	}{
		{name: "preview", plan: true, params: map[string]any{"icon": nil}, wantSaved: boolPtr(false)},
		{name: "clear", params: map[string]any{"icon": nil}, wantWrites: 1, wantSaved: boolPtr(true), wantVerified: true},
		{name: "noop", params: map[string]any{"name": "Kitchen"}, wantSaved: boolPtr(false), wantVerified: true},
		{name: "numeric noop", params: map[string]any{"level": 0}, wantSaved: boolPtr(false), wantVerified: true},
		{name: "set order", params: map[string]any{"aliases": []string{"B", "A"}}, wantSaved: boolPtr(false), wantVerified: true},
		{name: "unknown field", params: map[string]any{"unsupported": "x"}, wantSaved: boolPtr(false), wantError: true},
		{name: "uncertain write", params: map[string]any{"icon": nil}, writeErr: errors.New("disconnected"), wantWrites: 1, wantError: true},
		{name: "failed readback", params: map[string]any{"icon": nil}, readErr: errors.New("disconnected"), wantWrites: 1, wantSaved: boolPtr(true), wantError: true},
		{name: "mismatch", params: map[string]any{"icon": nil}, readback: map[string]any{"icon": "mdi:home"}, wantWrites: 1, wantSaved: boolPtr(true), wantError: true},
		{name: "missing cleared field", params: map[string]any{"icon": nil}, readback: map[string]any{}, wantWrites: 1, wantSaved: boolPtr(true), wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{entry: map[string]any{"name": "Kitchen", "icon": "mdi:home", "level": float64(0), "aliases": []any{"A", "B"}}, readback: tt.readback, writeErr: tt.writeErr, readErr: tt.readErr}
			result, err := Run(store, "id", tt.params, tt.plan)
			if (err != nil) != tt.wantError {
				t.Fatalf("error=%v", err)
			}
			if store.writes != tt.wantWrites || result.Verified != tt.wantVerified {
				t.Fatalf("writes=%d result=%+v", store.writes, result)
			}
			if (result.Saved == nil) != (tt.wantSaved == nil) || result.Saved != nil && *result.Saved != *tt.wantSaved {
				t.Fatalf("saved=%v", result.Saved)
			}
			if tt.plan && (store.reads != 1 || result.Changes["icon"].Before != "mdi:home" || !result.WouldChange) {
				t.Fatalf("bad preview: %+v", result)
			}
		})
	}
}
func boolPtr(b bool) *bool { return &b }

func TestRequireIDs(t *testing.T) {
	calls := 0
	list := func() ([]any, error) { calls++; return []any{map[string]any{"label_id": "known"}}, nil }
	if err := RequireIDs(list, "label_id", nil); err != nil || calls != 0 {
		t.Fatal("clear should not list references")
	}
	if err := RequireIDs(list, "label_id", []string{"known"}); err != nil {
		t.Fatal(err)
	}
	if err := RequireIDs(list, "label_id", []string{"missing"}); err == nil {
		t.Fatal("unknown reference accepted")
	}
}
