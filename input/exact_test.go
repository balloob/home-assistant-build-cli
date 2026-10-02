package input

import (
	"encoding/json"
	"testing"
)

func TestParseInputExactRetainsNumbers(t *testing.T) {
	for _, tc := range []struct{ format, body string }{
		{"json", `{"value":9007199254740993}`},
		{"yaml", "value: 9007199254740993"},
	} {
		t.Run(tc.format, func(t *testing.T) {
			result, err := ParseInputExact(tc.body, "", tc.format)
			if err != nil {
				t.Fatal(err)
			}
			if result["value"] != json.Number("9007199254740993") {
				t.Fatalf("precision lost: %#v", result)
			}
		})
	}
	for _, body := range []string{`{"value":1} {"value":2}`, `[1]`, `{"value":`} {
		if _, err := ParseInputExact(body, "", "json"); err == nil {
			t.Fatalf("invalid object accepted: %s", body)
		}
	}
}
