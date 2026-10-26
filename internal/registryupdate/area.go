package registryupdate

import (
	"fmt"
	"strings"
)

// ValidateAreaSensors mirrors HA's requirement for existing sensor states with
// the corresponding temperature/humidity device class. Clearing needs no lookup.
func ValidateAreaSensors(list func() ([]any, error), params map[string]any) error {
	sensors := map[string]string{}
	for _, class := range []string{"temperature", "humidity"} {
		if id, ok := params[class+"_entity_id"].(string); ok {
			sensors[class] = id
		}
	}
	if len(sensors) == 0 {
		return nil
	}
	states, err := list()
	if err != nil {
		return fmt.Errorf("validate area sensors: %w", err)
	}
	for class, id := range sensors {
		valid := false
		for _, item := range states {
			state, ok := item.(map[string]any)
			if !ok || state["entity_id"] != id {
				continue
			}
			attrs, ok := state["attributes"].(map[string]any)
			valid = ok && strings.HasPrefix(id, "sensor.") && attrs["device_class"] == class
			break
		}
		if !valid {
			return fmt.Errorf("%s must be an existing %s sensor", id, class)
		}
	}
	return nil
}
