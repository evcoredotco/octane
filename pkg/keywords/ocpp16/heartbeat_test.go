package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── sendHeartbeat tests ───────────────────────────────────────────────────────

// Test_sendHeartbeat_sendsCALLFrame verifies that the keyword sends a CALL
// frame with action="Heartbeat" and an empty payload map.
func Test_sendHeartbeat_sendsCALLFrame(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCallResult, "octane-heartbeat-1",
		map[string]any{"currentTime": currentTimeValid},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternSendHeartbeat)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("sendHeartbeat: unexpected error: %v", err)
	}

	payload := requireSentCallPayload(
		t,
		station,
		"sendHeartbeat",
		actionHeartbeat,
	)
	if len(payload) != emptyCollectionCount {
		t.Errorf("Heartbeat payload: want empty map, got %v", payload)
	}
}

// ── csmsRespondsToHeartbeat tests ────────────────────────────────────────────

// Test_csmsRespondsToHeartbeat_passesOnValidResponse verifies that the keyword
// passes when the CSMS returns a valid Heartbeat.conf with currentTime.
func Test_csmsRespondsToHeartbeat_passesOnValidResponse(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-heartbeat-1",
		map[string]any{"currentTime": currentTimeValid},
	})

	sendFn := resolveFunc(t, patternSendHeartbeat)
	sendArgs := api.NewArgs(map[string]any{"station": stationHandle})

	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendHeartbeat: %v", err)
	}

	respondFn := resolveFunc(t, patternCSMSRespondsHB)
	respondArgs := api.NewArgs(map[string]any{"timeout": defaultTimeout})

	err := respondFn(context.Background(), state, respondArgs)
	if err != nil {
		t.Errorf("csmsRespondsToHeartbeat: want nil, got %v", err)
	}
}

// ── heartbeatResponseIncludesCurrentTime tests ───────────────────────────────

// Test_heartbeatResponseIncludesCurrentTime_passesValid verifies that the
// keyword passes when the stashed payload has a valid RFC 3339 currentTime.
func Test_heartbeatResponseIncludesCurrentTime_passesValid(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash(lastPayloadKeyTest, map[string]any{
		"currentTime": currentTimeValid,
	})

	fn := resolveFunc(t, patternHBCurrentTime)
	args := api.NewArgs(map[string]any{})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("heartbeatResponseIncludesCurrentTime: want nil, got %v", err)
	}
}

// Test_heartbeatResponseIncludesCurrentTime_failsMissing verifies that the
// keyword returns an error when the stashed payload is missing currentTime.
func Test_heartbeatResponseIncludesCurrentTime_failsMissing(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash(lastPayloadKeyTest, map[string]any{})

	fn := resolveFunc(t, patternHBCurrentTime)
	args := api.NewArgs(map[string]any{})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error(
			"heartbeatResponseIncludesCurrentTime: want error for missing currentTime, got nil",
		)
	}
}
