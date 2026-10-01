package cmd

import (
	"testing"
)

type fakeBackupAgents struct {
	agentIDs []string
}

func (f fakeBackupAgents) BackupAgentsInfo() (map[string]interface{}, error) {
	agents := []interface{}{}
	for _, id := range f.agentIDs {
		agents = append(agents, map[string]interface{}{"agent_id": id, "name": id})
	}
	return map[string]interface{}{"agents": agents}, nil
}

func TestFindLocalBackupAgent(t *testing.T) {
	tests := []struct {
		name     string
		agentIDs []string
		want     string
		wantErr  bool
	}{
		{name: "core", agentIDs: []string{"cloud.cloud", "backup.local"}, want: "backup.local"},
		{name: "supervisor", agentIDs: []string{"backup.local", "hassio.local"}, want: "hassio.local"},
		{name: "none", agentIDs: []string{"cloud.cloud"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findLocalBackupAgent(fakeBackupAgents{agentIDs: tt.agentIDs})
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

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
