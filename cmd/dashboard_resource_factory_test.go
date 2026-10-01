package cmd

import "testing"

func TestViewIndexByPath(t *testing.T) {
	config := map[string]interface{}{
		"views": []interface{}{
			map[string]interface{}{"title": "Home", "path": "home"},
			map[string]interface{}{"title": "Lights", "path": "lights"},
		},
	}

	index, err := viewIndexByPath(config, "lights")
	if err != nil || index != 1 {
		t.Fatalf("got %d, %v; want 1", index, err)
	}

	if _, err := viewIndexByPath(config, "garage"); err == nil {
		t.Fatal("expected an error for an unknown path")
	}
}
