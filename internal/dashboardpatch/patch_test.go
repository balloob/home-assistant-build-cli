package dashboardpatch

import (
	"encoding/json"
	"testing"
)

func TestPatchRetainsLargeNumbers(t *testing.T) {
	config := map[string]any{"custom_number": json.Number("9007199254740993"), "title": "Before"}
	result, changes, err := Apply(config, Spec{Merge: map[string]any{"title": "After"}})
	if err != nil {
		t.Fatal(err)
	}
	if result["custom_number"] != json.Number("9007199254740993") || len(changes) != 1 {
		t.Fatalf("retained number changed: %v %+v", result, changes)
	}
}

func TestEditsDoNotMutateReusableSpec(t *testing.T) {
	value := map[string]any{"action": "toggle", "confirmation": true}
	spec := Spec{Target: "/views/0/cards/0", Edits: []Edit{{Path: "/tap_action", Value: value}, {Path: "/tap_action/confirmation", Remove: true}}}
	if _, _, err := Apply(fixture(), spec); err != nil {
		t.Fatal(err)
	}
	if value["confirmation"] != true {
		t.Fatal("patch changed its input spec")
	}
}

func TestExplicitPointers(t *testing.T) {
	for _, target := range []string{"views/0", "/views/-1", "/views/01", "/views/2", "/views/-", "/views/0/cards/0/entity", "/bad~2escape"} {
		t.Run(target, func(t *testing.T) {
			if _, _, err := Apply(fixture(), Spec{Target: target}); err == nil {
				t.Fatal("invalid or non-object selector accepted")
			}
		})
	}
	config := map[string]any{"a/b": map[string]any{"~key": "old", "keep": true}}
	result, changes, err := Apply(config, Spec{Target: "/a~1b", Edits: []Edit{{Path: "/~0key", Value: nil}, {Path: "/missing", Remove: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || !changes[0].AfterExists || changes[0].After != nil {
		t.Fatalf("null confused with deletion: %+v", changes)
	}
	if result["a/b"].(map[string]any)["keep"] != true || config["a/b"].(map[string]any)["~key"] != "old" {
		t.Fatal("input or retained field changed")
	}
}

func TestRemoveAndArrayReplacement(t *testing.T) {
	spec := Spec{Target: "/views/0/cards/0", Edits: []Edit{{Path: "/tap_action/confirmation", Remove: true}, {Path: "/features", Value: []any{map[string]any{"type": "light-brightness"}}}}}
	result, changes, err := Apply(fixture(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 {
		t.Fatalf("%+v", changes)
	}
	if _, err := resolve(result, []string{"views", "0", "cards", "0", "tap_action", "confirmation"}); err == nil {
		t.Fatal("field not removed")
	}
	for _, edit := range []Edit{{Path: ""}, {Path: "/missing/child", Value: true}, {Path: "/views/0", Remove: true}} {
		if _, _, err := Apply(fixture(), Spec{Edits: []Edit{edit}}); err == nil {
			t.Fatalf("unsafe edit accepted: %+v", edit)
		}
	}
}
