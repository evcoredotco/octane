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

// TestSubprotocolMismatch asserts that when the server negotiates a different
// subprotocol than the one requested, Dial returns *transport.SubprotocolMismatchError
// with the Got field set to the server-selected subprotocol.
func TestSubprotocolMismatch(t *testing.T) {
	t.Parallel()

	// A fake CSMS that accepts WebSocket but returns an unsupported subprotocol,
	// simulating a server that does not speak OCPP 1.6.
	const wrongSubprotocol = "ocpp_unsupported"

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = websocket.Accept(w, r, &websocket.AcceptOptions{
				Subprotocols: []string{wrongSubprotocol},
			})
		}),
	)
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

	var mismatch *transport.SubprotocolMismatchError
	if !errors.As(err, &mismatch) {
		t.Errorf(
			"expected *transport.SubprotocolMismatchError, got %T: %v",
			err,
			err,
		)
	}

	if mismatch.Got != wrongSubprotocol {
		t.Errorf("Got = %q, want %q", mismatch.Got, wrongSubprotocol)
	}
}
