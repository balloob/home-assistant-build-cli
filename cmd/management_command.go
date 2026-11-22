package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/home-assistant/hab/internal/redact"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

type managementSpec struct {
	Use, Short                  string
	Transport, Method, Endpoint string
	Args                        []string
	OptionalArg                 bool
	Input, RequireInput         bool
	Required                    []string
	Effect                      string
	// Prepare handles API payloads whose shape differs from positional arguments.
	Prepare func(map[string]any) error
}

type managementFlags struct {
	input               InputFlags
	plan, dryRun, force bool
	timeout             time.Duration
}

func newManagementCommand(s managementSpec) *cobra.Command {
	f := &managementFlags{}
	cmd := &cobra.Command{Use: s.Use, Short: s.Short}
	cmd.Args = cobra.ExactArgs(len(s.Args))
	if s.OptionalArg {
		cmd.Args = cobra.RangeArgs(len(s.Args)-1, len(s.Args))
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		endpoint, data, err := prepareManagementRequest(s, f, args)
		if err != nil {
			return err
		}
		if planRequested(f.plan, f.dryRun) {
			printMutationPlan(MutationPlan{WouldChange: true, Target: map[string]any{"command": schemaPath(cmd)}, Inputs: map[string]any{"payload": redact.Hidden}, Steps: []string{s.Short}, RequiresConfirmation: s.Effect == "destructive" && !f.force}, getTextMode(), defaultResourceTypeForCommand(cmd))
			return nil
		}
		if s.Effect == "destructive" {
			if err := confirmAction(f.force, s.Short+"?", s.Short); err != nil {
				return err
			}
		}
		result, err := executeManagementRequest(cmd, s, endpoint, data, f.timeout)
		if err != nil {
			return err
		}
		output.PrintOutput(result, getTextMode(), "")
		return nil
	}
	if s.Input {
		f.input.Register(cmd)
	}
	cmd.Flags().DurationVar(&f.timeout, "timeout", 30*time.Second, "Maximum request duration (use 10m for long operations)")
	if s.Effect != "read" {
		registerMutationPlanFlags(cmd, &f.plan, &f.dryRun)
	}
	if s.Effect == "destructive" {
		cmd.Flags().BoolVar(&f.force, "force", false, "Skip confirmation")
	}
	caps := []string{"auth", s.Transport}
	if s.Transport == "supervisor" {
		caps = []string{"auth", "ws", "supervisor"}
	}
	if strings.HasPrefix(s.Endpoint, "hassio/") {
		caps = append(caps, "supervisor")
	}
	annotation := SchemaAnnotation{SideEffect: s.Effect, OutputMode: "json_envelope", Capabilities: caps, GuideTopic: "management", InputSources: []string{"args", "flags"}}
	if s.Input {
		annotation.InputSources = append(annotation.InputSources, "data", "file", "stdin")
		annotation.PayloadSchema = &SchemaObjectContract{Type: "object", Open: true, Description: "Home Assistant request fields. Server validation applies."}
		for _, key := range s.Required {
			annotation.PayloadSchema.Fields = append(annotation.PayloadSchema.Fields, SchemaField{Name: key, Type: "any", Required: true})
		}
	}
	mergeSchemaAnnotation(cmd, annotation)
	return cmd
}

