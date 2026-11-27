package cmd

import (
	"strings"
	"testing"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/config"
)

func TestESPHomeEnvOverridesRejectedForNamedContext(t *testing.T) {
	for _, name := range []string{"HAB_URL", "HAB_TOKEN", "HAB_REFRESH_TOKEN", config.ContextEnv, "HAB_ESPHOME_URL", "HAB_ESPHOME_TOKEN", "HA_ACCESS_TOKEN"} {
		t.Setenv(name, "")
	}
	root := t.TempDir()
	for _, dir := range []string{config.GetContextDir(root, config.DefaultContext), config.GetContextDir(root, "brother")} {
		if err := auth.SaveCredentials(&auth.Credentials{URL: "http://instance", AccessToken: "token"}, dir); err != nil {
			t.Fatal(err)
		}
	}

	for _, env := range []string{"HAB_ESPHOME_URL", "HAB_ESPHOME_TOKEN", "HA_ACCESS_TOKEN"} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("HAB_ESPHOME_URL", "http://other-esphome")
			t.Setenv(env, "set")
			if env != "HAB_ESPHOME_URL" {
				t.Setenv("HAB_ESPHOME_URL", "")
			}

			useLoginManager(t, auth.NewContextManager(root, "brother"))
			if _, err := getESPHomeClient(); err == nil || !strings.Contains(err.Error(), "conflicting ESPHome settings") {
				t.Fatalf("named context: err = %v", err)
			}
		})
	}

	// The default context keeps the existing overrides.
	t.Setenv("HAB_ESPHOME_URL", "http://other-esphome")
	useLoginManager(t, auth.NewContextManager(root, ""))
	if _, err := getESPHomeClient(); err != nil {
		t.Fatalf("default context: %v", err)
	}
}
