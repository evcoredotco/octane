package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── sendStatusNotification tests ─────────────────────────────────────────────

// Test_sendStatusNotification_sendsCALLFrame verifies that the keyword sends a
// CALL frame with action="StatusNotification" and the correct payload fields.
func Test_sendStatusNotification_sendsCALLFrame(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCallResult, "octane-statusnotification-1",
		map[string]any{},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternSendStatus)
	args := api.NewArgs(map[string]any{
		"station":     stationHandle,
		"connectorId": connectorIDOne,
		"status":      statusAvailable,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("sendStatusNotification: unexpected error: %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("sendStatusNotification: want 1 sent frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame[0] != msgTypeCall {
		t.Errorf("frame[0]: want %v (CALL), got %v", msgTypeCall, frame[0])
	}

	if frame[2] != actionStatusNotification {
		t.Errorf("frame[2]: want %q, got %v", actionStatusNotification, frame[2])
	}

	payload, ok := frame[3].(map[string]any)
	if !ok {
		t.Fatalf("frame[3]: want map[string]any, got %T", frame[3])
	}

	if payload["connectorId"] != connectorIDOne {
		t.Errorf("payload.connectorId: want %d, got %v", connectorIDOne, payload["connectorId"])
	}

	if payload["status"] != statusAvailable {
		t.Errorf("payload.status: want %q, got %v", statusAvailable, payload["status"])
	}

	if _, exists := payload["errorCode"]; !exists {
		t.Error("payload missing errorCode field")
	}

	if _, exists := payload["timestamp"]; !exists {
		t.Error("payload missing timestamp field")
	}
}

// ── csmsAcknowledgesStatus tests ──────────────────────────────────────────────

// Test_csmsAcknowledgesStatus_passesOnEmptyPayload verifies that the keyword
// passes when the CSMS returns an empty CALLRESULT payload.
func Test_csmsAcknowledgesStatus_passesOnEmptyPayload(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-statusnotification-1",
		map[string]any{},
	})

	sendFn := resolveFunc(t, patternSendStatus)
	sendArgs := api.NewArgs(map[string]any{
		"station":     stationHandle,
		"connectorId": connectorIDOne,
		"status":      statusAvailable,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendStatusNotification: %v", err)
	}

	ackFn := resolveFunc(t, patternCSMSAcksStatus)
	ackArgs := api.NewArgs(map[string]any{"timeout": defaultTimeout})

	err := ackFn(context.Background(), state, ackArgs)
	if err != nil {
		t.Errorf("csmsAcknowledgesStatus: want nil, got %v", err)
	}
}

// ── connectorIsInState tests ──────────────────────────────────────────────────

// Test_connectorIsInState_passesWhenMatches verifies that the keyword passes
// when the stashed connector status matches the expected value.
func Test_connectorIsInState_passesWhenMatches(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	// connectorKey = "ocpp16:connector:{station}:{connectorId}"
	state.Stash("ocpp16:connector:"+stationHandle+":1", statusAvailable)

	fn := resolveFunc(t, patternConnectorState)
	args := api.NewArgs(map[string]any{
		"connectorId": connectorIDOne,
		"station":     stationHandle,
		"state":       statusAvailable,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("connectorIsInState: want nil, got %v", err)
	}
}

// Test_connectorIsInState_failsWhenMismatch verifies that the keyword returns
// an error when the stashed connector status does not match the expected value.
func Test_connectorIsInState_failsWhenMismatch(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:connector:"+stationHandle+":1", statusAvailable)

	fn := resolveFunc(t, patternConnectorState)
	args := api.NewArgs(map[string]any{
		"connectorId": connectorIDOne,
		"station":     stationHandle,
		"state":       statusCharging,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("connectorIsInState: want error for status mismatch, got nil")
	}
}

// Test_connectorIsInState_setsStateFromSendPair verifies that after a
// sendStatusNotification + csmsAcknowledgesStatus pair, the connector state
// is readable by connectorIsInState.
func Test_connectorIsInState_setsStateFromSendPair(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCallResult, "octane-statusnotification-1",
		map[string]any{},
	})
	state := newState(t, station)

	sendFn := resolveFunc(t, patternSendStatus)
	sendArgs := api.NewArgs(map[string]any{
		"station":     stationHandle,
		"connectorId": connectorIDOne,
		"status":      statusAvailable,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendStatusNotification: %v", err)
	}

	ackFn := resolveFunc(t, patternCSMSAcksStatus)
	ackArgs := api.NewArgs(map[string]any{"timeout": defaultTimeout})
	if err := ackFn(context.Background(), state, ackArgs); err != nil {
		t.Fatalf("csmsAcknowledgesStatus: %v", err)
	}

	checkFn := resolveFunc(t, patternConnectorState)
	checkArgs := api.NewArgs(map[string]any{
		"connectorId": connectorIDOne,
		"station":     stationHandle,
		"state":       statusAvailable,
	})

	err := checkFn(context.Background(), state, checkArgs)
	if err != nil {
		t.Errorf("connectorIsInState after send pair: want nil, got %v", err)
	}
}
