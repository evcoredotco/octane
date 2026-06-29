package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── csmsEnqueuesChangeAvailability tests ──────────────────────────────────────

// Test_csmsEnqueuesChangeAvailability_stashesCallID verifies that the keyword
// parses the inbound ChangeAvailability CALL and stashes the uniqueID.
func Test_csmsEnqueuesChangeAvailability_stashesCallID(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionChangeAvailability,
		map[string]any{
			"connectorId": float64(connectorIDOne),
			"type":        statusOperative,
		},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendAvail)
	args := api.NewArgs(map[string]any{
		"connectorId": connectorIDOne,
		"availType":   statusOperative,
		"station":     stationHandle,
		"timeout":     defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("csmsEnqueuesChangeAvailability: want nil, got %v", err)
	}

	val, ok := state.Pop("ocpp16:csms_call:" + stationHandle + ":ChangeAvailability")
	if !ok {
		t.Fatal("csmsEnqueuesChangeAvailability: want stashed uniqueID, got nothing")
	}

	uid, ok := val.(string)
	if !ok {
		t.Fatalf("stashed uniqueID: want string, got %T", val)
	}

	if uid != csmsUniqueID {
		t.Errorf("stashed uniqueID: want %q, got %q", csmsUniqueID, uid)
	}
}

// Test_csmsEnqueuesChangeAvailability_errorOnWrongConnector verifies that the
// keyword returns an error when the connectorId in the payload does not match.
func Test_csmsEnqueuesChangeAvailability_errorOnWrongConnector(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionChangeAvailability,
		map[string]any{
			"connectorId": float64(connectorIDTwo),
			"type":        statusOperative,
		},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendAvail)
	args := api.NewArgs(map[string]any{
		"connectorId": connectorIDOne,
		"availType":   statusOperative,
		"station":     stationHandle,
		"timeout":     defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("csmsEnqueuesChangeAvailability: want error for wrong connectorId, got nil")
	}
}

// ── stationRespondsToChangeAvailability tests ─────────────────────────────────

// Test_stationRespondsToChangeAvailability_sendsCALLRESULT verifies that the
// respond keyword sends a CALLRESULT frame with the given status.
func Test_stationRespondsToChangeAvailability_sendsCALLRESULT(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:csms_call:"+stationHandle+":ChangeAvailability", csmsUniqueID)

	fn := resolveFunc(t, patternStationRespondsAvail)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"status":  statusAccepted,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("stationRespondsToChangeAvailability: want nil, got %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("stationRespondsToChangeAvailability: want 1 sent frame, got %d", len(frames))
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
