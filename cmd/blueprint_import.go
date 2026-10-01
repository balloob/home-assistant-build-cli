package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var blueprintImportOverride bool

var blueprintImportCmd = &cobra.Command{
	Use:   "import <url>",
	Short: "Import a blueprint from URL",
	Long: `Import a blueprint from a URL and save it to Home Assistant.

The blueprint is saved under the filename that Home Assistant suggests for the URL.
The command fails if the blueprint has validation errors, or if a blueprint with
that filename already exists and --override is not set.`,
	Args: cobra.ExactArgs(1),
	RunE: runBlueprintImport,
}

func init() {
	blueprintCmd.AddCommand(blueprintImportCmd)
	blueprintImportCmd.Flags().BoolVar(&blueprintImportOverride, "override", false, "Replace an existing blueprint with the same filename")
}

func runBlueprintImport(cmd *cobra.Command, args []string) error {
	url := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := importBlueprint(ws, url, blueprintImportOverride)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Blueprint imported from %s and saved as %s/%s", url, result["domain"], result["path"]))
	return nil
}

// importBlueprint fetches a blueprint with blueprint/import and stores it
// with blueprint/save, like the frontend import dialog.
func importBlueprint(ws client.WebSocketCommander, url string, override bool) (map[string]interface{}, error) {
	raw, err := ws.SendCommand("blueprint/import", map[string]interface{}{
		"url": url,
	})
	if err != nil {
		return nil, err
	}
	imported, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected blueprint/import response: %v", raw)
	}

	if errs, ok := imported["validation_errors"].([]interface{}); ok && len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = fmt.Sprint(e)
		}
		return nil, fmt.Errorf("blueprint has validation errors: %s", strings.Join(msgs, "; "))
	}

	filename, _ := imported["suggested_filename"].(string)
	yaml, _ := imported["raw_data"].(string)
	var domain string
	if bp, ok := imported["blueprint"].(map[string]interface{}); ok {
		if md, ok := bp["metadata"].(map[string]interface{}); ok {
			domain, _ = md["domain"].(string)
		}
	}
	if filename == "" || yaml == "" || domain == "" {
		return nil, fmt.Errorf("blueprint/import response is missing suggested_filename, raw_data or blueprint domain")
	}
	path := filename + ".yaml"

	exists, _ := imported["exists"].(bool)
	if exists && !override {
		return nil, fmt.Errorf("blueprint %s/%s already exists; use --override to replace it", domain, path)
	}

	saved, err := ws.SendCommand("blueprint/save", map[string]interface{}{
		"domain":         domain,
		"path":           path,
		"yaml":           yaml,
		"source_url":     url,
		"allow_override": override,
	})
	if err != nil {
		return nil, err
	}

	overrides := false
	if m, ok := saved.(map[string]interface{}); ok {
		overrides, _ = m["overrides_existing"].(bool)
	}
	return map[string]interface{}{
		"domain":             domain,
		"path":               path,
		"source_url":         url,
		"overrides_existing": overrides,
	}, nil
}
