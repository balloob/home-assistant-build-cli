package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var templateRenderFile string

var templateRenderCmd = &cobra.Command{
	Use:   "render [template]",
	Short: "Render a Jinja2 template",
	Long: `Render a Jinja2 template using the Home Assistant template engine.

The template can be provided as:
  - A positional argument
  - A file via -f/--file
  - Stdin (when no argument or file is given)`,
	Example: `  hab template render "{{ states('sensor.temperature') }}"
  hab template render "It is {{ now().strftime('%H:%M') }}"
  hab template render -f my_template.j2
  echo "{{ states('sun.sun') }}" | hab template render`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTemplateRender,
}

func init() {
	templateCmd.AddCommand(templateRenderCmd)
	templateRenderCmd.Flags().StringVarP(&templateRenderFile, "file", "f", "", "Read template from file")
}

func runTemplateRender(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	var tmpl string

	switch {
	case len(args) > 0:
		// Inline argument
		tmpl = args[0]
	case templateRenderFile != "":
		// File input
		data, err := os.ReadFile(templateRenderFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		tmpl = string(data)
	default:
		// Stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		tmpl = string(data)
	}

	if tmpl == "" {
		return fmt.Errorf("template is empty — provide a template argument, --file, or pipe via stdin")
	}

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	rendered, err := restClient.RenderTemplate(tmpl)
	if err != nil {
		return err
	}

	if textMode {
		fmt.Println(rendered)
	} else {
		output.PrintOutput(map[string]interface{}{"result": rendered}, false, "")
	}
	return nil
}