func prepareManagementRequest(s managementSpec, f *managementFlags, args []string) (string, map[string]any, error) {
	if f.timeout <= 0 {
		return "", nil, fmt.Errorf("--timeout must be positive")
	}
	if f.input.Data != "" && f.input.File != "" {
		return "", nil, fmt.Errorf("--data and --file are mutually exclusive")
	}
	data := map[string]any{}
	if s.Input && (f.input.Data != "" || f.input.File != "") {
		var err error
		file := f.input.File
		if file == "-" {
			file = ""
		}
		data, err = input.ParseInputExact(f.input.Data, file, f.input.Format)
		if err != nil || data == nil {
			return "", nil, fmt.Errorf("invalid JSON/YAML object; use --data or --file (- for stdin)")
		}
	} else if s.RequireInput {
		return "", nil, fmt.Errorf("provide a JSON/YAML object with --data or --file (- for stdin)")
	}
	if s.Transport == "ws" {
		for _, key := range []string{"type", "id"} {
			if _, ok := data[key]; ok {
				return "", nil, fmt.Errorf("payload must not override WebSocket field %s", key)
			}
		}
	}
	endpoint := s.Endpoint
	for i, arg := range args {
		key := s.Args[i]
		if _, ok := data[key]; ok {
			return "", nil, fmt.Errorf("%s is supplied both as an argument and in the payload", key)
		}
		if strings.Contains(endpoint, "{"+key+"}") {
			if err := validateAPIPathSegment(arg); err != nil {
				return "", nil, err
			}
			endpoint = strings.ReplaceAll(endpoint, "{"+key+"}", url.PathEscape(arg))
		} else {
			data[key] = arg
		}
	}
	for _, key := range s.Required {
		_, ok := data[key]
		if !ok {
			return "", nil, fmt.Errorf("payload field %s is required", key)
		}
	}
	if s.Prepare != nil {
		if err := s.Prepare(data); err != nil {
			return "", nil, err
		}
	}
	return endpoint, data, nil
}

func validateAPIPathSegment(value string) error {
	if value == "" || value == "." || value == ".." || strings.ContainsAny(value, "/\\?#%\x00\r\n") {
		return fmt.Errorf("invalid API resource identifier")
	}
	return nil
}

func executeManagementRequest(cmd *cobra.Command, s managementSpec, endpoint string, data map[string]any, timeout time.Duration) (any, error) {
	signalCtx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, timeout)
	defer cancel()
	var result any
	var err error
	if s.Transport == "rest" {
		rest, clientErr := getAuthManager().GetRestClient()
		if clientErr != nil {
			return nil, clientErr
		}
		rest.Timeout = timeout
		noteTransport("rest")
		var body any
		if s.Method != "GET" && s.Method != "DELETE" {
			body = data
		}
		result, err = rest.RequestContext(ctx, s.Method, endpoint, body)
	} else {
		ws, clientErr := getWSClientContext(ctx)
		if clientErr != nil {
			return nil, clientErr
		}
		defer func() {
			if err := ws.Close(); err != nil {
				noteWarning("Management connection close failed.")
			}
		}()
		ws.Timeout = timeout
		command := endpoint
		params := data
		if s.Transport == "supervisor" {
			command = "supervisor/api"
			params = map[string]any{"endpoint": endpoint, "method": s.Method, "timeout": nil}
			if len(data) > 0 || s.Input || s.Method == "post" || s.Method == "put" {
				params["data"] = data
			}
		}
		result, err = ws.SendCommandContext(ctx, command, params)
	}
	if err != nil && s.Effect != "read" {
		return nil, managementMutationError(err)
	}
	return result, err
}

func managementMutationError(err error) error {
	code, message, _ := classifyError(err)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.ErrClosedPipe) {
		code = client.ErrCodeConnectionError
	}
	if strings.EqualFold(code, client.ErrCodeTimeout) {
		code = client.ErrCodeTimeout
	}
	if code != client.ErrCodeTimeout && code != client.ErrCodeCancelled && code != client.ErrCodeConnectionError {
		return err
	}
	return &client.APIError{Code: code, Message: "Operation outcome is unknown. The server may still complete it. Inspect the resource before retrying.", Category: "unknown_outcome", Details: map[string]any{"retryable": false, "cause": message}, SuggestedFix: "Inspect the affected resource before retrying."}
}

func managementGroup(parent *cobra.Command, name, description string) *cobra.Command {
	cmd := &cobra.Command{Use: name, Short: description}
	if parent == rootCmd {
		cmd.GroupID = "other"
	}
	parent.AddCommand(cmd)
	return cmd
}

func addManagementCommands(parent *cobra.Command, specs []managementSpec) {
	for _, s := range specs {
		parent.AddCommand(newManagementCommand(s))
	}
}
