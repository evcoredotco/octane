// Package primitive_test exercises the connection-open primitive keywords
// (spec 004 §10, items 1–2) against a local WebSocket server created with
// net/http/httptest. These tests do not reach the public internet.
//
// Task: T-004-05
// AC1: A registered Station handle appears in mock.State after the open
// keyword's Func executes successfully.
package primitive_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
	// Blank import registers all primitive keywords at init() time.
	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

// ── Named constants ───────────────────────────────────────────────────────────

const (
	// handleStation is the station handle name used across open tests.
	handleStation = "CP01"

	// subprotocolOCPP16 is the OCPP 1.6 subprotocol identifier.
	subprotocolOCPP16 = "ocpp1.6"

	// patternOpen is the step text for the no-subprotocol open keyword.
	patternOpen = "open a WebSocket to {url:string} as station {station:string}"

	// patternOpenWithSubprotocol is the step text for the subprotocol variant.
	patternOpenWithSubprotocol = "open a WebSocket to {url:string} as station" +
		" {station:string} with subprotocol {subprotocol:string}"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// newEchoServer starts a local httptest WebSocket server that accepts any
// connection with no subprotocol preference.  It returns the server and its
// ws:// URL.  The caller must call server.Close() when done.
func newEchoServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := websocket.Accept(w, r, nil)
			if err != nil {
				t.Logf("echo server: Accept error: %v", err)
			}
		}),
	)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	return srv, wsURL
}

// newSubprotocolServer starts a local httptest WebSocket server that accepts
// connections and echoes back the first offered subprotocol.  This mimics a
// CSMS that accepts any proposed subprotocol so that the transport layer
// validation passes.
func newSubprotocolServer(
	t *testing.T,
	subprotocol string,
) (*httptest.Server, string) {
	t.Helper()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := websocket.Accept(
				w,
				r,
				&websocket.AcceptOptions{ //nolint:exhaustruct // only Subprotocols is relevant for test servers
					Subprotocols: []string{subprotocol},
				},
			)
			if err != nil {
				t.Logf("subprotocol server: Accept error: %v", err)
			}
		}),
	)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	return srv, wsURL
}

// resolveFunc finds the Func for the given pattern from the global registry.
// It fails the test immediately if no keyword matches.
func resolveFunc(t *testing.T, pattern string) api.Func {
	t.Helper()

	for _, kw := range registry.All() {
		if kw.Pattern == pattern {
			return kw.Func
		}
	}

	t.Fatalf("resolveFunc: no keyword registered for pattern %q", pattern)

	return nil
}

// ── tests ─────────────────────────────────────────────────────────────────────

// Test_primitive_openWebSocket verifies that calling the open keyword's Func
// registers a station in MockState under the given handle name (AC1).
func Test_primitive_openWebSocket(t *testing.T) {
	t.Parallel()

	srv, wsURL := newEchoServer(t)

	defer srv.Close()

	state := mock.NewMockState()
	keywordFunc := resolveFunc(t, patternOpen)

	args := api.NewArgs(map[string]any{
		"url":     wsURL,
		"station": handleStation,
	})

	err := keywordFunc(context.Background(), state, args)
	if err != nil {
		t.Fatalf("open keyword Func: unexpected error: %v", err)
	}

	// Invariant: a station must be registered under the given handle.
	sta, lookupErr := state.Station(handleStation)
	if lookupErr != nil {
		t.Fatalf(
			"state.Station(%q) after open: unexpected error: %v",
			handleStation,
			lookupErr,
		)
	}

	if sta == nil {
		t.Errorf(
			"state.Station(%q) after open: want non-nil Station, got nil",
			handleStation,
		)
	}

	_ = sta.Close()
}

// Test_primitive_openWebSocket_StationIsOpen verifies that the registered
// station reports IsOpen() == true immediately after the open keyword succeeds.
func Test_primitive_openWebSocket_StationIsOpen(t *testing.T) {
	t.Parallel()

	srv, wsURL := newEchoServer(t)

	defer srv.Close()

	state := mock.NewMockState()
	keywordFunc := resolveFunc(t, patternOpen)

	args := api.NewArgs(map[string]any{
		"url":     wsURL,
		"station": handleStation,
	})

	if err := keywordFunc(context.Background(), state, args); err != nil {
		t.Fatalf("open keyword Func: unexpected error: %v", err)
	}

	sta, err := state.Station(handleStation)
	if err != nil {
		t.Fatalf("state.Station(%q): unexpected error: %v", handleStation, err)
	}

	// Invariant: a freshly opened station must report IsOpen() == true.
	if !sta.IsOpen() {
		t.Errorf(
			"Station(%q).IsOpen() after open: want true, got false",
			handleStation,
		)
	}

	_ = sta.Close()
}

