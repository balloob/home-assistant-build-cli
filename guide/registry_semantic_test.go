package guide

import (
	"strings"
	"testing"
)

func TestTopicRegistryFilesMatchEmbeddedGuides(t *testing.T) {
	entries, err := Guides.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir(.) error = %v", err)
	}

	embeddedMarkdown := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		embeddedMarkdown[entry.Name()] = struct{}{}
	}

	registeredMarkdown := make(map[string]string)
	for _, topic := range topicRegistry {
		filename := topic.Filename + ".md"
		if _, seen := registeredMarkdown[filename]; seen {
			t.Fatalf("duplicate topic filename registration: %s", filename)
		}
		if _, readErr := Guides.ReadFile(filename); readErr != nil {
			t.Fatalf("topic %q references missing guide file %q: %v", topic.ID, filename, readErr)
		}
		registeredMarkdown[filename] = topic.ID
	}

	for filename := range embeddedMarkdown {
		if _, ok := registeredMarkdown[filename]; !ok {
			t.Fatalf("embedded guide file %q is not registered in topicRegistry", filename)
		}
	}
}

func TestTopicRegistryHasUniqueIDsAndAliases(t *testing.T) {
	seen := make(map[string]string)

	for _, topic := range topicRegistry {
		normalizedID := normalizeTopicName(topic.ID)
		if owner, ok := seen[normalizedID]; ok {
			t.Fatalf("topic ID collision: %q conflicts with %q", topic.ID, owner)
		}
		seen[normalizedID] = topic.ID

		for _, alias := range topic.Aliases {
			normalizedAlias := normalizeTopicName(alias)
			if owner, ok := seen[normalizedAlias]; ok {
				t.Fatalf("alias collision: %q (topic %q) conflicts with %q", alias, topic.ID, owner)
			}
			seen[normalizedAlias] = topic.ID
		}
	}
}

func TestTopicRegistryStructuredFieldsPresent(t *testing.T) {
	for _, topic := range topicRegistry {
		if len(topic.SuggestedCommands) == 0 {
			t.Fatalf("topic %q missing suggested commands", topic.ID)
		}
		if len(topic.Prerequisites) == 0 {
			t.Fatalf("topic %q missing prerequisites", topic.ID)
		}
		if len(topic.DiscoverySteps) == 0 {
			t.Fatalf("topic %q missing discovery steps", topic.ID)
		}
		if len(topic.MutationPatterns) == 0 {
			t.Fatalf("topic %q missing mutation patterns", topic.ID)
		}
		if len(topic.VerificationCommands) == 0 {
			t.Fatalf("topic %q missing verification commands", topic.ID)
		}
		if len(topic.Pitfalls) == 0 {
			t.Fatalf("topic %q missing pitfalls", topic.ID)
		}
	}
}

func TestTopicRegistrySuggestedCommandsPrefixedWithHab(t *testing.T) {
	for _, topic := range topicRegistry {
		for _, command := range topic.SuggestedCommands {
			if !strings.HasPrefix(command, "hab ") {
				t.Fatalf("topic %q has non-hab suggested command %q", topic.ID, command)
			}
		}
		for _, command := range topic.VerificationCommands {
			if !strings.HasPrefix(command, "hab ") {
				t.Fatalf("topic %q has non-hab verification command %q", topic.ID, command)
			}
		}
	}
}

func TestTopicRegistryRecipesPresentAndStructured(t *testing.T) {
	for _, topicDef := range topicRegistry {
		topic := enrichTopic(topicDef.Topic)
		if len(topic.Recipes) == 0 {
			t.Fatalf("topic %q missing recipes", topic.ID)
		}
		for _, recipe := range topic.Recipes {
			if recipe.ID == "" {
				t.Fatalf("topic %q has recipe with empty id", topic.ID)
			}
			if len(recipe.Steps) == 0 {
				t.Fatalf("topic %q recipe %q missing steps", topic.ID, recipe.ID)
			}
			for _, step := range recipe.Steps {
				if step.ID == "" {
					t.Fatalf("topic %q recipe %q has step with empty id", topic.ID, recipe.ID)
				}
				if step.Summary == "" {
					t.Fatalf("topic %q recipe %q step %q missing summary", topic.ID, recipe.ID, step.ID)
				}
			}
			if len(recipe.RequiredCapabilities) == 0 {
				t.Fatalf("topic %q recipe %q missing required capabilities", topic.ID, recipe.ID)
			}
		}
	}
}

func TestTopicRegistryRecipeCapabilitiesKnown(t *testing.T) {
	allowed := map[string]struct{}{
		"local": {}, "auth": {}, "rest": {}, "ws": {}, "supervisor": {}, "esphome": {},
	}
	for _, topicDef := range topicRegistry {
		topic := enrichTopic(topicDef.Topic)
		for _, recipe := range topic.Recipes {
			for _, capability := range recipe.RequiredCapabilities {
				if _, ok := allowed[capability]; !ok {
					t.Fatalf("topic %q recipe %q uses unknown capability %q", topic.ID, recipe.ID, capability)
				}
			}
		}
	}
}

func TestTopicRegistryRecipeReferencesValidStepIDs(t *testing.T) {
	for _, topicDef := range topicRegistry {
		topic := enrichTopic(topicDef.Topic)
		for _, recipe := range topic.Recipes {
			ids := map[string]struct{}{}
			for _, step := range recipe.Steps {
				ids[step.ID] = struct{}{}
			}
			for _, branch := range recipe.Branches {
				if _, ok := ids[branch.StepID]; !ok {
					t.Fatalf("topic %q recipe %q branch references unknown step %q", topic.ID, recipe.ID, branch.StepID)
				}
			}
			for _, step := range recipe.Steps {
				for _, onFailure := range step.OnFailure {
					if _, ok := ids[onFailure]; !ok {
						t.Fatalf("topic %q recipe %q step %q on_failure references unknown step %q", topic.ID, recipe.ID, step.ID, onFailure)
					}
				}
			}
		}
	}
}
