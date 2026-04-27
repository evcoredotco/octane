//go:build reference

package integration_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/octane-project/octane/pkg/transport"
)

// TestSubprotocolMismatch asserts that when the server negotiates a different
// subprotocol than the one requested, Dial returns *transport.ErrSubprotocolMismatch
// with the Got field set to the server-selected subprotocol.
func TestSubprotocolMismatch(t *testing.T) {
	t.Parallel()

	// A fake CSMS that accepts WebSocket but returns "ocpp2.0.1" subprotocol
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = websocket.Accept(w, r, &websocket.AcceptOptions{
			Subprotocols: []string{"ocpp2.0.1"},
		})
	}))
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := transport.Dial(ctx, wsURL, transport.DialOptions{
		Subprotocols: []string{"ocpp1.6"},
	})
	if err == nil {
		t.Fatal("expected subprotocol mismatch error, got nil")
	}

	var mismatch *transport.ErrSubprotocolMismatch
	if !errors.As(err, &mismatch) {
		t.Errorf("expected *transport.ErrSubprotocolMismatch, got %T: %v", err, err)
	}

	if mismatch.Got != "ocpp2.0.1" {
		t.Errorf("Got = %q, want %q", mismatch.Got, "ocpp2.0.1")
	}
}
