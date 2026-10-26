package entityupdate

import (
	"reflect"
	"strings"
	"testing"
)

func boolPtr(value bool) *bool              { return &value }
func stringsPtr(values ...string) *[]string { return &values }

func TestRegistryFields(t *testing.T) {
	for _, tc := range []struct {
		name   string
		opts   Options
		params map[string]any
	}{
		{"area", Options{AreaID: stringPtr("kitchen")}, map[string]any{"area_id": "kitchen"}},
		{"clear area", Options{AreaID: stringPtr("")}, map[string]any{"area_id": nil}},
		{"device class", Options{DeviceClass: stringPtr("outlet")}, map[string]any{"device_class": "outlet"}},
		{"clear device class", Options{DeviceClass: stringPtr("")}, map[string]any{"device_class": nil}},
		{"labels", Options{Labels: stringsPtr("lighting", "downstairs", "lighting")}, map[string]any{"labels": []any{"downstairs", "lighting"}}},
		{"clear labels", Options{Labels: stringsPtr()}, map[string]any{"labels": []any{}}},
		{"ordered aliases", Options{Aliases: stringsPtr(" Kitchen plug ", "", " Socket ")}, map[string]any{"aliases": []any{"Kitchen plug", "Socket"}}},
		{"clear aliases", Options{Aliases: stringsPtr()}, map[string]any{"aliases": []any{}}},
		{"categories", Options{Categories: map[string]string{"helpers": "helpers_category", "automation": "automation_category"}}, map[string]any{"categories": map[string]any{"helpers": "helpers_category", "automation": "automation_category"}}},
		{"remove category", Options{Categories: map[string]string{"helpers": ""}}, map[string]any{"categories": map[string]any{"helpers": nil}}},
		{"hidden", Options{Hidden: boolPtr(true)}, map[string]any{"hidden_by": "user"}},
		{"visible", Options{Hidden: boolPtr(false)}, map[string]any{"hidden_by": nil}},
		{"disabled", Options{Disabled: boolPtr(true)}, map[string]any{"disabled_by": "user"}},
		{"enabled", Options{Disabled: boolPtr(false)}, map[string]any{"disabled_by": nil}},
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
					t.Fatalf("params=%#v want=%#v", store.update, tc.params)
				}
				for key, want := range tc.params {
					if key == "categories" {
						if result.Changes[key].After.(map[string]any)["script"] != "keep" {
							t.Fatal("unrelated category lost")
						}
						if result.Changes[key].Before.(map[string]any)["helpers"] != "old_category" {
							t.Fatal("source categories were mutated")
						}
						continue
					}
					if !reflect.DeepEqual(result.Changes[key].After, want) {
						t.Fatalf("incorrect plan: %+v", result.Changes)
					}
				}
			}
		})
	}
}

func TestInvalidReferencesPreventAllWrites(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts Options
		fail string
	}{
		{"unknown area", Options{AreaID: stringPtr("missing")}, ""},
		{"unknown label", Options{Labels: stringsPtr("lighting", "missing")}, ""},
		{"unknown category", Options{Categories: map[string]string{"helpers": "missing"}}, ""},
		{"area read fails", Options{AreaID: stringPtr("kitchen")}, "areas"},
		{"label read fails", Options{Labels: stringsPtr("lighting")}, "labels"},
		{"category read fails", Options{Categories: map[string]string{"helpers": "helpers_category"}}, "categories:helpers"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts.Name = stringPtr("Must not be saved")
			for _, plan := range []bool{true, false} {
				store := &fakeStore{fail: tc.fail}
				result, err := Run(store, "switch.old", tc.opts, plan)
				if err == nil || store.update != nil || *result.Saved || result.Verified {
					t.Fatalf("partial update: %+v err=%v", result, err)
				}
			}
		})
	}
}

func TestCollectionVerification(t *testing.T) {
	for _, tc := range []struct {
		name   string
		opts   Options
		fields map[string]any
		valid  bool
	}{
		{"label order", Options{Labels: stringsPtr("downstairs", "lighting")}, map[string]any{"labels": []any{"lighting", "downstairs"}}, true},
		{"missing label", Options{Labels: stringsPtr("downstairs", "lighting")}, map[string]any{"labels": []any{"lighting"}}, false},
		{"alias order matters", Options{Aliases: stringsPtr("Primary", "Secondary")}, map[string]any{"aliases": []any{"Secondary", "Primary"}}, false},
		{"removed category still present", Options{Categories: map[string]string{"helpers": ""}}, map[string]any{"categories": map[string]any{"script": "keep", "helpers": "old_category"}}, false},
		{"other category lost", Options{Categories: map[string]string{"helpers": "helpers_category"}}, map[string]any{"categories": map[string]any{"helpers": "helpers_category"}}, false},
		{"device still disabled", Options{Disabled: boolPtr(false)}, map[string]any{"disabled_by": "device"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.fields["entity_id"], tc.fields["unique_id"] = "switch.old", "integration-id"
			store := &fakeStore{readBack: tc.fields}
			result, err := Run(store, "switch.old", tc.opts, false)
			if (err == nil) != tc.valid || result.Verified != tc.valid || result.Saved == nil || !*result.Saved {
				t.Fatalf("incorrect verification: %+v err=%v", result, err)
			}
		})
	}
}

func TestEnableActivationWarnings(t *testing.T) {
	for _, tc := range []struct {
		ack     map[string]any
		warning string
	}{
		{map[string]any{"require_restart": true}, "restart is required"},
		{map[string]any{"reload_delay": float64(5)}, "reload delay of 5 seconds"},
	} {
		result, err := Run(&fakeStore{ack: tc.ack}, "switch.old", Options{Disabled: boolPtr(false)}, false)
		if err != nil || !result.Verified || !strings.Contains(strings.Join(result.Warnings, " "), tc.warning) {
			t.Fatalf("missing activation warning: %+v err=%v", result, err)
		}
	}
}
