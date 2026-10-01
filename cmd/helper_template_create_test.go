package cmd

import (
	"sort"
	"testing"
)

// TestTemplateFormBuildersUseCoreKeys checks each builder against the keys
// that generate_schema in homeassistant/components/template/config_flow.py
// accepts for that template type.
func TestTemplateFormBuildersUseCoreKeys(t *testing.T) {
	saved := tplOpts
	t.Cleanup(func() { tplOpts = saved })

	tplOpts = templateCreateOpts{
		StateTemplate: "s", Unit: "u", DeviceClass: "d", StateClass: "c",
		TurnOn: "a", TurnOff: "a", Press: "a",
		Open: "a", Close: "a", Stop: "a", Position: "p", SetPos: "a",
		Lock: "a", Unlock: "a", URL: "u", SetValue: "a",
		Options: []string{"x"}, SelectOption: "a",
		Condition: "c", Temperature: "t", Humidity: "h",
		Brightness: "b", SetBrightness: "a", HS: "h", SetHS: "a", SetTemperature: "a",
		Percentage: "p", SetPct: "a",
		Start: "a", Pause: "a", ReturnToBase: "a", Clean: "a", Locate: "a",
		SetFanSpeed: "a", FanSpeed: "f",
	}

	coreKeys := map[string][]string{
		"alarm_control_panel": {"value_template"},
		"binary_sensor":       {"state", "device_class"},
		"button":              {"press"},
		"cover":               {"state", "device_class", "position", "open_cover", "close_cover", "stop_cover", "set_cover_position"},
		"fan":                 {"state", "turn_on", "turn_off", "percentage", "set_percentage"},
		"image":               {"url"},
		"light":               {"state", "turn_on", "turn_off", "level", "set_level", "hs", "set_hs", "temperature", "set_temperature"},
		"lock":                {"state", "lock", "unlock", "open"},
		"number":              {"state", "min", "max", "step", "set_value"},
		"select":              {"state", "options", "select_option"},
		"sensor":              {"state", "unit_of_measurement", "device_class", "state_class"},
		"switch":              {"value_template", "turn_on", "turn_off"},
		"vacuum":              {"state", "start", "stop", "pause", "return_to_base", "clean_spot", "locate", "set_fan_speed", "fan_speed"},
		"weather":             {"condition", "temperature", "humidity"},
	}

	for typ, builder := range templateFormBuilders {
		want, ok := coreKeys[typ]
		if !ok {
			t.Errorf("no core keys listed for template type %s", typ)
			continue
		}
		m := map[string]interface{}{}
		builder(m)
		got := make([]string, 0, len(m))
		for k := range m {
			got = append(got, k)
		}
		sort.Strings(got)
		sort.Strings(want)
		if len(got) != len(want) {
			t.Errorf("%s: keys = %v, want %v", typ, got, want)
			continue
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("%s: keys = %v, want %v", typ, got, want)
				break
			}
		}
	}
}
