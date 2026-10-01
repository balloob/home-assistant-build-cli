package client

import "testing"

func TestAggregateHealthEventsFailedUpdate(t *testing.T) {
	eventCh := make(chan map[string]interface{}, 3)
	data := map[string]interface{}{}
	doneCh := make(chan struct{})

	// Event shapes match components/system_health/__init__.py in core.
	eventCh <- map[string]interface{}{
		"type": "initial",
		"data": map[string]interface{}{
			"cloud": map[string]interface{}{
				"info": map[string]interface{}{
					"can_reach_cloud": map[string]interface{}{"type": "pending"},
				},
			},
		},
	}
	eventCh <- map[string]interface{}{
		"type":    "update",
		"domain":  "cloud",
		"key":     "can_reach_cloud",
		"success": false,
		"error":   map[string]interface{}{"type": "failed", "error": "unknown"},
	}
	eventCh <- map[string]interface{}{"type": "finish"}

	aggregateHealthEvents(eventCh, data, doneCh)

	info := data["cloud"].(map[string]interface{})["info"].(map[string]interface{})
	got, ok := info["can_reach_cloud"].(map[string]interface{})
	if !ok {
		t.Fatalf("can_reach_cloud = %#v, want error map", info["can_reach_cloud"])
	}
	if got["error"] != true || got["value"] != "unknown" {
		t.Errorf("can_reach_cloud = %#v, want error=true value=unknown", got)
	}
}
