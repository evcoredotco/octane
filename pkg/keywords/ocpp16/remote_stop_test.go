package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

const (
	// transactionIDOther is a different transaction ID for mismatch tests.
	transactionIDOther = 99
)

// ── csmsEnqueuesRemoteStop tests ──────────────────────────────────────────────

// Test_csmsEnqueuesRemoteStop_stashesCallID verifies that the keyword parses
// the inbound RemoteStopTransaction CALL and stashes the uniqueID.
func Test_csmsEnqueuesRemoteStop_stashesCallID(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionRemoteStopTransaction,
		map[string]any{"transactionId": float64(transactionIDPositive)},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendRemoteStop)
	args := api.NewArgs(map[string]any{
		"transactionId": transactionIDPositive,
		"station":       stationHandle,
		"timeout":       defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("csmsEnqueuesRemoteStop: want nil, got %v", err)
	}

	val, ok := state.Pop("ocpp16:csms_call:" + stationHandle + ":RemoteStopTransaction")
	if !ok {
		t.Fatal("csmsEnqueuesRemoteStop: want stashed uniqueID, got nothing")
	}

	uid, ok := val.(string)
	if !ok {
		t.Fatalf("stashed uniqueID: want string, got %T", val)
	}

	if uid != csmsUniqueID {
		t.Errorf("stashed uniqueID: want %q, got %q", csmsUniqueID, uid)
	}
}

// Test_csmsEnqueuesRemoteStop_errorOnWrongTransactionId verifies that the
// keyword returns an error when the transactionId field does not match.
func Test_csmsEnqueuesRemoteStop_errorOnWrongTransactionId(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionRemoteStopTransaction,
		map[string]any{"transactionId": float64(transactionIDOther)},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendRemoteStop)
	args := api.NewArgs(map[string]any{
		"transactionId": transactionIDPositive,
		"station":       stationHandle,
		"timeout":       defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("csmsEnqueuesRemoteStop: want error for wrong transactionId, got nil")
	}
}

// ── stationRespondsToRemoteStop tests ─────────────────────────────────────────

// Test_stationRespondsToRemoteStop_sendsAccepted verifies that the respond
// keyword sends a CALLRESULT with status="Accepted".
func Test_stationRespondsToRemoteStop_sendsAccepted(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:csms_call:"+stationHandle+":RemoteStopTransaction", csmsUniqueID)

	fn := resolveFunc(t, patternStationRespondsRemoteStop)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"status":  statusAccepted,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("stationRespondsToRemoteStop: want nil, got %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("stationRespondsToRemoteStop: want 1 sent frame, got %d", len(frames))
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
