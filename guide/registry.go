package guide

import (
	"errors"
	"fmt"
	"strings"
)

// ErrTopicNotFound is returned when a guide topic cannot be resolved.
var ErrTopicNotFound = errors.New("guide topic not found")

// Topic describes a guide topic and its metadata.
type Topic struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Summary           string   `json:"summary"`
	Aliases           []string `json:"aliases,omitempty"`
	RelatedTopics     []string `json:"related_topics,omitempty"`
	SuggestedCommands []string `json:"suggested_commands,omitempty"`
}

// TopicContent is a fully loaded guide topic with markdown content.
type TopicContent struct {
	Topic   Topic  `json:"topic"`
	Content string `json:"content"`
}

type topicDef struct {
	Topic
	Filename string
}

var topicRegistry = []topicDef{
	{
		Topic: Topic{
			ID:      "index",
			Title:   "hab Guide",
			Summary: "Start here for LLM-friendly hab usage patterns and command routing.",
			Aliases: []string{"start", "getting-started"},
			RelatedTopics: []string{
				"discovery",
				"registry",
				"automation",
				"dashboard",
				"helpers",
			},
			SuggestedCommands: []string{
				"hab guide discovery",
				"hab overview --json",
				"hab entity search kitchen --json",
			},
		},
		Filename: "index",
	},
	{
		Topic: Topic{
			ID:      "discovery",
			Title:   "Discovery Workflows",
			Summary: "How to inspect a Home Assistant instance before making changes.",
			Aliases: []string{"discover"},
			RelatedTopics: []string{
				"index",
				"registry",
				"automation",
			},
			SuggestedCommands: []string{
				"hab overview --json",
				"hab entity search temperature --json",
				"hab action docs light.turn_on --json",
			},
		},
		Filename: "discovery",
	},
	{
		Topic: Topic{
			ID:      "registry",
			Title:   "Registry Commands",
			Summary: "Use area/floor/device/entity/label/person/zone commands with the right IDs.",
			Aliases: []string{"registry-entities"},
			RelatedTopics: []string{
				"index",
				"discovery",
				"dashboard",
			},
			SuggestedCommands: []string{
				"hab area list --json",
				"hab device list --json",
				"hab entity list --domain light --json",
			},
		},
		Filename: "registry",
	},
	{
		Topic: Topic{
			ID:      "automation",
			Title:   "Automation Commands",
			Summary: "Workflows for actions, automations, scripts, scenes, categories, and templates.",
			Aliases: []string{"automations", "actions"},
			RelatedTopics: []string{
				"index",
				"discovery",
				"helpers",
			},
			SuggestedCommands: []string{
				"hab action list light --json",
				"hab automation list --json",
				"hab script list --json",
			},
		},
		Filename: "automation",
	},
	{
		Topic: Topic{
			ID:      "dashboard",
			Title:   "Dashboard Design Guide",
			Summary: "Best practices for creating maintainable Home Assistant dashboards.",
			Aliases: []string{"dashboards"},
			RelatedTopics: []string{
				"index",
				"discovery",
				"registry",
			},
			SuggestedCommands: []string{
				"hab dashboard list --json",
				"hab dashboard view list my-dashboard --json",
				"hab dashboard card list my-dashboard 0 --json",
			},
		},
		Filename: "dashboard",
	},
	{
		Topic: Topic{
			ID:      "helpers",
			Title:   "Helper Commands",
			Summary: "Create and manage helper entities with consistent list/create/delete flows.",
			Aliases: []string{"helper", "helper-types"},
			RelatedTopics: []string{
				"index",
				"discovery",
				"automation",
			},
			SuggestedCommands: []string{
				"hab helper types --json",
				"hab helper input-boolean list --json",
				"hab helper statistics create \"Temp Average\" --entity sensor.temperature --characteristic mean",
			},
		},
		Filename: "helpers",
	},
}

// List returns all guide topics in deterministic display order.
func List() []Topic {
	result := make([]Topic, len(topicRegistry))
	for i, topic := range topicRegistry {
		result[i] = topic.Topic
	}
	return result
}

// TopicIDs returns all canonical topic IDs in deterministic display order.
func TopicIDs() []string {
	result := make([]string, len(topicRegistry))
	for i, topic := range topicRegistry {
		result[i] = topic.ID
	}
	return result
}

// Resolve resolves a topic name or alias to a canonical topic.
func Resolve(name string) (Topic, error) {
	normalized := normalizeTopicName(name)
	for _, topic := range topicRegistry {
		if topic.ID == normalized {
			return topic.Topic, nil
		}
		for _, alias := range topic.Aliases {
			if normalizeTopicName(alias) == normalized {
				return topic.Topic, nil
			}
		}
	}

	return Topic{}, fmt.Errorf("%w: %s", ErrTopicNotFound, name)
}

// GetTopic loads a topic's metadata and markdown content by topic name or alias.
func GetTopic(name string) (TopicContent, error) {
	topic, err := Resolve(name)
	if err != nil {
		return TopicContent{}, err
	}

	for _, def := range topicRegistry {
		if def.ID != topic.ID {
			continue
		}

		content, readErr := Guides.ReadFile(def.Filename + ".md")
		if readErr != nil {
			return TopicContent{}, fmt.Errorf("failed to read guide topic '%s': %w", topic.ID, readErr)
		}

		return TopicContent{
			Topic:   topic,
			Content: string(content),
		}, nil
	}

	return TopicContent{}, fmt.Errorf("%w: %s", ErrTopicNotFound, name)
}

// Get returns topic markdown content by topic name or alias.
func Get(name string) (string, error) {
	topic, err := GetTopic(name)
	if err != nil {
		return "", err
	}
	return topic.Content, nil
}

func normalizeTopicName(name string) string {
	normalized := strings.TrimSpace(strings.ToLower(name))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, " ", "-")
	return normalized
}