// Test_primitive_openWebSocket_DialError verifies that the open keyword's
// Func returns a wrapped error when the dial target is unreachable.
func Test_primitive_openWebSocket_DialError(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	keywordFunc := resolveFunc(t, patternOpen)

	// Use a URL that cannot possibly accept a connection.
	args := api.NewArgs(map[string]any{
		"url":     "ws://127.0.0.1:1",
		"station": handleStation,
	})

	err := keywordFunc(context.Background(), state, args)

	// Invariant: a dial failure must produce a non-nil error.
	if err == nil {
		t.Fatal("open keyword Func: want error on unreachable URL, got nil")
	}
}

// Test_primitive_openWebSocketWithSubprotocol verifies that calling the
// subprotocol-variant open keyword registers a station under the given handle
// and that the connection negotiated the expected subprotocol (AC1).
func Test_primitive_openWebSocketWithSubprotocol(t *testing.T) {
	t.Parallel()

	srv, wsURL := newSubprotocolServer(t, subprotocolOCPP16)

	defer srv.Close()

	state := mock.NewMockState()
	keywordFunc := resolveFunc(t, patternOpenWithSubprotocol)

	args := api.NewArgs(map[string]any{
		"url":         wsURL,
		"station":     handleStation,
		"subprotocol": subprotocolOCPP16,
	})

	err := keywordFunc(context.Background(), state, args)
	if err != nil {
		t.Fatalf(
			"open-with-subprotocol keyword Func: unexpected error: %v",
			err,
		)
	}

	// Invariant: a station must be registered under the given handle.
	sta, lookupErr := state.Station(handleStation)
	if lookupErr != nil {
		t.Fatalf(
			"state.Station(%q) after open-with-subprotocol: unexpected error: %v",
			handleStation,
			lookupErr,
		)
	}

	if sta == nil {
		t.Errorf(
			"state.Station(%q): want non-nil Station, got nil",
			handleStation,
		)
	}

	_ = sta.Close()
}

// Test_primitive_openWebSocketWithSubprotocol_Mismatch verifies that the
// subprotocol-variant open keyword returns an error when the server does not
// agree on the requested subprotocol.
func Test_primitive_openWebSocketWithSubprotocol_Mismatch(t *testing.T) {
	t.Parallel()

	// The server accepts only "ocpp2.0.1"; the client requests "ocpp1.6".
	srv, wsURL := newSubprotocolServer(t, "ocpp2.0.1")

	defer srv.Close()

	state := mock.NewMockState()
	keywordFunc := resolveFunc(t, patternOpenWithSubprotocol)

	args := api.NewArgs(map[string]any{
		"url":         wsURL,
		"station":     handleStation,
		"subprotocol": subprotocolOCPP16,
	})

	err := keywordFunc(context.Background(), state, args)

	// Invariant: a subprotocol mismatch must produce a non-nil error.
	if err == nil {
		t.Fatal(
			"open-with-subprotocol keyword Func: want error on subprotocol " +
				"mismatch, got nil",
		)
	}
}

// Test_primitive_openWebSocket_LogsConnectionMessage verifies that the open
// keyword emits a log message that includes the handle name after a
// successful connection.
func Test_primitive_openWebSocket_LogsConnectionMessage(t *testing.T) {
	t.Parallel()

	srv, wsURL := newEchoServer(t)

	defer srv.Close()

	state := mock.NewMockState()
	keywordFunc := resolveFunc(t, patternOpen)

	args := api.NewArgs(map[string]any{
		"url":     wsURL,
		"station": handleStation,
	})

	if err := keywordFunc(context.Background(), state, args); err != nil {
		t.Fatalf("open keyword Func: unexpected error: %v", err)
	}

	logs := state.Logs()

	// Invariant: at least one log line must mention the handle name.
	found := false

	for _, line := range logs {
		if strings.Contains(line, handleStation) {
			found = true

			break
		}
	}

	if !found {
		t.Errorf(
			"open keyword Func: no log line mentioning handle %q; got logs: %v",
			handleStation,
			logs,
		)
	}

	sta, _ := state.Station(handleStation)
	_ = sta.Close()
}
