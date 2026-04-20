package cmd

import (
	"strings"
	"testing"

	"github.com/home-assistant/hab/guide"
)

func TestGuideRecipesReferenceRealCommands(t *testing.T) {
	for _, topic := range guide.List() {
		for _, recipe := range topic.Recipes {
			for _, step := range recipe.Steps {
				template := step.CommandTemplate
				if step.CommandPath != "" {
					template = step.CommandPath
				}
				if template == "" {
					continue
				}
				path := commandPathFromTemplate(template)
				if len(path) == 0 {
					t.Fatalf("topic %q recipe %q step %q has unparseable command template %q", topic.ID, recipe.ID, step.ID, template)
				}
				if _, _, err := rootCmd.Find(path); err != nil {
					t.Fatalf("topic %q recipe %q step %q references unknown command %q: %v", topic.ID, recipe.ID, step.ID, strings.Join(path, " "), err)
				}
			}
		}
	}
}

func commandPathFromTemplate(template string) []string {
	tokens := strings.Fields(template)
	if len(tokens) == 0 {
		return nil
	}
	if tokens[0] == "hab" {
		tokens = tokens[1:]
	}
	current := rootCmd
	path := make([]string, 0)
	for _, token := range tokens {
		if strings.HasPrefix(token, "-") || strings.Contains(token, "{{") || strings.HasPrefix(token, "'") || strings.HasPrefix(token, "\"") || strings.Contains(token, "{") {
			break
		}
		matched := false
		for _, child := range current.Commands() {
			if child.Name() == token || slicesContains(child.Aliases, token) {
				path = append(path, token)
				current = child
				matched = true
				break
			}
		}
		if !matched {
			break
		}
	}
	return path
}

func slicesContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
