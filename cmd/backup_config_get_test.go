package cmd

import (
	"testing"
)

func TestRedactBackupPassword(t *testing.T) {
	cfg := map[string]interface{}{
		"create_backup": map[string]interface{}{
			"agent_ids": []interface{}{"backup.local"},
			"password":  "secret",
		},
	}
	redactBackupPassword(cfg)
	createBackup := cfg["create_backup"].(map[string]interface{})
	if _, ok := createBackup["password"]; ok {
		t.Fatal("password still present")
	}
	if createBackup["password_set"] != true {
		t.Fatalf("password_set = %v, want true", createBackup["password_set"])
	}

	cfg = map[string]interface{}{
		"create_backup": map[string]interface{}{"password": nil},
	}
	redactBackupPassword(cfg)
	if cfg["create_backup"].(map[string]interface{})["password_set"] != false {
		t.Fatal("password_set should be false for a null password")
	}
}
