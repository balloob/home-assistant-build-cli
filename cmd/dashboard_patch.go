package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/home-assistant/hab/internal/dashboardpatch"
	"github.com/home-assistant/hab/internal/redact"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

type dashboardPatchFlags struct {
	input           InputFlags
	target, ifMatch string
	set, remove     []string
	plan, dryRun    bool
	diffLimit       int
	timeout         time.Duration
}

func init() { dashboardCmd.AddCommand(newDashboardPatchCommand()) }

func newDashboardPatchCommand() *cobra.Command {
	f := &dashboardPatchFlags{}
	cmd := &cobra.Command{
		Use: "patch <url_path>", Short: "Patch dashboard fields with a diff, freshness check, and read-back verification",
		Long:    "Deep-merge --data/--file into an exact JSON-pointer --target. Arrays are replaced explicitly. --set and --remove use JSON pointers relative to the target. Preview with --plan, then apply with its --if-match base_revision. Verification checks stored JSON, not rendering.",
		GroupID: dashboardGroupCommands, Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { return runDashboardPatch(cmd, args[0], f) },
	}
	f.input.Register(cmd)
	cmd.Flags().StringVar(&f.target, "target", "", "Exact JSON pointer to an object (empty selects root)")
	cmd.Flags().StringVar(&f.ifMatch, "if-match", "", "Required on apply: base_revision from a previous preview")
	cmd.Flags().StringArrayVar(&f.set, "set", nil, "Set /field=JSON relative to target (repeatable; strings must be JSON quoted)")
	cmd.Flags().StringArrayVar(&f.remove, "remove", nil, "Remove object field by relative JSON pointer (repeatable)")
	cmd.Flags().IntVar(&f.diffLimit, "diff-limit", 50, "Maximum returned differences (1-500)")
	cmd.Flags().DurationVar(&f.timeout, "timeout", 30*time.Second, "Deadline for connection and dashboard requests")
	registerMutationPlanFlags(cmd, &f.plan, &f.dryRun)
	mergeSchemaAnnotation(cmd, SchemaAnnotation{SideEffect: "write", Capabilities: []string{"auth", "ws"}, ResourceType: "dashboard", PayloadSchema: &SchemaObjectContract{Type: "object", Open: true, Description: "Deep merge into --target; arrays replace atomically, null is a value. Use --remove for deletion. Followed by --set edits then --remove edits."}})
	return cmd
}

func parseDashboardPatch(f *dashboardPatchFlags) (dashboardpatch.Options, error) {
	opts := dashboardpatch.Options{Spec: dashboardpatch.Spec{Target: f.target}, Plan: planRequested(f.plan, f.dryRun), IfMatch: f.ifMatch, DiffLimit: f.diffLimit}
	if f.timeout <= 0 || f.diffLimit < 1 || f.diffLimit > 500 {
		return opts, fmt.Errorf("invalid timeout or diff-limit")
	}
	if !opts.Plan && opts.IfMatch == "" {
		return opts, client.NewError("PRECONDITION_REQUIRED", "Apply requires --if-match from a dashboard patch --plan result.")
	}
	if f.input.Data != "" && f.input.File != "" {
		return opts, fmt.Errorf("provide only one of --data or --file")
	}
	if f.input.Data != "" || f.input.File != "" {
		merge, err := input.ParseInputExact(f.input.Data, f.input.File, f.input.Format)
		if err != nil {
			return opts, fmt.Errorf("invalid patch object input")
		}
		opts.Spec.Merge = merge
	}
	for _, set := range f.set {
		path, raw, ok := strings.Cut(set, "=")
		if !ok {
			return opts, fmt.Errorf("invalid --set: expected /field=JSON")
		}
		var value any
		if err := input.DecodeJSONExact([]byte(raw), &value); err != nil {
			return opts, fmt.Errorf("invalid --set JSON value")
		}
		opts.Spec.Edits = append(opts.Spec.Edits, dashboardpatch.Edit{Path: path, Value: value})
	}
	for _, path := range f.remove {
		opts.Spec.Edits = append(opts.Spec.Edits, dashboardpatch.Edit{Path: path, Remove: true})
	}
	if len(opts.Spec.Merge) == 0 && len(opts.Spec.Edits) == 0 {
		return opts, fmt.Errorf("provide a nonempty patch via --data, --file, --set, or --remove")
	}
	return opts, nil
}

func runDashboardPatch(cmd *cobra.Command, urlPath string, f *dashboardPatchFlags) error {
	opts, err := parseDashboardPatch(f)
	if err != nil {
		return err
	}
	signalCtx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, f.timeout)
	defer cancel()
	ws, err := getWSClientContext(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := ws.Close(); err != nil {
			noteWarning("Dashboard connection close failed.")
		}
	}()
	result, runErr := dashboardpatch.Run(ctx, client.DashboardStore{WS: ws, URLPath: urlPath}, opts)
	result.Dashboard = urlPath
	redactDashboardResult(&result)
	if runErr != nil {
		code := "DASHBOARD_PATCH_FAILED"
		var failure *dashboardpatch.Failure
		if errors.As(runErr, &failure) {
			code = failure.Code
		}
		details := map[string]any{"result": result, "retryable": false}
		var apiErr *client.APIError
		if errors.As(runErr, &apiErr) {
			details["cause_code"] = apiErr.Code
		}
		return &client.APIError{Code: code, Message: "Dashboard patch did not complete; inspect result before retrying.", Details: details, Category: "dashboard_patch"}
	}
	if result.Saved == nil {
		noteWarning("Save acknowledgement was lost; desired stored state was verified by read-back.")
	}
	output.PrintOutputWithContext(result, getTextMode(), "", output.EnvelopeContext{Operation: "patch", ResourceType: "dashboard"})
	return nil
}

func redactDashboardResult(result *dashboardpatch.Result) {
	for i := range result.Changes {
		change := &result.Changes[i]
		change.Before = redact.Value(change.Path, change.Before)
		change.After = redact.Value(change.Path, change.After)
	}
}
