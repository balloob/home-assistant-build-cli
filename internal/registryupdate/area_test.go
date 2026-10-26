package registryupdate

import (
	"errors"
	"testing"
)

func TestValidateAreaSensors(t *testing.T) {
	for _, tt := range []struct {
		name                string
		params              map[string]any
		failList, wantError bool
		calls               int
	}{
		{name: "unrelated", params: map[string]any{"name": "Kitchen"}},
		{name: "clear", params: map[string]any{"temperature_entity_id": nil, "humidity_entity_id": nil}},
		{name: "valid", params: map[string]any{"temperature_entity_id": "sensor.temp", "humidity_entity_id": "sensor.humidity"}, calls: 1},
		{name: "wrong class", params: map[string]any{"temperature_entity_id": "sensor.humidity"}, wantError: true, calls: 1},
		{name: "missing", params: map[string]any{"temperature_entity_id": "sensor.missing"}, wantError: true, calls: 1},
		{name: "wrong domain", params: map[string]any{"temperature_entity_id": "binary_sensor.temp"}, wantError: true, calls: 1},
		{name: "read failure", params: map[string]any{"temperature_entity_id": "sensor.temp"}, failList: true, wantError: true, calls: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			list := func() ([]any, error) {
				calls++
				if tt.failList {
					return nil, errors.New("disconnected")
				}
				return []any{
					map[string]any{"entity_id": "sensor.temp", "attributes": map[string]any{"device_class": "temperature"}},
					map[string]any{"entity_id": "sensor.humidity", "attributes": map[string]any{"device_class": "humidity"}},
					map[string]any{"entity_id": "binary_sensor.temp", "attributes": map[string]any{"device_class": "temperature"}},
				}, nil
			}
			err := ValidateAreaSensors(list, tt.params)
			if (err != nil) != tt.wantError || calls != tt.calls {
				t.Fatalf("error=%v calls=%d", err, calls)
			}
		})
	}
}
