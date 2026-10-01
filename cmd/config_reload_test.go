package cmd

import (
	"testing"
	"time"
)

type fakeConfigPoster struct {
	calls   []string
	onPost  func()
	results interface{}
}

func (f *fakeConfigPoster) Post(endpoint string, body interface{}) (interface{}, error) {
	f.calls = append(f.calls, "post "+endpoint)
	if f.onPost != nil {
		f.onPost()
	}
	return f.results, nil
}

func (f *fakeConfigPoster) CallService(domain, service string, data map[string]interface{}) (interface{}, error) {
	f.calls = append(f.calls, "service "+domain+"."+service)
	return nil, nil
}

type fakeEventSubscriber struct {
	events       chan map[string]interface{}
	eventType    string
	unsubscribed bool
	log          *[]string
}

func (f *fakeEventSubscriber) SubscribeEvents(eventType string) (<-chan map[string]interface{}, func(), error) {
	f.eventType = eventType
	*f.log = append(*f.log, "subscribe "+eventType)
	return f.events, func() { f.unsubscribed = true }, nil
}

func TestPostConfigWaitsForReloadEvent(t *testing.T) {
	rest := &fakeConfigPoster{results: map[string]interface{}{"result": "ok"}}
	sub := &fakeEventSubscriber{events: make(chan map[string]interface{}, 1), log: &rest.calls}
	// Home Assistant fires the event only after the save
	rest.onPost = func() { sub.events <- map[string]interface{}{"event_type": "automation_reloaded"} }

	result, err := postConfigAndWait(rest, sub, "config/automation/config/", "a1", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(map[string]interface{})["result"] != "ok" {
		t.Fatalf("result = %v", result)
	}
	want := []string{"subscribe automation_reloaded", "post config/automation/config/a1"}
	if len(rest.calls) != 2 || rest.calls[0] != want[0] || rest.calls[1] != want[1] {
		t.Fatalf("calls = %v, want %v", rest.calls, want)
	}
	if !sub.unsubscribed {
		t.Fatal("subscription not ended")
	}
}

func TestPostConfigReturnsAfterTimeout(t *testing.T) {
	orig := configReloadTimeout
	configReloadTimeout = 10 * time.Millisecond
	defer func() { configReloadTimeout = orig }()

	rest := &fakeConfigPoster{}
	sub := &fakeEventSubscriber{events: make(chan map[string]interface{}), log: &rest.calls}
	if _, err := postConfigAndWait(rest, sub, "config/scene/config/", "s1", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.eventType != "scene_reloaded" {
		t.Fatalf("eventType = %q", sub.eventType)
	}
}

func TestPostConfigReloadsScripts(t *testing.T) {
	rest := &fakeConfigPoster{}
	if _, err := postConfigAndWait(rest, nil, "config/script/config/", "x", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"post config/script/config/x", "service script.reload"}
	if len(rest.calls) != 2 || rest.calls[0] != want[0] || rest.calls[1] != want[1] {
		t.Fatalf("calls = %v, want %v", rest.calls, want)
	}
}

func TestPostConfigOtherPrefixOnlyPosts(t *testing.T) {
	rest := &fakeConfigPoster{}
	if _, err := postConfigAndWait(rest, nil, "config/other/", "x", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rest.calls) != 1 {
		t.Fatalf("calls = %v", rest.calls)
	}
}
