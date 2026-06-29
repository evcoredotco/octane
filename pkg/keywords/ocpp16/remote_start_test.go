package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

const (
	// idTagOther is a different idTag value for mismatch tests.
	idTagOther = "ZZYYXX"
)

// ── csmsEnqueuesRemoteStart tests ─────────────────────────────────────────────

// Test_csmsEnqueuesRemoteStart_stashesCallID verifies that the keyword parses
// the inbound RemoteStartTransaction CALL and stashes the uniqueID.
func Test_csmsEnqueuesRemoteStart_stashesCallID(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionRemoteStartTransaction,
		map[string]any{
			"connectorId": float64(connectorIDOne),
			"idTag":       idTagValue,
		},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendRemoteStart)
	args := api.NewArgs(map[string]any{
		"connectorId": connectorIDOne,
		"idTag":       idTagValue,
		"station":     stationHandle,
		"timeout":     defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("csmsEnqueuesRemoteStart: want nil, got %v", err)
	}

	val, ok := state.Pop("ocpp16:csms_call:" + stationHandle + ":RemoteStartTransaction")
	if !ok {
		t.Fatal("csmsEnqueuesRemoteStart: want stashed uniqueID, got nothing")
	}

	uid, ok := val.(string)
	if !ok {
		t.Fatalf("stashed uniqueID: want string, got %T", val)
	}

	if uid != csmsUniqueID {
		t.Errorf("stashed uniqueID: want %q, got %q", csmsUniqueID, uid)
	}
}

// Test_csmsEnqueuesRemoteStart_errorOnWrongIdTag verifies that the keyword
// returns an error when the idTag field does not match the expected value.
func Test_csmsEnqueuesRemoteStart_errorOnWrongIdTag(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionRemoteStartTransaction,
		map[string]any{
			"connectorId": float64(connectorIDOne),
			"idTag":       idTagOther,
		},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendRemoteStart)
	args := api.NewArgs(map[string]any{
		"connectorId": connectorIDOne,
		"idTag":       idTagValue,
		"station":     stationHandle,
		"timeout":     defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("csmsEnqueuesRemoteStart: want error for wrong idTag, got nil")
	}
}

// ── stationRespondsToRemoteStart tests ────────────────────────────────────────

// Test_stationRespondsToRemoteStart_sendsAccepted verifies that the respond
// keyword sends a CALLRESULT with status="Accepted".
func Test_stationRespondsToRemoteStart_sendsAccepted(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:csms_call:"+stationHandle+":RemoteStartTransaction", csmsUniqueID)

	fn := resolveFunc(t, patternStationRespondsRemoteStart)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"status":  statusAccepted,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("stationRespondsToRemoteStart: want nil, got %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("stationRespondsToRemoteStart: want 1 sent frame, got %d", len(frames))
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

	if respPayload["status"] != statusAccepted {
		t.Errorf("payload.status: want %q, got %v", statusAccepted, respPayload["status"])
	}
}
