package cmd

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/home-assistant/hab/client"
)

// dependentIntegrations lists loaded integrations whose manifest requires domain.
// after_dependencies only orders setup when both are present, so it is ignored.
func dependentIntegrations(ws client.WebSocketAPI, domain string) ([]string, error) {
	result, err := ws.SendCommand("manifest/list", nil)
	if err != nil {
		return nil, err
	}
	manifests, _ := result.([]any)
	dependents := []string{}
	for _, item := range manifests {
		manifest, _ := item.(map[string]any)
		name, _ := manifest["domain"].(string)
		deps, _ := manifest["dependencies"].([]any)
		if name != "" && name != domain && slices.Contains(deps, any(domain)) {
			dependents = append(dependents, name)
		}
	}
	slices.Sort(dependents)
	return dependents, nil
}

// integrationDeleteImpact gathers what removing a config entry affects using
// read-only calls. Previews stay usable offline: any failure is reported as
// not_determined instead of failing the plan.
func integrationDeleteImpact(args []string, timeout time.Duration) (map[string]any, []string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	type outcome struct {
		impact map[string]any
		err    error
	}
	// The reads take no context, so stop waiting at the deadline; the command exits after printing.
	done := make(chan outcome, 1)
	go func() {
		impact, err := readIntegrationDeleteImpact(ctx, args[0], timeout)
		done <- outcome{impact, err}
	}()
	var impact map[string]any
	var err error
	select {
	case result := <-done:
		impact, err = result.impact, result.err
	case <-ctx.Done():
		err = ctx.Err()
	}
	if err != nil {
		_, message, _ := classifyError(err)
		return map[string]any{"status": "not_determined", "reason": message}, nil
	}
	dependents, _ := impact["dependent_integrations"].([]string)
	if len(dependents) == 0 {
		return impact, nil
	}
	return impact, []string{fmt.Sprintf("This is the last enabled %s entry; loaded integrations that require %s will break: %s", impact["domain"], impact["domain"], strings.Join(dependents, ", "))}
}

func readIntegrationDeleteImpact(ctx context.Context, raw string, timeout time.Duration) (map[string]any, error) {
	ws, err := getWSClientContext(ctx)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	defer func() {
		if err := ws.Close(); err != nil {
			noteWarning("Impact connection close failed.")
		}
	}()
	ws.Timeout = timeout
	// Live short-ID lookup is skipped for previews; resolve on this connection instead.
	entryID, err := resolveShortSelector(idNamespace{Kind: "config_entry"}, raw, (&idCatalog{ws: ws}).list)
	if err != nil {
		return nil, err
	}
	entry, err := ws.ConfigEntryGet(entryID)
	if err != nil {
		return nil, err
	}
	domain, _ := entry["domain"].(string)
	siblings, err := ws.ConfigEntriesList(domain)
	if err != nil {
		return nil, err
	}
	// Ignored and disabled entries do not provide the domain to dependents.
	active := func(e map[string]any) bool { return e["source"] != "ignore" && e["disabled_by"] == nil }
	others := 0
	for _, item := range siblings {
		if sibling, _ := item.(map[string]any); sibling["entry_id"] != entryID && active(sibling) {
			others++
		}
	}
	dependents := []string{}
	if others == 0 && active(entry) {
		if dependents, err = dependentIntegrations(ws, domain); err != nil {
			return nil, err
		}
	}
	related, err := ws.SearchRelated("config_entry", entryID)
	if err != nil {
		return nil, err
	}
	devices, entities := related["device"], related["entity"]
	if devices == nil {
		devices = []string{}
	}
	if entities == nil {
		entities = []string{}
	}
	return map[string]any{
		"status":                 "determined",
		"entry_id":               entryID,
		"domain":                 domain,
		"other_domain_entries":   others,
		"dependent_integrations": dependents,
		"devices":                devices,
		"entities":               entities,
	}, nil
}
