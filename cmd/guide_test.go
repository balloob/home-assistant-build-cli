package cmd

import (
	"strings"
	"testing"

	"github.com/home-assistant/hab/guide"
)

func TestBuildGuideTopicOutput(t *testing.T) {
	topic := guide.TopicContent{
		Topic: guide.Topic{
			ID:                   "discovery",
			Title:                "Discovery Workflows",
			Summary:              "How to inspect before changes.",
			Aliases:              []string{"discover"},
			RelatedTopics:        []string{"index"},
			SuggestedCommands:    []string{"hab overview --json"},
			Prerequisites:        []string{"Authenticated session"},
			DiscoverySteps:       []string{"Run overview"},
			MutationPatterns:     []string{"Mutate one command at a time"},
			VerificationCommands: []string{"hab entity get light.kitchen --json"},
			Pitfalls:             []string{"Guessing entity IDs"},
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
	if got, _ := result["prerequisites"].([]string); len(got) != 1 {
		t.Fatalf("prerequisites length = %d, want 1", len(got))
	}
}

func TestBuildGuideListOutput(t *testing.T) {
	result := buildGuideListOutput([]guide.Topic{{
		ID:                "index",
		Title:             "hab Guide",
		Summary:           "Start here.",
		Aliases:           []string{"start"},
		RelatedTopics:     []string{"discovery"},
		SuggestedCommands: []string{"hab guide discovery"},
	}})

	if len(result) != 1 {
		t.Fatalf("list output length = %d, want 1", len(result))
	}
	if got, _ := result[0]["id"].(string); got != "index" {
		t.Fatalf("id = %q, want index", got)
	}
	if _, ok := result[0]["prerequisites"]; ok {
		t.Fatal("list output unexpectedly includes structured topic fields")
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
