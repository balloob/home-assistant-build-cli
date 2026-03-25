package cmd

import (
	"strings"
	"testing"

	"github.com/home-assistant/hab/guide"
)

func TestBuildGuideTopicOutput(t *testing.T) {
	topic := guide.TopicContent{
		Topic: guide.Topic{
			ID:                "discovery",
			Title:             "Discovery Workflows",
			Summary:           "How to inspect before changes.",
			Aliases:           []string{"discover"},
			RelatedTopics:     []string{"index"},
			SuggestedCommands: []string{"hab overview --json"},
		},
		Content: "# Discovery Workflows",
	}

	result := buildGuideTopicOutput(topic)
	if got, _ := result["topic"].(string); got != "discovery" {
		t.Fatalf("topic = %q, want discovery", got)
	}
	if got, _ := result["title"].(string); got != "Discovery Workflows" {
		t.Fatalf("title = %q, want Discovery Workflows", got)
	}
	if got, _ := result["content"].(string); got != "# Discovery Workflows" {
		t.Fatalf("content = %q, want # Discovery Workflows", got)
	}
}

func TestRunGuideTopicUnknown(t *testing.T) {
	err := runGuideTopic("not-a-real-topic")
	if err == nil {
		t.Fatal("runGuideTopic returned nil error for unknown topic")
	}
	if !strings.Contains(err.Error(), "Available topics") {
		t.Fatalf("runGuideTopic error = %q, expected available topics hint", err.Error())
	}
}
