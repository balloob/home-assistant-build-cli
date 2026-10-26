// Package shortid resolves unambiguous prefixes of opaque Home Assistant IDs.
package shortid

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

const MinLength = 8

var (
	hexID      = regexp.MustCompile(`^[0-9a-fA-F]{16,}$`)
	uuidID     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}($|@)`)
	ulidID     = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}$`)
	hexPrefix  = regexp.MustCompile(`^[0-9a-fA-F-]+$`)
	ulidPrefix = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]+$`)
)

// Opaque excludes human-readable names, entity IDs, and already-short IDs.
func Opaque(id string) bool {
	if _, suffix, ok := strings.Cut(id, "."); ok {
		return Opaque(suffix)
	}
	return hexID.MatchString(id) || uuidID.MatchString(id) || ulidID.MatchString(id)
}

// Possible avoids discovery for readable IDs. Full opaque IDs still need an
// exact-match check, since a prefix can itself have a valid full-ID length.
func Possible(id string) bool {
	if _, suffix, ok := strings.Cut(id, "."); ok {
		return Possible(suffix)
	}
	return len(id) >= MinLength && (Opaque(id) || hexPrefix.MatchString(id) || ulidPrefix.MatchString(id))
}

type Candidate struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type Error struct {
	Input      string      `json:"input"`
	Candidates []Candidate `json:"candidates"`
}

func (e *Error) Error() string {
	if len(e.Candidates) == 0 {
		return fmt.Sprintf("no ID matches %q; copy an ID from discovery output", e.Input)
	}
	var matches []string
	for _, c := range e.Candidates {
		matches = append(matches, fmt.Sprintf("%s (%s)", c.ID, c.Name))
	}
	return fmt.Sprintf("ambiguous ID prefix %q: %s; use a longer prefix or the full ID", e.Input, strings.Join(matches, ", "))
}

// Resolve always prefers an exact match, even when that ID is readable/short.
func Resolve(input string, candidates []Candidate) (string, error) {
	for _, c := range candidates {
		if c.ID == input {
			return input, nil
		}
	}
	var matches []Candidate
	seen := map[string]bool{}
	if len(input) >= MinLength {
		for _, c := range candidates {
			if Opaque(c.ID) && strings.HasPrefix(c.ID, input) && !seen[c.ID] {
				matches = append(matches, c)
				seen[c.ID] = true
			}
		}
	}
	slices.SortFunc(matches, func(a, b Candidate) int { return strings.Compare(a.ID, b.ID) })
	if len(matches) == 1 {
		return matches[0].ID, nil
	}
	return "", &Error{Input: input, Candidates: matches}
}

// Abbreviations uses the entire namespace, including short exact IDs, so an
// abbreviation never resolves to another resource. IDs themselves are retained.
func Abbreviations(candidates []Candidate) map[string]string {
	out := map[string]string{}
	for _, c := range candidates {
		if !Opaque(c.ID) {
			continue
		}
		for n := MinLength; n <= len(c.ID); n++ {
			prefix := c.ID[:n]
			if !Possible(prefix) {
				continue
			}
			resolved, err := Resolve(prefix, candidates)
			if err == nil && resolved == c.ID {
				out[c.ID] = prefix
				break
			}
		}
	}
	return out
}

func Candidates(items []any, key string) []Candidate {
	out := []Candidate{}
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id, ok := m[key].(string)
		if !ok || id == "" {
			continue
		}
		name := ""
		for _, field := range []string{"name_by_user", "name", "title", "summary", "message", "network_name", "entity_id"} {
			if value, ok := m[field].(string); ok && value != "" {
				name = value
				break
			}
		}
		out = append(out, Candidate{ID: id, Name: name})
	}
	return out
}

// Annotate runs before filtering/limiting discovery output. It copies entries
// rather than adding display-only fields to shared API response maps.
func Annotate(items []any, key string) []any {
	abbreviations := Abbreviations(Candidates(items, key))
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
		if entry, ok := item.(map[string]any); ok {
			id, _ := entry[key].(string)
			if short, ok := abbreviations[id]; ok {
				copy := maps.Clone(entry)
				copy["short_id"] = short
				result[i] = copy
			}
		}
	}
	return result
}
