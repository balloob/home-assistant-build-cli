package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

func newAppListCommand() *cobra.Command {
	var available bool
	var search string
	var flags *ListFlags
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed apps",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := "/addons"
			if available {
				endpoint = "/store"
			}
			result, err := appRequest(cmd, "get", endpoint, nil, 30*time.Second)
			if err != nil {
				return err
			}
			apps, err := filterApps(result, search)
			if err != nil {
				return err
			}
			if flags.RenderCount(len(apps), getTextMode()) {
				return nil
			}
			apps = flags.ApplyLimit(apps)
			if flags.RenderBrief(apps, getTextMode(), "slug", "name") {
				return nil
			}
			if getTextMode() {
				for i, item := range apps {
					app := item.(map[string]any)
					apps[i] = map[string]any{"slug": app["slug"], "name": app["name"], "state": app["state"], "version": app["version"], "version_latest": app["version_latest"]}
				}
			}
			output.PrintOutput(apps, getTextMode(), "")
			return nil
		},
	}
	cmd.Flags().BoolVar(&available, "available", false, "Include apps available from the app store")
	cmd.Flags().StringVar(&search, "search", "", "Filter by name, slug, or description")
	flags = RegisterListFlags(cmd, "slug")
	annotateAppCommand(cmd, "read")
	return cmd
}

func filterApps(result any, search string) ([]any, error) {
	response, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid Supervisor app list response")
	}
	apps, ok := response["addons"].([]any)
	if !ok {
		return nil, fmt.Errorf("Supervisor response does not contain an app list")
	}
	filtered := make([]any, 0, len(apps))
	search = strings.ToLower(search)
	for _, item := range apps {
		app, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid app in Supervisor response")
		}
		for _, key := range []string{"slug", "name", "description"} {
			value, _ := app[key].(string)
			if strings.Contains(strings.ToLower(value), search) {
				filtered = append(filtered, app)
				break
			}
		}
	}
	return filtered, nil
}
