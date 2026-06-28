// Package lifecycle_test exercises the lifecycle keyword functions against
// local httptest WebSocket servers and mock states.

package lifecycle_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
	"github.com/evcoreco/octane/pkg/keywords/lifecycle"
	"github.com/evcoreco/octane/pkg/keywords/primitive"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

// ── named constants ──────────────────────────────────────────────────────────

const (
	// stationHandle is the station handle used across lifecycle tests.
	stationHandle = "CP01"

	// subprotocolOCPP16 is the OCPP 1.6 WebSocket subprotocol identifier.
	subprotocolOCPP16 = "ocpp1.6"

	// patternConnect is the step pattern for the connect keyword.
	patternConnect = "station {station:string} connects to the CSMS"

	// patternHandshake is the step pattern for the handshake keyword.
	patternHandshake = "the OCPP-J handshake completes within {timeout:duration}"

	// patternStatus is the step pattern for the connected-state keyword.
	patternStatus = "station {station:string} is in the connected state"

	// connectingStashKey is the stash key set by the connect keyword and
	// consumed by the handshake keyword. Matches lifecycle.connectingStationKey.
	connectingStashKey = "lifecycle:connecting_station"

	// shortDialTimeout is a brief timeout used in unreachable-CSMS tests.
	shortDialTimeout = 500 * time.Millisecond

	// handshakeTimeout is the timeout value passed to the handshake keyword.
	handshakeTimeout = 5 * time.Second
)

// ── TestMain ─────────────────────────────────────────────────────────────────

// TestMain registers all keyword packages once before any test runs.
func TestMain(m *testing.M) {
	primitive.Register()
	lifecycle.Register()
	m.Run()
}

// ── helpers ──────────────────────────────────────────────────────────────────

// newOCPP16Server starts a local WebSocket test server that accepts the
// "ocpp1.6" subprotocol. It returns the server and its ws:// base URL.
// The caller must call srv.Close() when done.
func newOCPP16Server(t *testing.T) (*httptest.Server, string) {
	t.Helper()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//nolint:exhaustruct // only Subprotocols is relevant
			_, err := websocket.Accept(w, r, &websocket.AcceptOptions{
				Subprotocols: []string{subprotocolOCPP16},
			})
			if err != nil {
				t.Logf("ocpp16 server: Accept: %v", err)
			}
		}),
	)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	return srv, wsURL
}

// resolveFunc returns the keyword Func registered for the given pattern.
// It fails the test immediately if no matching keyword is found.
func resolveFunc(t *testing.T, stepPattern string) api.Func {
	t.Helper()

	for _, kw := range registry.All() {
		if kw.Pattern == stepPattern {
			return kw.Func
		}
	}

	t.Fatalf("resolveFunc: no keyword registered for pattern %q", stepPattern)

	return nil
}

// ── connect keyword tests ─────────────────────────────────────────────────────

// Test_lifecycle_connect_errorWhenNoEndpoint verifies that the connect keyword
// returns an error when no CSMS endpoint is configured in the state.
func Test_lifecycle_connect_errorWhenNoEndpoint(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	// CSMSBaseURL is empty by default — no endpoint configured.

	fn := resolveFunc(t, patternConnect)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Fatal("connect keyword: want error when CSMSBaseURL is empty, got nil")
	}
}

// Test_lifecycle_connect_errorWhenCSMSUnreachable verifies that the connect
// keyword returns an error when the CSMS endpoint is configured but the server
// is not reachable.
func Test_lifecycle_connect_errorWhenCSMSUnreachable(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetCSMSBaseURL("ws://127.0.0.1:19999") // nothing listens here

	fn := resolveFunc(t, patternConnect)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	ctx, cancel := context.WithTimeout(context.Background(), shortDialTimeout)
	defer cancel()

	err := fn(ctx, state, args)
	if err == nil {
		t.Fatal("connect keyword: want error for unreachable CSMS, got nil")
	}
}

// Test_lifecycle_connect_registersStationOnSuccess verifies that after the
// connect keyword runs against a live OCPP-J WebSocket server, the station is
// registered in the state and its connection is open.
func Test_lifecycle_connect_registersStationOnSuccess(t *testing.T) {
	t.Parallel()

	srv, wsBaseURL := newOCPP16Server(t)
	defer srv.Close()

	state := mock.NewMockState()
	state.SetCSMSBaseURL(wsBaseURL)

	fn := resolveFunc(t, patternConnect)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("connect keyword: unexpected error: %v", err)
	}

	sta, lookupErr := state.Station(stationHandle)
	if lookupErr != nil {
		t.Fatalf("Station(%q) after connect: unexpected error: %v", stationHandle, lookupErr)
	}

	if !sta.IsOpen() {
		t.Errorf("Station(%q).IsOpen(): want true after connect, got false", stationHandle)
	}
}

