package redact

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCredentialKeysAndURLs(t *testing.T) {
	for _, key := range []string{"access-token", "api-key", "api-encryption-key", "/views/0/cards/0/private_key", "/auth/key", "passphrase", "psk", "alarm_pin"} {
		t.Run(key, func(t *testing.T) {
			if Value(key, "sensitive-value") != Hidden {
				t.Fatal("credential key was not redacted")
			}
		})
	}
	original := map[string]any{"config": map[string]any{"api-key": "sensitive-value"}, "url": "https://user:sensitive-value@example.com/path?api-key=sensitive-value&view=home"}
	redacted := Map(original)
	encoded, err := json.Marshal(redacted)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "sensitive-value") {
		t.Fatal("credential leaked")
	}
	if original["config"].(map[string]any)["api-key"] != "sensitive-value" {
		t.Fatal("redaction changed mutation input")
	}
	if !strings.Contains(string(encoded), "view=home") {
		t.Fatal("unrelated URL query was removed")
	}
}
