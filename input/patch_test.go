package input

import "testing"

func TestParseValueScalar(t *testing.T) {
	value, err := ParseValue("true")
	if err != nil {
		t.Fatalf("ParseValue: %v", err)
	}
	parsed, ok := value.(bool)
	if !ok || !parsed {
		t.Fatalf("value = %#v, want true", value)
	}
}

func TestParseValueObject(t *testing.T) {
	value, err := ParseValue("foo:\n  bar: 1\n")
	if err != nil {
		t.Fatalf("ParseValue: %v", err)
	}
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value = %#v, want map", value)
	}
	inner := object["foo"].(map[string]any)
	if inner["bar"] != 1 {
		t.Fatalf("bar = %#v, want 1", inner["bar"])
	}
}

func TestMergeMaps(t *testing.T) {
	merged := MergeMaps(
		map[string]any{
			"wifi":   map[string]any{"ssid": "old"},
			"logger": map[string]any{"level": "INFO"},
		},
		map[string]any{
			"wifi": map[string]any{"password": "secret"},
		},
	)

	wifi := merged["wifi"].(map[string]any)
	if wifi["ssid"] != "old" {
		t.Fatalf("ssid = %#v, want old", wifi["ssid"])
	}
	if wifi["password"] != "secret" {
		t.Fatalf("password = %#v, want secret", wifi["password"])
	}
}

func TestSetPathValueNestedMap(t *testing.T) {
	root := map[string]any{}
	if err := SetPathValue(root, "wifi.ssid", "MyWifi"); err != nil {
		t.Fatalf("SetPathValue: %v", err)
	}
	wifi := root["wifi"].(map[string]any)
	if wifi["ssid"] != "MyWifi" {
		t.Fatalf("ssid = %#v, want MyWifi", wifi["ssid"])
	}
}

func TestSetPathValueListIndex(t *testing.T) {
	root := map[string]any{}
	if err := SetPathValue(root, "ota[0].password", "updated"); err != nil {
		t.Fatalf("SetPathValue: %v", err)
	}
	ota := root["ota"].([]any)
	first := ota[0].(map[string]any)
	if first["password"] != "updated" {
		t.Fatalf("password = %#v, want updated", first["password"])
	}
}

func TestSetPathValueRejectsIndexRoot(t *testing.T) {
	if err := SetPathValue(map[string]any{}, "[0].foo", "bar"); err == nil {
		t.Fatal("expected error for root list index")
	}
}

func TestMarshalYAML(t *testing.T) {
	result, err := MarshalYAML(map[string]any{"wifi": map[string]any{"ssid": "guest"}})
	if err != nil {
		t.Fatalf("MarshalYAML: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty YAML output")
	}
}
