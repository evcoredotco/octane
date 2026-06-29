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

	"github.com/evcoreco/octane/pkg/transport"
)

// TestTLSValidationError asserts that dialling a wss:// endpoint whose
// certificate is signed by an untrusted CA causes Dial to return
// *transport.TLSValidationError, exercising the TLS error-wrapping path without
// requiring a live remote server.
func TestTLSValidationError(t *testing.T) {
	t.Parallel()

	// Start a local HTTPS server (self-signed cert) that accepts WebSocket connections
	srv := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := websocket.Accept(w, r, &websocket.AcceptOptions{
				Subprotocols: []string{"ocpp1.6"},
			})
			if err != nil {
				return
			}
		}),
	)
	defer srv.Close()

	// Convert https:// to wss://
	wsURL := strings.Replace(srv.URL, "https://", "wss://", 1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := transport.Dial(ctx, wsURL, transport.DialOptions{
		Subprotocols: []string{"ocpp1.6"},
	})
	if err == nil {
		t.Fatal("expected TLS error, got nil")
	}

	var tlsErr *transport.TLSValidationError
	if !errors.As(err, &tlsErr) {
		t.Errorf("expected *transport.TLSValidationError, got %T: %v", err, err)
	}
}
