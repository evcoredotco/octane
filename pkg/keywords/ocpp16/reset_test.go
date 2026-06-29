package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── csmsEnqueuesReset tests ───────────────────────────────────────────────────

// Test_csmsEnqueuesReset_stashesCallID verifies that the keyword parses the
// inbound Reset CALL and stashes the uniqueID for the response keyword.
func Test_csmsEnqueuesReset_stashesCallID(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionReset,
		map[string]any{"type": resetTypeSoft},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendReset)
	args := api.NewArgs(map[string]any{
		"resetType": resetTypeSoft,
		"station":   stationHandle,
		"timeout":   defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("csmsEnqueuesReset: want nil, got %v", err)
	}

	// csmsCallIDKey = "ocpp16:csms_call:{station}:Reset"
	val, ok := state.Pop("ocpp16:csms_call:" + stationHandle + ":Reset")
	if !ok {
		t.Fatal("csmsEnqueuesReset: want stashed uniqueID, got nothing")
	}

	uid, ok := val.(string)
	if !ok {
		t.Fatalf("stashed uniqueID: want string, got %T", val)
	}

	if uid != csmsUniqueID {
		t.Errorf("stashed uniqueID: want %q, got %q", csmsUniqueID, uid)
	}
}

// Test_csmsEnqueuesReset_errorOnWrongAction verifies that the keyword returns
// an error when the inbound CALL action is not "Reset".
func Test_csmsEnqueuesReset_errorOnWrongAction(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionBootNotification,
		map[string]any{"type": resetTypeSoft},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendReset)
	args := api.NewArgs(map[string]any{
		"resetType": resetTypeSoft,
		"station":   stationHandle,
		"timeout":   defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("csmsEnqueuesReset: want error for wrong action, got nil")
	}
}

// Test_csmsEnqueuesReset_errorOnWrongType verifies that the keyword returns an
// error when the Reset type field does not match the expected value.
func Test_csmsEnqueuesReset_errorOnWrongType(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionReset,
		map[string]any{"type": resetTypeHard},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendReset)
	args := api.NewArgs(map[string]any{
		"resetType": resetTypeSoft,
		"station":   stationHandle,
		"timeout":   defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("csmsEnqueuesReset: want error for wrong reset type, got nil")
	}
}

// ── stationRespondsToReset tests ──────────────────────────────────────────────

// Test_stationRespondsToReset_sendsCALLRESULT verifies that after enqueuing
// the Reset CALL, the respond keyword sends a CALLRESULT with the correct uniqueID
// and status.
func Test_stationRespondsToReset_sendsCALLRESULT(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionReset,
		map[string]any{"type": resetTypeSoft},
	})
	state := newState(t, station)

	enqueueFn := resolveFunc(t, patternCSMSSendReset)
	enqueueArgs := api.NewArgs(map[string]any{
		"resetType": resetTypeSoft,
		"station":   stationHandle,
		"timeout":   defaultTimeout,
	})
	if err := enqueueFn(context.Background(), state, enqueueArgs); err != nil {
		t.Fatalf("csmsEnqueuesReset: %v", err)
	}

	// Re-stash so respond keyword can pop it.
	state.Stash("ocpp16:csms_call:"+stationHandle+":Reset", csmsUniqueID)

	respondFn := resolveFunc(t, patternStationRespondsReset)
	respondArgs := api.NewArgs(map[string]any{
		"station": stationHandle,
		"status":  statusAccepted,
	})

	err := respondFn(context.Background(), state, respondArgs)
	if err != nil {
		t.Fatalf("stationRespondsToReset: want nil, got %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("stationRespondsToReset: want 1 sent frame, got %d", len(frames))
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

// Test_stationRespondsToReset_errorWithNoStash verifies that the respond keyword
// returns an error when no uniqueID is stashed.
func Test_stationRespondsToReset_errorWithNoStash(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	fn := resolveFunc(t, patternStationRespondsReset)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"status":  statusAccepted,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("stationRespondsToReset: want error without prior enqueue, got nil")
	}
}
