package guide

import (
	"errors"
	"testing"
)

func TestTopicIDsDeterministicOrder(t *testing.T) {
	got := TopicIDs()
	want := []string{"index", "discovery", "auth", "input-output", "registry", "automation", "dashboard", "helpers", "calendar-todo", "esphome", "operations"}

	if len(got) != len(want) {
		t.Fatalf("TopicIDs length = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("TopicIDs[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolveAlias(t *testing.T) {
	tests := []struct {
		name  string
		alias string
		want  string
	}{
		{name: "canonical", alias: "dashboard", want: "dashboard"},
		{name: "alias", alias: "dashboards", want: "dashboard"},
		{name: "underscore normalized", alias: "getting_started", want: "index"},
		{name: "new topic alias", alias: "data entry", want: "input-output"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic, err := Resolve(tt.alias)
			if err != nil {
				t.Fatalf("Resolve(%q) error = %v", tt.alias, err)
			}
			if topic.ID != tt.want {
				t.Fatalf("Resolve(%q).ID = %q, want %q", tt.alias, topic.ID, tt.want)
			}
		})
	}
}

func TestResolveMissingTopic(t *testing.T) {
	_, err := Resolve("does-not-exist")
	if err == nil {
		t.Fatal("Resolve returned nil error for missing topic")
	}
	if !errors.Is(err, ErrTopicNotFound) {
		t.Fatalf("Resolve error = %v, want ErrTopicNotFound", err)
	}
}

func TestGetTopic(t *testing.T) {
	topic, err := GetTopic("index")
	if err != nil {
		t.Fatalf("GetTopic(index) error = %v", err)
	}
	if topic.Topic.ID != "index" {
		t.Fatalf("GetTopic(index).Topic.ID = %q, want index", topic.Topic.ID)
	}
	if topic.Content == "" {
		t.Fatal("GetTopic(index).Content is empty")
	}
}
