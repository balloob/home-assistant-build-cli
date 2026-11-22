package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

type transferFlags struct {
	output, device, mediaID, passwordFile string
	agents                                []string
	timeout                               time.Duration
	plan, dryRun                          bool
}

func init() {
	backupCmd.AddCommand(newTransferCommand("backup", false), newTransferCommand("backup", true))
	diagnosticsCmd.AddCommand(newTransferCommand("diagnostics", false))
	mediaCmd.AddCommand(newTransferCommand("media", true))
}

func newTransferCommand(kind string, upload bool) *cobra.Command {
	f := &transferFlags{}
	use, short := "download <id>", "Download to a new file"
	if upload {
		use, short = "upload <file>", "Upload a local file"
	}
	cmd := &cobra.Command{Use: use, Short: short, Args: cobra.ExactArgs(1), Annotations: map[string]string{"hab.local_preview": "true"}}
	cmd.Flags().DurationVar(&f.timeout, "timeout", 10*time.Minute, "Maximum transfer duration")
	if upload {
		registerMutationPlanFlags(cmd, &f.plan, &f.dryRun)
	} else {
		cmd.Flags().StringVarP(&f.output, "output", "o", "", "Destination path (must not exist)")
	}
	if kind == "backup" {
		cmd.Flags().StringArrayVar(&f.agents, "agent", nil, "Backup agent ID; repeat for uploads")
		if !upload {
			cmd.Flags().StringVar(&f.passwordFile, "password-file", "", "Read a decryption password from a private file")
		}
	}
	if kind == "media" {
		cmd.Flags().StringVar(&f.mediaID, "media-content-id", "", "Destination media directory ID from media browse")
	}
	if kind == "diagnostics" {
		cmd.Flags().StringVar(&f.device, "device", "", "Download diagnostics for this device in the config entry")
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return runTransfer(cmd, kind, upload, args[0], f) }
	effect := "read"
	if upload {
		effect = "write"
	}
	mergeSchemaAnnotation(cmd, SchemaAnnotation{SideEffect: effect, OutputMode: "json_envelope", Capabilities: []string{"auth", "rest"}, GuideTopic: "management", InputSources: []string{"args", "flags"}})
	return cmd
}

func prepareTransfer(kind string, upload bool, arg string, f *transferFlags) (string, url.Values, map[string]string, error) {
	query, fields := url.Values{}, map[string]string{}
	if f.timeout <= 0 {
		return "", nil, nil, fmt.Errorf("--timeout must be positive")
	}
	if !upload {
		if f.output == "" {
			return "", nil, nil, fmt.Errorf("--output is required")
		}
		if err := validateAPIPathSegment(arg); err != nil {
			return "", nil, nil, err
		}
	}
	endpoint := kind + "/download/" + url.PathEscape(arg)
	if upload {
		endpoint = kind + "/upload"
	}
	switch kind {
	case "backup":
		if len(f.agents) == 0 || (!upload && len(f.agents) != 1) {
			return "", nil, nil, fmt.Errorf("provide --agent (exactly one for downloads)")
		}
		for _, agent := range f.agents {
			if strings.TrimSpace(agent) == "" {
				return "", nil, nil, fmt.Errorf("backup agent ID cannot be empty")
			}
			query.Add("agent_id", agent)
		}
		if f.passwordFile != "" {
			password, err := os.ReadFile(f.passwordFile)
			if err != nil {
				return "", nil, nil, err
			}
			query.Set("password", strings.TrimRight(string(password), "\r\n"))
		}
	case "media":
		if f.mediaID == "" {
			return "", nil, nil, fmt.Errorf("--media-content-id is required")
		}
		endpoint = "media_source/local_source/upload"
		fields["media_content_id"] = f.mediaID
	case "diagnostics":
		endpoint = "diagnostics/config_entry/" + url.PathEscape(arg)
		if f.device != "" {
			if err := validateAPIPathSegment(f.device); err != nil {
				return "", nil, nil, err
			}
			endpoint += "/device/" + url.PathEscape(f.device)
		}
	}
	return endpoint, query, fields, nil
}

func runTransfer(cmd *cobra.Command, kind string, upload bool, arg string, f *transferFlags) error {
	endpoint, query, fields, err := prepareTransfer(kind, upload, arg, f)
	if err != nil {
		return err
	}
	if upload {
		info, err := os.Stat(arg)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("upload requires a regular file")
		}
		if planRequested(f.plan, f.dryRun) {
			printMutationPlan(MutationPlan{WouldChange: true, Target: map[string]any{"command": schemaPath(cmd)}, Steps: []string{"Upload the selected local file"}}, getTextMode(), kind)
			return nil
		}
	}
	rest, err := getAuthManager().GetRestClient()
	if err != nil {
		return err
	}
	rest.Timeout = f.timeout
	noteTransport("rest")
	signalCtx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, f.timeout)
	defer cancel()
	var result any
	if upload {
		result, err = rest.UploadFile(ctx, endpoint, query, arg, fields)
		if err != nil {
			return managementMutationError(err)
		}
	} else {
		size, downloadErr := rest.DownloadFile(ctx, endpoint, query, f.output)
		if downloadErr != nil {
			return downloadErr
		}
		result = map[string]any{"path": f.output, "bytes": size}
	}
	output.PrintOutput(result, getTextMode(), "")
	return nil
}
