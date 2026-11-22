package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
)

func newAppConfigureCommand() *cobra.Command {
	f := &appMutationFlags{}
	var in InputFlags
	cmd := &cobra.Command{
		Use:   "configure <slug>",
		Short: "Set app options and runtime settings",
		Args:  appSlugArgs,
		Long: `Set app configuration using JSON or YAML from --data, --file, or stdin.

Supported fields: options, boot, auto_update, watchdog, ingress_panel,
network, audio_input, and audio_output. Omitted fields stay unchanged.
The options and network objects replace their current values. Read the current
configuration with 'hab app get' before changing them. Restart the app if needed.`,
		Example: `  hab app configure core_mosquitto --data '{"boot":"auto","watchdog":true}'
  hab app configure core_mosquitto --file options.yaml --plan`,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := input.ParseInputExact(in.Data, in.File, in.Format)
			if err != nil {
				return fmt.Errorf("invalid app configuration; provide a JSON or YAML object with --data, --file, or stdin")
			}
			if err := validateAppConfiguration(data); err != nil {
				return err
			}
			return runAppMutation(cmd, args[0], "configure", "options", data, f)
		},
	}
	in.Register(cmd)
	f.register(cmd, false)
	return cmd
}

func validateAppConfiguration(data map[string]any) error {
	if len(data) == 0 {
		return fmt.Errorf("app configuration must contain at least one setting")
	}
	for key, value := range data {
		valid := false
		switch key {
		case "options", "network":
			_, valid = value.(map[string]any)
			valid = valid || value == nil
		case "boot":
			valid = value == "auto" || value == "manual"
		case "auto_update", "watchdog", "ingress_panel":
			_, valid = value.(bool)
		case "audio_input", "audio_output":
			_, valid = value.(string)
			valid = valid || value == nil
		default:
			return fmt.Errorf("unsupported app setting; see 'hab app configure --help'")
		}
		if !valid {
			return fmt.Errorf("invalid value type for app setting %s", key)
		}
	}
	return nil
}
