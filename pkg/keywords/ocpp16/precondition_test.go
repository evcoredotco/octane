package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── csmsIsReachable tests ─────────────────────────────────────────────────────

// Test_csmsIsReachable_passesWhenStationOpen verifies that the keyword passes
// when a non-empty CSMS base URL is configured.
func Test_csmsIsReachable_passesWhenStationOpen(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetCSMSBaseURL("ws://localhost:9210")

	fn := resolveFunc(t, patternCSMSReachable)
	args := api.NewArgs(map[string]any{})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("csmsIsReachable: want nil when URL is set, got %v", err)
	}
}

// Test_csmsIsReachable_failsWhenStationClosed verifies that the keyword returns
// an error when no CSMS base URL is configured.
func Test_csmsIsReachable_failsWhenStationClosed(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	// CSMSBaseURL defaults to empty — no endpoint configured.

	fn := resolveFunc(t, patternCSMSReachable)
	args := api.NewArgs(map[string]any{})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("csmsIsReachable: want error when URL is empty, got nil")
	}
}

// ── operatorProvisionedIdTag tests ────────────────────────────────────────────

// Test_operatorProvisionedIdTag_alwaysPasses verifies that the documentation-only
// precondition keyword always returns nil regardless of idTag or status values.
func Test_operatorProvisionedIdTag_alwaysPasses(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()

	fn := resolveFunc(t, patternOperatorProvisioned)
	args := api.NewArgs(map[string]any{
		"idTag":  idTagValue,
		"status": statusAccepted,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("operatorProvisionedIdTag: want nil (no-op), got %v", err)
	}
}

// ── stationIsRegistered tests ─────────────────────────────────────────────────

// Test_stationIsRegistered_passesWhenStashSet verifies that the keyword passes
// when the station is registered and its connection is open.
func Test_stationIsRegistered_passesWhenStashSet(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	fn := resolveFunc(t, patternStationIsRegistered)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("stationIsRegistered: want nil for open station, got %v", err)
	}
}

// Test_stationIsRegistered_failsWhenNotStashed verifies that the keyword returns
// an error when the station's connection is closed.
func Test_stationIsRegistered_failsWhenNotStashed(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	_ = station.Close()
	state := newState(t, station)

	fn := resolveFunc(t, patternStationIsRegistered)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("stationIsRegistered: want error for closed station, got nil")
	}
}
