package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestContextReadsPreserveNumbersWithoutChangingLegacyReads(t *testing.T) {
	server := contextTestServer(t, func(conn *websocket.Conn) error {
		if err := testAuthenticate(conn); err != nil {
			return err
		}
		for range 2 {
			var request map[string]any
			if err := conn.ReadJSON(&request); err != nil {
				return err
			}
			if err := conn.WriteJSON(map[string]any{"id": request["id"], "type": "result", "success": true, "result": json.Number("9007199254740993")}); err != nil {
				return err
			}
		}
		return nil
	})
	client := NewWebSocketClient(server.URL, "test")
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	value, err := client.SendCommandContext(context.Background(), "lovelace/config", nil)
	if err != nil || value != json.Number("9007199254740993") {
		t.Fatalf("precision lost: %v %v", value, err)
	}
	value, err = client.SendCommand("lovelace/config", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := value.(float64); !ok {
		t.Fatalf("legacy number type changed: %T", value)
	}
}

func contextTestServer(t *testing.T, handler func(*websocket.Conn) error) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer func() {
			// The peer may already have closed after cancellation.
			_ = conn.Close()
		}()
		if err := handler(conn); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func testAuthenticate(conn *websocket.Conn) error {
	if err := conn.WriteJSON(map[string]any{"type": "auth_required"}); err != nil {
		return err
	}
	var auth map[string]any
	if err := conn.ReadJSON(&auth); err != nil {
		return err
	}
	return conn.WriteJSON(map[string]any{"type": "auth_ok"})
}

func TestCommandCancellationAndLateResponse(t *testing.T) {
	received := make(chan struct{})
	release := make(chan struct{})
	server := contextTestServer(t, func(conn *websocket.Conn) error {
		if err := testAuthenticate(conn); err != nil {
			return err
		}
		var first, second map[string]any
		if err := conn.ReadJSON(&first); err != nil {
			return err
		}
		close(received)
		<-release
		if err := conn.WriteJSON(map[string]any{"id": first["id"], "type": "result", "success": true, "result": "late"}); err != nil {
			return err
		}
		if err := conn.ReadJSON(&second); err != nil {
			return err
		}
		return conn.WriteJSON(map[string]any{"id": second["id"], "type": "result", "success": true, "result": "current"})
	})
	client := NewWebSocketClient(server.URL, "test")
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := client.SendCommandContext(ctx, "lovelace/config", nil); done <- err }()
	<-received
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	client.pendingMu.RLock()
	pending := len(client.pending)
	client.pendingMu.RUnlock()
	if pending != 0 {
		t.Fatal("cancelled request left pending state")
	}
	close(release)
	result, err := client.SendCommand("lovelace/config", nil)
	if err != nil || result != "current" {
		t.Fatalf("late response corrupted subsequent read: %v %v", result, err)
	}
}

func TestCommandDeadlineDoesNotRetry(t *testing.T) {
	requests := make(chan map[string]any, 1)
	release := make(chan struct{})
	server := contextTestServer(t, func(conn *websocket.Conn) error {
		if err := testAuthenticate(conn); err != nil {
			return err
		}
		var request map[string]any
		if err := conn.ReadJSON(&request); err != nil {
			return err
		}
		requests <- request
		<-release
		return nil
	})
	defer close(release)
	client := NewWebSocketClient(server.URL, "test")
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := client.SendCommandContext(ctx, "lovelace/config/save", map[string]any{"config": map[string]any{"views": []any{}}})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
	if request := <-requests; request["type"] != "lovelace/config/save" {
		t.Fatalf("wrong request: %v", request)
	}
	if client.messageID.Load() != 1 {
		t.Fatal("write retried")
	}
}

func TestConnectCancellationDuringAuthentication(t *testing.T) {
	connected := make(chan struct{})
	release := make(chan struct{})
	server := contextTestServer(t, func(conn *websocket.Conn) error {
		close(connected)
		<-release
		return nil
	})
	defer close(release)
	client := NewWebSocketClient(server.URL, "test")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- client.ConnectContext(ctx) }()
	<-connected
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("authentication did not preserve cancellation: %v", err)
	}
}
