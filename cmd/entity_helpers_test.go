package cmd

import (
	"fmt"
	"testing"

	"github.com/home-assistant/hab/client"
)

func TestAppendStatelessEntitiesListsRegistryOnlyEntries(t *testing.T) {
	states := []interface{}{
		map[string]interface{}{"entity_id": "sensor.live", "state": "1", "attributes": map[string]interface{}{"friendly_name": "Live"}},
	}
	registry := []interface{}{
		map[string]interface{}{"entity_id": "sensor.live", "disabled_by": nil},
		map[string]interface{}{"entity_id": "sensor.off", "disabled_by": "user", "name": nil, "original_name": "Off", "device_id": "dev1"},
		map[string]interface{}{"entity_id": "light.off", "disabled_by": "integration", "name": "Lamp"},
	}

	all := appendStatelessEntities(states, registry)
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}

	entities := filterAndBuildEntities(all, buildRegistryMap(registry), nil, nil, entityFilterCriteria{Domain: "sensor", IncludeDisabled: true})
	if len(entities) != 2 {
		t.Fatalf("expected 2 sensor entities, got %d", len(entities))
	}
	live, off := entities[0], entities[1]
	if live["state"] != "1" || live["disabled_by"] != nil || live["disabled"] != false {
		t.Fatalf("unexpected live entity: %#v", live)
	}
	if off["entity_id"] != "sensor.off" || off["state"] != nil || off["disabled_by"] != "user" || off["disabled"] != true || off["name"] != "Off" || off["device_id"] != "dev1" {
		t.Fatalf("unexpected stateless entity: %#v", off)
	}
}

func TestFilterAndBuildEntitiesOmitsDisabledByByDefault(t *testing.T) {
	states := []interface{}{map[string]interface{}{"entity_id": "sensor.live", "state": "1"}}
	entities := filterAndBuildEntities(states, nil, nil, nil, entityFilterCriteria{})
	if _, ok := entities[0]["disabled_by"]; ok {
		t.Fatalf("disabled_by should not be present without IncludeDisabled: %#v", entities[0])
	}
}

func TestIsNotFoundError(t *testing.T) {
	if !isNotFoundError(fmt.Errorf("wrapped: %w", &client.APIError{Code: client.ErrCodeNotFound})) {
		t.Fatal("expected NOT_FOUND to match")
	}
	if isNotFoundError(&client.APIError{Code: client.ErrCodeAuthenticationError}) {
		t.Fatal("auth error must not match")
	}
	if isNotFoundError(fmt.Errorf("network down")) {
		t.Fatal("plain error must not match")
	}
}

func TestStatelessEntityNote(t *testing.T) {
	if got := statelessEntityNote("sensor.x", map[string]interface{}{"disabled_by": "user"}); got != "sensor.x is disabled (disabled_by: user) and has no state." {
		t.Fatalf("unexpected note: %q", got)
	}
	if got := statelessEntityNote("sensor.x", map[string]interface{}{"disabled_by": nil}); got != "sensor.x is in the entity registry but has no state." {
		t.Fatalf("unexpected note: %q", got)
	}
}
