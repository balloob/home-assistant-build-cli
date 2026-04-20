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
				path, _, _ := commandPathFromTemplate(template)
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

func TestGuideRecipeTemplatesUseKnownFlags(t *testing.T) {
	for _, topic := range guide.List() {
		for _, recipe := range topic.Recipes {
			for _, step := range recipe.Steps {
				if step.CommandTemplate == "" {
					continue
				}

				path, consumed, tokens := commandPathFromTemplate(step.CommandTemplate)
				if step.CommandPath != "" {
					pathFromPath, _, _ := commandPathFromTemplate(step.CommandPath)
					if len(pathFromPath) == 0 {
						t.Fatalf("topic %q recipe %q step %q has unparseable command_path %q", topic.ID, recipe.ID, step.ID, step.CommandPath)
					}
					path = pathFromPath
					consumed = 0
				}
				if len(path) == 0 {
					t.Fatalf("topic %q recipe %q step %q has unparseable command template %q", topic.ID, recipe.ID, step.ID, step.CommandTemplate)
				}

				cmd, _, err := rootCmd.Find(path)
				if err != nil {
					t.Fatalf("topic %q recipe %q step %q references unknown command %q: %v", topic.ID, recipe.ID, step.ID, strings.Join(path, " "), err)
				}

				longFlags, shortFlags := extractTemplateFlags(tokens[consumed:])
				for _, longFlag := range longFlags {
					if cmd.Flags().Lookup(longFlag) == nil && cmd.InheritedFlags().Lookup(longFlag) == nil {
						t.Fatalf("topic %q recipe %q step %q uses unknown flag --%s for command %q", topic.ID, recipe.ID, step.ID, longFlag, strings.Join(path, " "))
					}
				}
				for _, shortFlag := range shortFlags {
					if cmd.Flags().ShorthandLookup(shortFlag) == nil && cmd.InheritedFlags().ShorthandLookup(shortFlag) == nil {
						t.Fatalf("topic %q recipe %q step %q uses unknown shorthand -%s for command %q", topic.ID, recipe.ID, step.ID, shortFlag, strings.Join(path, " "))
					}
				}

			}
		}
	}
}

func commandPathFromTemplate(template string) (path []string, consumed int, tokens []string) {
	tokens = strings.Fields(template)
	if len(tokens) == 0 {
		return nil, 0, nil
	}
	if tokens[0] == "hab" {
		tokens = tokens[1:]
	}
	current := rootCmd
	path = make([]string, 0)
	for _, token := range tokens {
		if strings.HasPrefix(token, "-") || strings.Contains(token, "{{") || strings.HasPrefix(token, "'") || strings.HasPrefix(token, "\"") || strings.Contains(token, "{") {
			break
		}
		matched := false
		for _, child := range current.Commands() {
			if child.Name() == token || slicesContains(child.Aliases, token) {
				path = append(path, token)
				current = child
				consumed++
				matched = true
				break
			}
		}
		if !matched {
			break
		}
	}
	return path, consumed, tokens
}

func extractTemplateFlags(tokens []string) (longFlags []string, shortFlags []string) {
	longSeen := map[string]bool{}
	shortSeen := map[string]bool{}

	for i := range tokens {
		token := tokens[i]
		if token == "--" {
			break
		}

		if strings.HasPrefix(token, "--") {
			name := strings.TrimPrefix(token, "--")
			if idx := strings.Index(name, "="); idx >= 0 {
				name = name[:idx]
			}
			if name != "" && !longSeen[name] {
				longSeen[name] = true
				longFlags = append(longFlags, name)
			}
			continue
		}

		if strings.HasPrefix(token, "-") && len(token) > 1 && !strings.HasPrefix(token, "--") {
			if isNegativeNumberToken(token) {
				continue
			}
			group := token[1:]
			if idx := strings.Index(group, "="); idx >= 0 {
				group = group[:idx]
			}
			for _, r := range group {
				if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
					continue
				}
				name := string(r)
				if shortSeen[name] {
					continue
				}
				shortSeen[name] = true
				shortFlags = append(shortFlags, name)
			}
		}
	}

	return longFlags, shortFlags
}

func isNegativeNumberToken(token string) bool {
	if len(token) <= 1 || token[0] != '-' {
		return false
	}
	for _, r := range token[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func slicesContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
