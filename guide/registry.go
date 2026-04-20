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
	ID                   string   `json:"id"`
	Title                string   `json:"title"`
	Summary              string   `json:"summary"`
	Aliases              []string `json:"aliases,omitempty"`
	RelatedTopics        []string `json:"related_topics,omitempty"`
	SuggestedCommands    []string `json:"suggested_commands,omitempty"`
	Prerequisites        []string `json:"prerequisites,omitempty"`
	DiscoverySteps       []string `json:"discovery_steps,omitempty"`
	MutationPatterns     []string `json:"mutation_patterns,omitempty"`
	VerificationCommands []string `json:"verification_commands,omitempty"`
	Pitfalls             []string `json:"pitfalls,omitempty"`
	Recipes              []Recipe `json:"recipes,omitempty"`
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
				"auth",
				"input-output",
				"discovery",
				"registry",
				"automation",
				"dashboard",
				"helpers",
				"calendar-todo",
				"esphome",
				"operations",
			},
			SuggestedCommands: []string{
				"hab guide auth",
				"hab guide input-output",
				"hab guide discovery",
				"hab overview --json",
				"hab entity search kitchen --json",
			},
			Prerequisites: []string{
				"Authenticate before mutating resources with hab auth login.",
				"Use --json for machine parsing and deterministic automation flows.",
			},
			DiscoverySteps: []string{
				"Start with hab guide list to discover workflow topics.",
				"Use hab overview --json and targeted list/search commands before edits.",
			},
			MutationPatterns: []string{
				"Apply one mutation command at a time.",
				"Prefer explicit IDs from discovery output over guessed names.",
			},
			VerificationCommands: []string{
				"hab overview --json",
				"hab entity get <entity_id> --json",
			},
			Pitfalls: []string{
				"Skipping discovery and mutating with guessed IDs.",
				"Mixing positional and flag styles inconsistently in one workflow.",
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
				"auth",
				"registry",
				"automation",
			},
			SuggestedCommands: []string{
				"hab overview --json",
				"hab entity search temperature --json",
				"hab search related entity light.kitchen --json",
				"hab action docs light.turn_on --json",
			},
			Prerequisites: []string{
				"Authentication is optional for read-only discovery but required for follow-up writes.",
				"Use JSON mode to keep outputs machine-parseable.",
			},
			DiscoverySteps: []string{
				"Start with hab overview --json for global context.",
				"Map entities, devices, and areas with list and related commands.",
				"Inspect action schema before planning any action call.",
			},
			MutationPatterns: []string{
				"Do not mutate during discovery unless explicitly requested.",
				"Capture IDs from output and reuse them exactly.",
			},
			VerificationCommands: []string{
				"hab entity get <entity_id> --json",
				"hab device entities <device_id> --json",
			},
			Pitfalls: []string{
				"Choosing entities by name similarity without checking relationships.",
				"Calling actions without inspecting required data fields.",
			},
		},
		Filename: "discovery",
	},
	{
		Topic: Topic{
			ID:      "auth",
			Title:   "Authentication Workflows",
			Summary: "Log in, validate session state, and recover credentials before running workflows.",
			Aliases: []string{"authentication", "login"},
			RelatedTopics: []string{
				"index",
				"discovery",
				"input-output",
			},
			SuggestedCommands: []string{
				"hab auth status --json",
				"hab auth login",
				"hab auth refresh --json",
			},
			Prerequisites: []string{
				"A reachable Home Assistant URL.",
				"An access token or OAuth browser flow for login.",
			},
			DiscoverySteps: []string{
				"Use hab auth status --json to detect whether credentials already exist.",
				"Use hab auth discover to find local Home Assistant instances when URL is unknown.",
			},
			MutationPatterns: []string{
				"Authenticate once, then reuse the session across all workflows.",
				"Use hab auth refresh when a long-running automation hits token expiry.",
			},
			VerificationCommands: []string{
				"hab auth status --json",
				"hab overview --json",
			},
			Pitfalls: []string{
				"Attempting write operations before confirming auth status.",
				"Mixing credentials from different Home Assistant instances.",
			},
		},
		Filename: "auth",
	},
	{
		Topic: Topic{
			ID:      "input-output",
			Title:   "Input and Output Patterns",
			Summary: "Use JSON output and safe data input patterns for deterministic agent workflows.",
			Aliases: []string{"input", "io", "data-entry"},
			RelatedTopics: []string{
				"index",
				"auth",
				"automation",
				"dashboard",
			},
			SuggestedCommands: []string{
				"hab overview --json",
				"hab action docs light.turn_on --json",
				"hab automation create my_automation -f automation.yaml",
			},
			Prerequisites: []string{
				"Choose JSON mode when output will be parsed by automation.",
				"Prefer files or heredocs for complex payloads.",
			},
			DiscoverySteps: []string{
				"Inspect command-specific help for supported input flags.",
				"Use docs/list commands to inspect required fields before writing payloads.",
			},
			MutationPatterns: []string{
				"Use -d for small inline payloads and -f for larger schemas.",
				"Keep one payload source per command invocation for predictability.",
			},
			VerificationCommands: []string{
				"hab automation list --json",
				"hab dashboard get <dashboard_id> --json",
			},
			Pitfalls: []string{
				"Mixing YAML and JSON syntax in one payload.",
				"Parsing text-mode output in automation instead of JSON envelopes.",
			},
		},
		Filename: "input-output",
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
				"operations",
			},
			SuggestedCommands: []string{
				"hab area list --json",
				"hab device list --json",
				"hab entity list --domain light --json",
			},
			Prerequisites: []string{
				"Use discovery commands to collect valid IDs first.",
				"Understand that entity IDs and registry IDs are different identifiers.",
			},
			DiscoverySteps: []string{
				"List resources for the target scope (area, floor, device, label, person, zone).",
				"Use related/entity commands to verify cross-resource links.",
			},
			MutationPatterns: []string{
				"Create resources with explicit names and verify generated IDs.",
				"Use update/delete only after a get/list confirmation step.",
			},
			VerificationCommands: []string{
				"hab area get <area_id> --json",
				"hab entity get <entity_id> --json",
			},
			Pitfalls: []string{
				"Passing an entity ID where a registry device ID is required.",
				"Deleting by name without confirming unique IDs.",
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
				"input-output",
			},
			SuggestedCommands: []string{
				"hab action list light --json",
				"hab automation list --json",
				"hab script list --json",
			},
			Prerequisites: []string{
				"Resolve target entities and expected action schemas before mutations.",
				"Use JSON output for list/get workflows that feed later commands.",
			},
			DiscoverySteps: []string{
				"Inspect action list/docs to understand available services and fields.",
				"List existing automations, scripts, and scenes before create/update.",
			},
			MutationPatterns: []string{
				"Test an action call directly before embedding it in automations or scripts.",
				"Use category scopes explicitly when assigning categories.",
			},
			VerificationCommands: []string{
				"hab automation get <automation_id> --json",
				"hab script list --json",
			},
			Pitfalls: []string{
				"Calling actions with undocumented fields.",
				"Confusing category scope with entity domain prefixes.",
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
				"input-output",
			},
			SuggestedCommands: []string{
				"hab dashboard list --json",
				"hab dashboard view list my-dashboard --json",
				"hab dashboard card list my-dashboard 0 --json",
			},
			Prerequisites: []string{
				"Discover relevant entities/devices before composing views.",
				"Decide whether to create resources incrementally or from YAML payloads.",
			},
			DiscoverySteps: []string{
				"List dashboards and existing views before edits.",
				"Inspect current card/section structure before patching.",
			},
			MutationPatterns: []string{
				"Use view/section/card subcommands for focused edits.",
				"Use YAML input with files or heredocs when creating complete views.",
			},
			VerificationCommands: []string{
				"hab dashboard get <dashboard_id> --json",
				"hab dashboard card list <dashboard_id> <view_index> --json",
			},
			Pitfalls: []string{
				"Building dashboards from entity names only and missing device context.",
				"Editing sections/cards without first confirming view indexes.",
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
				"input-output",
			},
			SuggestedCommands: []string{
				"hab helper types --json",
				"hab helper input-boolean list --json",
				"hab helper statistics create \"Temp Average\" --entity sensor.temperature --characteristic mean",
			},
			Prerequisites: []string{
				"Identify the helper subtype command before creation.",
				"Inspect required flags with --help for subtype create commands.",
			},
			DiscoverySteps: []string{
				"Use hab helper types --json to enumerate available helper families.",
				"Inspect existing helpers before creating duplicates.",
			},
			MutationPatterns: []string{
				"Create helpers with explicit names/icons where available.",
				"Use helper delete by entity ID and verify removal immediately.",
			},
			VerificationCommands: []string{
				"hab helper list --json",
				"hab entity get <entity_id> --json",
			},
			Pitfalls: []string{
				"Choosing the wrong helper subtype for the intended behavior.",
				"Assuming helper IDs from names without confirming generated entity IDs.",
			},
		},
		Filename: "helpers",
	},
	{
		Topic: Topic{
			ID:      "calendar-todo",
			Title:   "Calendar and To-Do Workflows",
			Summary: "Plan event and task workflows with calendar and to-do list commands.",
			Aliases: []string{"calendar", "todo", "calendar-tasks"},
			RelatedTopics: []string{
				"index",
				"discovery",
				"input-output",
			},
			SuggestedCommands: []string{
				"hab calendar list calendar.personal --json",
				"hab todo lists --json",
				"hab todo items todo.shopping_list --json",
			},
			Prerequisites: []string{
				"Know the calendar entity_id or to-do list entity_id target.",
				"Use ISO date/time values for predictable parsing.",
			},
			DiscoverySteps: []string{
				"List available calendar events in a bounded time range.",
				"List to-do entities, then inspect items before mutating.",
			},
			MutationPatterns: []string{
				"Create/update one task or event at a time and store returned identifiers.",
				"Use complete/uncomplete transitions instead of deleting when history matters.",
			},
			VerificationCommands: []string{
				"hab calendar list <calendar_entity_id> --json",
				"hab todo items <todo_entity_id> --json",
			},
			Pitfalls: []string{
				"Using a list entity ID with calendar commands or vice versa.",
				"Mixing all-day and timed event formats in one payload.",
			},
		},
		Filename: "calendar-todo",
	},
	{
		Topic: Topic{
			ID:      "esphome",
			Title:   "ESPHome Workflows",
			Summary: "Discover, validate, build, and deploy ESPHome configurations safely.",
			Aliases: []string{"esp", "firmware"},
			RelatedTopics: []string{
				"index",
				"auth",
				"operations",
			},
			SuggestedCommands: []string{
				"hab esphome list --json",
				"hab esphome context use <configuration>",
				"hab esphome validate <configuration>",
			},
			Prerequisites: []string{
				"A reachable ESPHome dashboard via HAB_ESPHOME_URL or ingress discovery.",
				"A selected configuration argument or saved context.",
			},
			DiscoverySteps: []string{
				"List configurations and inspect metadata before builds.",
				"Read configuration YAML prior to patching or migration.",
			},
			MutationPatterns: []string{
				"Validate before build/upload or run commands.",
				"Prefer config-patch workflows for targeted YAML changes.",
			},
			VerificationCommands: []string{
				"hab esphome info <configuration> --json",
				"hab esphome validate <configuration>",
			},
			Pitfalls: []string{
				"Uploading firmware before validating config and target device.",
				"Running commands without setting --device or context.",
			},
		},
		Filename: "esphome",
	},
	{
		Topic: Topic{
			ID:      "operations",
			Title:   "Operations and Maintenance",
			Summary: "Operate Home Assistant safely across backups, health, diagnostics, and repairs.",
			Aliases: []string{"ops", "maintenance"},
			RelatedTopics: []string{
				"index",
				"auth",
				"registry",
				"esphome",
			},
			SuggestedCommands: []string{
				"hab system health --json",
				"hab backup list --json",
				"hab repairs list --json",
			},
			Prerequisites: []string{
				"Authenticated access to admin-level APIs.",
				"A clear rollback plan before applying disruptive operations.",
			},
			DiscoverySteps: []string{
				"Check system health and pending updates before running maintenance commands.",
				"Inspect repairs and diagnostics to scope issues before changes.",
			},
			MutationPatterns: []string{
				"Create backups before restarts, network changes, or repair actions.",
				"Apply one operational change at a time and re-check health.",
			},
			VerificationCommands: []string{
				"hab system health --json",
				"hab diagnostics list --json",
			},
			Pitfalls: []string{
				"Restarting or restoring without a recent backup.",
				"Ignoring repair issues without understanding downstream effects.",
			},
		},
		Filename: "operations",
	},
}

// List returns all guide topics in deterministic display order.
func List() []Topic {
	result := make([]Topic, len(topicRegistry))
	for i, topic := range topicRegistry {
		result[i] = enrichTopic(topic.Topic)
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
			return enrichTopic(topic.Topic), nil
		}
		for _, alias := range topic.Aliases {
			if normalizeTopicName(alias) == normalized {
				return enrichTopic(topic.Topic), nil
			}
		}
	}

	return Topic{}, fmt.Errorf("%w: %s", ErrTopicNotFound, name)
}

func enrichTopic(topic Topic) Topic {
	topic.Recipes = recipesForTopic(topic.ID)
	return topic
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
