// Package primitive_test exercises the connection-status assertion keywords
// (spec 004 §10, items 9–10) against mock.MockState and mock.MockStation.
//
// Task: T-004-05
// AC1: "the connection on station {station} is open" passes when
// MockStation.IsOpen() returns true and fails when it returns false.
// "the connection on station {station} is closed" behaves inversely.

package primitive_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── Named constants ──────────────────────────────────────────────────────────

const (
	// handleStatus is the station handle name used across status tests.
	handleStatus = "CP03"

	// patternIsOpen is the step text for the is-open assertion keyword.
	patternIsOpen = "the connection on station {station:string} is open"

	// patternIsClosed is the step text for the is-closed assertion keyword.
	patternIsClosed = "the connection on station {station:string} is closed"
)

// ── "is open" tests ──────────────────────────────────────────────────────────

// Test_primitive_assertConnectionOpen_Passes verifies that the is-open keyword
// returns nil when MockStation.IsOpen() is true.
func Test_primitive_assertConnectionOpen_Passes(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation() // IsOpen() returns true by default.
	state.RegisterStation(handleStatus, station)

	keywordFunc := resolveFunc(t, patternIsOpen)

	args := api.NewArgs(map[string]any{
		"station": handleStatus,
	})

	// Invariant: is-open must return nil when the connection is open.
	err := keywordFunc(context.Background(), state, args)
	if err != nil {
		t.Errorf(
			"assertConnectionOpen on open station: want nil, got %v",
			err,
		)
	}
}

// Test_primitive_assertConnectionOpen_Fails verifies that the is-open keyword
// returns a non-nil error when MockStation.IsOpen() is false.
func Test_primitive_assertConnectionOpen_Fails(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()

	// Force the station to report as closed.
	_ = station.Close()

	state.RegisterStation(handleStatus, station)

	keywordFunc := resolveFunc(t, patternIsOpen)

	args := api.NewArgs(map[string]any{
		"station": handleStatus,
	})

	err := keywordFunc(context.Background(), state, args)

	// Invariant: the is-open keyword must fail when the connection is closed.
	if err == nil {
		t.Error("assertConnectionOpen on closed station: want non-nil error")
	}
}

// ── "is closed" tests ────────────────────────────────────────────────────────

// Test_primitive_assertConnectionClosed_Passes verifies that the is-closed
// keyword returns nil when MockStation.IsOpen() is false.
func Test_primitive_assertConnectionClosed_Passes(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()

	// Force the station to report as closed.
	_ = station.Close()

	state.RegisterStation(handleStatus, station)

	keywordFunc := resolveFunc(t, patternIsClosed)

	args := api.NewArgs(map[string]any{
		"station": handleStatus,
	})

	// Invariant: the is-closed keyword must return nil when the connection
	// is indeed closed.
	err := keywordFunc(context.Background(), state, args)
	if err != nil {
		t.Errorf(
			"assertConnectionClosed on closed station: want nil, got %v",
			err,
		)
	}
}

// Test_primitive_assertConnectionClosed_Fails verifies that the is-closed
// keyword returns a non-nil error when MockStation.IsOpen() is true.
func Test_primitive_assertConnectionClosed_Fails(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation() // IsOpen() returns true by default.
	state.RegisterStation(handleStatus, station)

	keywordFunc := resolveFunc(t, patternIsClosed)

	args := api.NewArgs(map[string]any{
		"station": handleStatus,
	})

	err := keywordFunc(context.Background(), state, args)

	// Invariant: the is-closed keyword must fail when the connection is open.
	if err == nil {
		t.Error("assertConnectionClosed on open station: want non-nil error")
	}

	_ = station.Close()
}
