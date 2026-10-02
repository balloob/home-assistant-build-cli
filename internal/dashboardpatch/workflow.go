package dashboardpatch

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Store interface {
	Load(context.Context) (map[string]any, error)
	Save(context.Context, map[string]any) error
}

type Options struct {
	Spec      Spec
	Plan      bool
	IfMatch   string
	DiffLimit int
}

type Result struct {
	Dashboard        string   `json:"dashboard,omitempty"`
	Status           string   `json:"status"`
	Target           string   `json:"target"`
	BaseRevision     string   `json:"base_revision,omitempty"`
	ResultRevision   string   `json:"result_revision,omitempty"`
	ObservedRevision string   `json:"observed_revision,omitempty"`
	ObservedAt       string   `json:"observed_at"`
	WouldChange      bool     `json:"would_change"`
	Saved            *bool    `json:"saved"`
	Reloaded         string   `json:"reloaded"`
	Verified         bool     `json:"verified"`
	Changes          []Change `json:"changes"`
	ChangeCount      int      `json:"change_count"`
	DiffComplete     bool     `json:"diff_complete"`
	Requests         int      `json:"requests"`
	Validation       string   `json:"validation"`
	Limits           []string `json:"limits"`
}

type Failure struct {
	Code  string
	Cause error
}

func (e *Failure) Error() string { return e.Code }
func (e *Failure) Unwrap() error { return e.Cause }

func fail(result Result, code string, cause error) (Result, error) {
	return result, &Failure{code, cause}
}

// Run never retries a write. A failed acknowledgement leaves saved unknown;
// one read may still verify the desired state. A hash is not a server-side CAS.
func Run(ctx context.Context, store Store, options Options) (Result, error) {
	no := false
	result := Result{Status: "failed", Target: options.Spec.Target, Saved: &no, Reloaded: "not_applicable", Changes: []Change{}, Validation: "local_structure", Limits: []string{
		"HA has no conditional dashboard save: edits between the final read and save can be overwritten.",
		"Read-back verifies stored JSON at observation time, not rendering, referenced entities, resources, or future state.",
		"Preview does not check write permission or server acceptance; no reload is needed for storage dashboards.",
	}}
	if !options.Plan && options.IfMatch == "" {
		return fail(result, "PRECONDITION_REQUIRED", nil)
	}
	if options.DiffLimit < 1 {
		return fail(result, "INVALID_DIFF_LIMIT", nil)
	}
	before, err := load(ctx, store, &result)
	if err != nil {
		return fail(result, "READ_FAILED", err)
	}
	result.BaseRevision, err = Revision(before)
	if err != nil {
		return fail(result, "INVALID_CONFIG", err)
	}
	if options.IfMatch != "" && options.IfMatch != result.BaseRevision {
		return fail(result, "CONFLICT", nil)
	}
	after, changes, err := Apply(before, options.Spec)
	if err != nil {
		return fail(result, "INVALID_PATCH", err)
	}
	result.ResultRevision, err = Revision(after)
	if err != nil {
		return fail(result, "INVALID_PATCH", err)
	}
	result.WouldChange = result.BaseRevision != result.ResultRevision
	result.ChangeCount = len(changes)
	result.DiffComplete = len(changes) <= options.DiffLimit
	result.Changes = changes[:min(len(changes), options.DiffLimit)]
	if options.Plan {
		result.Status = "planned"
		return result, nil
	}
	if !result.WouldChange {
		result.Status = "noop"
		result.Verified = true
		return result, nil
	}
	current, err := load(ctx, store, &result)
	if err != nil {
		return fail(result, "READ_FAILED", err)
	}
	revision, err := Revision(current)
	if err != nil {
		return fail(result, "INVALID_CONFIG", err)
	}
	if revision != result.BaseRevision {
		return fail(result, "CONFLICT", nil)
	}
	return saveAndVerify(ctx, store, after, result)
}

func load(ctx context.Context, store Store, result *Result) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result.Requests++
	value, err := store.Load(ctx)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, fmt.Errorf("dashboard config is not an object")
	}
	result.ObservedAt = time.Now().UTC().Format(time.RFC3339Nano)
	result.ObservedRevision, err = Revision(value)
	return value, err
}

// Rejected marks a definite server rejection, distinct from a lost connection
// or deadline after sending a save. Adapters must not mark transport errors.
type Rejected struct{ Err error }

func (e *Rejected) Error() string { return "save rejected" }
func (e *Rejected) Unwrap() error { return e.Err }

func saveAndVerify(ctx context.Context, store Store, after map[string]any, result Result) (Result, error) {
	if err := ctx.Err(); err != nil {
		return fail(result, "CANCELLED", err)
	}
	result.Requests++
	saveErr := store.Save(ctx, after)
	if saveErr != nil {
		var rejected *Rejected
		if errors.As(saveErr, &rejected) {
			return fail(result, "SAVE_REJECTED", saveErr)
		}
		result.Saved = nil
		result.Status = "uncertain"
	} else {
		yes := true
		result.Saved = &yes
		result.Status = "saved"
	}
	if _, err := load(ctx, store, &result); err != nil {
		return fail(result, "VERIFICATION_UNAVAILABLE", err)
	}
	if result.ObservedRevision != result.ResultRevision {
		return fail(result, "VERIFICATION_MISMATCH", saveErr)
	}
	result.Verified = true
	result.Status = "verified"
	return result, nil
}
