package cmd

import (
	"testing"
)

type fakeMarketplaceSender struct {
	repositories []interface{}
	calls        int
}

func (f *fakeMarketplaceSender) SendCommand(cmdType string, params map[string]interface{}) (interface{}, error) {
	f.calls++
	return f.repositories, nil
}

func TestResolveMarketplaceRepository(t *testing.T) {
	ws := &fakeMarketplaceSender{repositories: []interface{}{
		map[string]interface{}{"id": "444350375", "full_name": "piitaya/lovelace-mushroom"},
	}}

	id, err := resolveMarketplaceRepository(ws, "444350375")
	if err != nil || id != "444350375" || ws.calls != 0 {
		t.Fatalf("ID: got %q, %v after %d calls", id, err, ws.calls)
	}

	id, err = resolveMarketplaceRepository(ws, "Piitaya/Lovelace-Mushroom")
	if err != nil || id != "444350375" {
		t.Fatalf("full name: got %q, %v", id, err)
	}

	if _, err := resolveMarketplaceRepository(ws, "nobody/nothing"); err == nil {
		t.Fatal("expected an error for an unknown full name")
	}
}

func TestMarketplaceRepositoryMatches(t *testing.T) {
	repository := map[string]interface{}{
		"name":        "Mushroom",
		"full_name":   "piitaya/lovelace-mushroom",
		"description": "Build a beautiful dashboard",
		"topics":      []interface{}{"lovelace-card"},
	}
	for _, search := range []string{"mushroom", "piitaya", "beautiful", "lovelace-card"} {
		if !marketplaceRepositoryMatches(repository, search) {
			t.Errorf("%q should match", search)
		}
	}
	if marketplaceRepositoryMatches(repository, "thermostat") {
		t.Error("thermostat should not match")
	}
}
