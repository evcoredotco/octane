// Package primitive_test exercises the close primitive keyword
// (spec 004 §10, item 3) against mock.MockState and mock.MockStation.
//
// Task: T-004-05
// AC1: The close keyword's Func calls Close() on the station registered
// under the given handle; a second IsOpen() call must return false.
//
// NOTE on "non-existent handle" path: mock.State.Station() panics (rather
// than returning an error) when the handle is not registered, because an
// unregistered handle is defined as a test-setup bug in the mock contract
// (pkg/keywords/api/mock/mock.go).  The production closeStation function
// wraps whatever error state.Station returns, but the mock cannot exercise
// that branch without a real runtime State.  See the hand-off note at the
// end of this file.

package primitive_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
	// Blank import registers all primitive keywords at init() time.
	_ "github.com/evcoreco/octane/pkg/keywords/primitive"
)

// ── Named constants ───────────────────────────────────────────────────────────

const (
	// handleClose is the station handle name used across close tests.
	handleClose = "CP02"

	// patternClose is the step text for the close keyword.
	patternClose = "close station {station:string}"
)

// ── tests ─────────────────────────────────────────────────────────────────────

// Test_primitive_closeStation verifies that calling the close keyword's Func
// invokes Close() on the mock station so that IsOpen() returns false (AC1).
func Test_primitive_closeStation(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()
	state.RegisterStation(handleClose, station)

	keywordFunc := resolveFunc(t, patternClose)

	args := api.NewArgs(map[string]any{
		"station": handleClose,
	})

	err := keywordFunc(context.Background(), state, args)
	if err != nil {
		t.Fatalf("close keyword Func: unexpected error: %v", err)
	}

	// Invariant: the station must be closed after the keyword executes.
	if station.IsOpen() {
		t.Errorf(
			"Station(%q).IsOpen() after close: want false, got true",
			handleClose,
		)
	}
}

// Test_primitive_closeStation_AlreadyClosed verifies that closing a station
// whose underlying mock.Station returns an error from Close() propagates
// that error from the keyword's Func.
func Test_primitive_closeStation_CloseError(t *testing.T) {
	t.Parallel()

	// closeErrorStation wraps mock.Station but makes Close() return an error.
	// This exercises the error-propagation branch in closeStation.
	state := mock.NewMockState()
	station := &closeErrorStation{inner: mock.NewMockStation()}
	state.RegisterStation(handleClose, station)

	keywordFunc := resolveFunc(t, patternClose)

	args := api.NewArgs(map[string]any{
		"station": handleClose,
	})

	err := keywordFunc(context.Background(), state, args)

	// Invariant: a Close() error must be surfaced by the keyword.
	if err == nil {
		t.Fatal(
			"close keyword Func: want error when Close() fails, got nil",
		)
	}

	if !errors.Is(err, errCloseStub) {
		t.Errorf(
			"close keyword Func: want errors.Is(err, errCloseStub), got %v",
			err,
		)
	}
}

// Test_primitive_closeStation_LogsCloseMessage verifies that the close keyword
// emits a log message that includes the handle name after closing.
func Test_primitive_closeStation_LogsCloseMessage(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()
	state.RegisterStation(handleClose, station)

	keywordFunc := resolveFunc(t, patternClose)

	args := api.NewArgs(map[string]any{
		"station": handleClose,
	})

	if err := keywordFunc(context.Background(), state, args); err != nil {
		t.Fatalf("close keyword Func: unexpected error: %v", err)
	}

	logs := state.Logs()

	// Invariant: at least one log line must mention the handle name.
	found := false

	for _, line := range logs {
		if strings.Contains(line, handleClose) {
			found = true

			break
		}
	}

	if !found {
		t.Errorf(
			"close keyword Func: no log line mentioning handle %q; logs: %v",
			handleClose,
			logs,
		)
	}
}

// ── test-local helpers ────────────────────────────────────────────────────────

// errCloseStub is the sentinel error returned by closeErrorStation.Close().
var errCloseStub = errors.New("stub: Close failed")

// closeErrorStation is a minimal api.Station wrapper whose Close() always
// returns errCloseStub. All other methods delegate to a real mock.Station.
type closeErrorStation struct {
	inner *mock.Station
}

func (s *closeErrorStation) Send(ctx context.Context, frame []any) error {
	return s.inner.Send(ctx, frame)
}

func (s *closeErrorStation) Expect(ctx context.Context) ([]any, error) {
	return s.inner.Expect(ctx)
}

func (s *closeErrorStation) Close() error {
	return errCloseStub
}

func (s *closeErrorStation) IsOpen() bool {
	return s.inner.IsOpen()
}

// ── Hand-off note to backend agent ───────────────────────────────────────────
//
// GAP IDENTIFIED (T-004-05 → backend):
//
// mock.State.Station() panics when a handle is not registered (see
// pkg/keywords/api/mock/mock.go:70-81). The production closeStation (and
// assertConnectionOpen / assertConnectionClosed) calls state.Station() and
// wraps its error return with fmt.Errorf. However, the mock never returns an
// error from Station() — it panics instead.
//
// This means the "close a non-existent station handle" code path in
// closeStation cannot be unit-tested through mock.State. To test it, either:
//   (a) mock.State.Station() should return an error on missing handles
//       instead of panicking, or
//   (b) a second mock variant (e.g., mock.NewMockStateStrict()) that returns
//       an error should be provided.
//
// Until one of these is resolved, the error-from-missing-handle branch in
// close.go, status.go is unreachable from black-box unit tests using mock.State.
