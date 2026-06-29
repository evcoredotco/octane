package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── csmsEnqueuesUnlockConnector tests ─────────────────────────────────────────

// Test_csmsEnqueuesUnlockConnector_stashesCallID verifies that the keyword
// parses the inbound UnlockConnector CALL and stashes the uniqueID.
func Test_csmsEnqueuesUnlockConnector_stashesCallID(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionUnlockConnector,
		map[string]any{"connectorId": float64(connectorIDOne)},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendUnlock)
	args := api.NewArgs(map[string]any{
		"connectorId": connectorIDOne,
		"station":     stationHandle,
		"timeout":     defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("csmsEnqueuesUnlockConnector: want nil, got %v", err)
	}

	val, ok := state.Pop("ocpp16:csms_call:" + stationHandle + ":UnlockConnector")
	if !ok {
		t.Fatal("csmsEnqueuesUnlockConnector: want stashed uniqueID, got nothing")
	}

	uid, ok := val.(string)
	if !ok {
		t.Fatalf("stashed uniqueID: want string, got %T", val)
	}

	if uid != csmsUniqueID {
		t.Errorf("stashed uniqueID: want %q, got %q", csmsUniqueID, uid)
	}
}

// Test_csmsEnqueuesUnlockConnector_errorOnWrongConnector verifies that the
// keyword returns an error when the connectorId in the payload does not match.
func Test_csmsEnqueuesUnlockConnector_errorOnWrongConnector(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionUnlockConnector,
		map[string]any{"connectorId": float64(connectorIDTwo)},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendUnlock)
	args := api.NewArgs(map[string]any{
		"connectorId": connectorIDOne,
		"station":     stationHandle,
		"timeout":     defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("csmsEnqueuesUnlockConnector: want error for wrong connectorId, got nil")
	}
}

// ── stationRespondsToUnlockConnector tests ────────────────────────────────────

// Test_stationRespondsToUnlockConnector_sendsUnlocked verifies that the respond
// keyword sends a CALLRESULT with status="Unlocked".
func Test_stationRespondsToUnlockConnector_sendsUnlocked(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:csms_call:"+stationHandle+":UnlockConnector", csmsUniqueID)

	fn := resolveFunc(t, patternStationRespondsUnlock)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"status":  statusUnlocked,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("stationRespondsToUnlockConnector: want nil, got %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("stationRespondsToUnlockConnector: want 1 sent frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame[0] != msgTypeCallResult {
		t.Errorf("frame[0]: want %v (CALLRESULT), got %v", msgTypeCallResult, frame[0])
	}

	if frame[1] != csmsUniqueID {
		t.Errorf("frame[1]: want %q, got %v", csmsUniqueID, frame[1])
	}

	respPayload, ok := frame[2].(map[string]any)
	if !ok {
		t.Fatalf("frame[2]: want map[string]any, got %T", frame[2])
	}

	if respPayload["status"] != statusUnlocked {
		t.Errorf("payload.status: want %q, got %v", statusUnlocked, respPayload["status"])
	}
}