// Test_lifecycle_connect_stashesHandleForHandshakeStep verifies that the
// connect keyword stashes the station handle under the well-known key so that
// the subsequent handshake keyword can consume it.
func Test_lifecycle_connect_stashesHandleForHandshakeStep(t *testing.T) {
	t.Parallel()

	srv, wsBaseURL := newOCPP16Server(t)
	defer srv.Close()

	state := mock.NewMockState()
	state.SetCSMSBaseURL(wsBaseURL)

	fn := resolveFunc(t, patternConnect)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("connect keyword: unexpected error: %v", err)
	}

	stashedAny, ok := state.Pop(connectingStashKey)
	if !ok {
		t.Fatalf("Stash(%q): want value after connect, got none", connectingStashKey)
	}

	stashedHandle, ok := stashedAny.(string)
	if !ok {
		t.Fatalf("Stash(%q): want string, got %T", connectingStashKey, stashedAny)
	}

	if stashedHandle != stationHandle {
		t.Errorf(
			"Stash(%q): want %q, got %q",
			connectingStashKey,
			stationHandle,
			stashedHandle,
		)
	}
}

// ── handshake keyword tests ───────────────────────────────────────────────────

// Test_lifecycle_handshake_errorWithoutPriorConnect verifies that the handshake
// keyword returns a descriptive error when no station handle has been stashed
// by a preceding connect step.
func Test_lifecycle_handshake_errorWithoutPriorConnect(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	// No stash entry — simulates running handshake without a connect step.

	fn := resolveFunc(t, patternHandshake)
	args := api.NewArgs(map[string]any{"timeout": handshakeTimeout})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Fatal("handshake keyword: want error without prior connect step, got nil")
	}
}

// Test_lifecycle_handshake_passesForOpenStation verifies that the handshake
// keyword returns nil when the stash contains a station handle and the station
// is open.
func Test_lifecycle_handshake_passesForOpenStation(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation() // open by default
	state.RegisterStation(stationHandle, station)
	state.Stash(connectingStashKey, stationHandle)

	fn := resolveFunc(t, patternHandshake)
	args := api.NewArgs(map[string]any{"timeout": handshakeTimeout})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("handshake keyword: want nil for open station, got %v", err)
	}
}

// Test_lifecycle_handshake_failsForClosedStation verifies that the handshake
// keyword returns an error when the station is registered but closed.
func Test_lifecycle_handshake_failsForClosedStation(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()
	_ = station.Close()
	state.RegisterStation(stationHandle, station)
	state.Stash(connectingStashKey, stationHandle)

	fn := resolveFunc(t, patternHandshake)
	args := api.NewArgs(map[string]any{"timeout": handshakeTimeout})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Fatal("handshake keyword: want error for closed station, got nil")
	}
}

// ── connected-state keyword tests ─────────────────────────────────────────────

// Test_lifecycle_status_errorForUnknownStation verifies that the
// connected-state keyword panics (mock.State panics on unknown handle,
// mirroring missing registration) when the station is not registered.
func Test_lifecycle_status_errorForUnknownStation(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	// No station registered.

	fn := resolveFunc(t, patternStatus)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	defer func() {
		if r := recover(); r == nil {
			t.Error(
				"connected-state keyword: expected panic on unregistered " +
					"station handle, but none occurred",
			)
		}
	}()

	_ = fn(context.Background(), state, args)
}

// Test_lifecycle_status_passesForOpenStation verifies that the connected-state
// keyword returns nil when the station is registered and its connection is open.
func Test_lifecycle_status_passesForOpenStation(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation() // open by default
	state.RegisterStation(stationHandle, station)

	fn := resolveFunc(t, patternStatus)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("connected-state keyword: want nil for open station, got %v", err)
	}
}

// Test_lifecycle_status_failsForClosedStation verifies that the connected-state
// keyword returns an error when the station is registered but closed.
func Test_lifecycle_status_failsForClosedStation(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()
	_ = station.Close()
	state.RegisterStation(stationHandle, station)

	fn := resolveFunc(t, patternStatus)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Fatal("connected-state keyword: want error for closed station, got nil")
	}
}
