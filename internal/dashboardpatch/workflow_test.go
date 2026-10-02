package dashboardpatch

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type memoryStore struct {
	config        map[string]any
	reads, writes int
	beforeRead    func(int)
	saveErr       error
	readErrAt     int
	applyOnError  bool
	afterSave     func()
}

func (s *memoryStore) Load(ctx context.Context) (map[string]any, error) {
	s.reads++
	if s.beforeRead != nil {
		s.beforeRead(s.reads)
	}
	if s.reads == s.readErrAt {
		return nil, errors.New("unavailable")
	}
	return clone(s.config)
}
func (s *memoryStore) Save(ctx context.Context, value map[string]any) error {
	s.writes++
	if s.saveErr == nil || s.applyOnError {
		s.config = value
	}
	if s.afterSave != nil {
		s.afterSave()
	}
	return s.saveErr
}

func fixture() map[string]any {
	return map[string]any{"title": "Home", "views": []any{map[string]any{"path": "home", "cards": []any{
		map[string]any{"type": "tile", "entity": "light.kitchen", "name": "Before", "tap_action": map[string]any{"action": "toggle", "confirmation": true}},
		map[string]any{"type": "markdown", "content": "Retain me"},
	}}}, "custom_field": map[string]any{"keep": true}}
}

func optionsFor(t *testing.T, config map[string]any) Options {
	t.Helper()
	rev, err := Revision(config)
	if err != nil {
		t.Fatal(err)
	}
	return Options{Spec: Spec{Target: "/views/0/cards/0", Merge: map[string]any{"name": "After", "tap_action": map[string]any{"action": "more-info"}}}, IfMatch: rev, DiffLimit: 50}
}

func TestPatchPreviewApplyVerifyAndRepeat(t *testing.T) {
	store := &memoryStore{config: fixture()}
	original, err := clone(store.config)
	if err != nil {
		t.Fatal(err)
	}
	opts := optionsFor(t, store.config)
	opts.Plan = true
	plan, err := Run(context.Background(), store, opts)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Status != "planned" || plan.ChangeCount != 2 || !plan.DiffComplete || plan.Requests != 1 || store.writes != 0 {
		t.Fatalf("bad preview: %+v", plan)
	}
	if !reflect.DeepEqual(store.config, original) {
		t.Fatal("preview mutated its source")
	}
	if plan.Changes[0].Path != "/views/0/cards/0/name" || plan.Changes[0].Before != "Before" || plan.Changes[0].After != "After" {
		t.Fatalf("bad diff: %+v", plan.Changes)
	}
	opts.Plan = false
	result, err := Run(context.Background(), store, opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "verified" || !result.Verified || result.Saved == nil || !*result.Saved || result.Requests != 4 || store.writes != 1 {
		t.Fatalf("bad apply: %+v", result)
	}
	want, _, err := Apply(original, opts.Spec)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(store.config, want) {
		t.Fatal("unexpected applied config")
	}
	kept, err := resolve(store.config, []string{"views", "0", "cards", "0", "tap_action", "confirmation"})
	if err != nil || kept != true {
		t.Fatal("nested field was lost")
	}
	if store.config["title"] != "Home" || !reflect.DeepEqual(store.config["custom_field"], original["custom_field"]) {
		t.Fatal("unrelated fields were lost")
	}
	opts.IfMatch = result.ResultRevision
	result, err = Run(context.Background(), store, opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "noop" || result.WouldChange || result.Requests != 1 || result.ChangeCount != 0 || store.writes != 1 {
		t.Fatalf("repeat was not idempotent: %+v", result)
	}
}

func TestWorkflowFailures(t *testing.T) {
	cases := []struct {
		name         string
		setup        func(*memoryStore, *Options, context.CancelFunc)
		code, status string
		writes       int
	}{
		{"missing_precondition", func(s *memoryStore, o *Options, c context.CancelFunc) { o.IfMatch = "" }, "PRECONDITION_REQUIRED", "failed", 0},
		{"stale_inspection", func(s *memoryStore, o *Options, c context.CancelFunc) { s.config["title"] = "Concurrent" }, "CONFLICT", "failed", 0},
		{"concurrent_unrelated_edit", func(s *memoryStore, o *Options, c context.CancelFunc) {
			s.beforeRead = func(n int) {
				if n == 2 {
					s.config["title"] = "Concurrent"
				}
			}
		}, "CONFLICT", "failed", 0},
		{"permission_denied", func(s *memoryStore, o *Options, c context.CancelFunc) {
			s.saveErr = &Rejected{errors.New("unauthorized")}
		}, "SAVE_REJECTED", "failed", 1},
		{"missing_capability", func(s *memoryStore, o *Options, c context.CancelFunc) { s.readErrAt = 1 }, "READ_FAILED", "failed", 0},
		{"verify_read_failed", func(s *memoryStore, o *Options, c context.CancelFunc) { s.readErrAt = 3 }, "VERIFICATION_UNAVAILABLE", "saved", 1},
		{"verify_mismatch", func(s *memoryStore, o *Options, c context.CancelFunc) {
			s.afterSave = func() { s.config["title"] = "Other writer" }
		}, "VERIFICATION_MISMATCH", "saved", 1},
		{"timeout_unknown_apply", func(s *memoryStore, o *Options, c context.CancelFunc) {
			s.saveErr = context.DeadlineExceeded
			s.readErrAt = 3
		}, "VERIFICATION_UNAVAILABLE", "uncertain", 1},
		{"cancel_before_save", func(s *memoryStore, o *Options, c context.CancelFunc) {
			s.beforeRead = func(n int) {
				if n == 2 {
					c()
				}
			}
		}, "CANCELLED", "failed", 0},
		{"cancel_after_apply", func(s *memoryStore, o *Options, c context.CancelFunc) {
			s.saveErr = context.Canceled
			s.applyOnError = true
			s.afterSave = c
		}, "VERIFICATION_UNAVAILABLE", "uncertain", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &memoryStore{config: fixture()}
			opts := optionsFor(t, store.config)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tc.setup(store, &opts, cancel)
			result, err := Run(ctx, store, opts)
			var failure *Failure
			if !errors.As(err, &failure) || failure.Code != tc.code || result.Status != tc.status || store.writes != tc.writes {
				t.Fatalf("result=%+v error=%v writes=%d", result, err, store.writes)
			}
			if result.Verified {
				t.Fatal("failure claimed verified")
			}
		})
	}
}

func TestTimeoutAfterApplyReconcilesWithoutRetry(t *testing.T) {
	store := &memoryStore{config: fixture(), saveErr: context.DeadlineExceeded, applyOnError: true}
	result, err := Run(context.Background(), store, optionsFor(t, store.config))
	if err != nil || !result.Verified || result.Saved != nil || store.writes != 1 {
		t.Fatalf("result=%+v err=%v writes=%d", result, err, store.writes)
	}
}

func TestDiffBounds(t *testing.T) {
	store := &memoryStore{config: fixture()}
	opts := optionsFor(t, store.config)
	opts.Plan = true
	opts.DiffLimit = 1
	result, err := Run(context.Background(), store, opts)
	if err != nil || result.ChangeCount != 2 || result.DiffComplete || len(result.Changes) != 1 {
		t.Fatalf("%+v %v", result, err)
	}
}
