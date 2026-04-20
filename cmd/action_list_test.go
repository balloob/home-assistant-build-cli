package cmd

import "testing"

func TestCollectActionsReturnsStableSortedOrder(t *testing.T) {
	services := []interface{}{
		map[string]interface{}{
			"domain": "light",
			"services": map[string]interface{}{
				"turn_on":  map[string]interface{}{"name": "Turn on"},
				"turn_off": map[string]interface{}{"name": "Turn off"},
			},
		},
		map[string]interface{}{
			"domain": "climate",
			"services": map[string]interface{}{
				"set_temperature": map[string]interface{}{"name": "Set temperature"},
			},
		},
	}

	result := collectActions(services, "", nil)
	if len(result) != 3 {
		t.Fatalf("len(result) = %d, want 3", len(result))
	}

	got := []string{
		result[0]["action"].(string),
		result[1]["action"].(string),
		result[2]["action"].(string),
	}
	want := []string{"climate.set_temperature", "light.turn_off", "light.turn_on"}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("action[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
