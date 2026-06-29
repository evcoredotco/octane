package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── csmsEnqueuesClearCache tests ──────────────────────────────────────────────

// Test_csmsEnqueuesClearCache_stashesCallID verifies that the keyword parses
// the inbound ClearCache CALL and stashes the uniqueID.
func Test_csmsEnqueuesClearCache_stashesCallID(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCall, csmsUniqueID, actionClearCache,
		map[string]any{},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSSendClearCache)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"timeout": defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("csmsEnqueuesClearCache: want nil, got %v", err)
	}

	val, ok := state.Pop("ocpp16:csms_call:" + stationHandle + ":ClearCache")
	if !ok {
		t.Fatal("csmsEnqueuesClearCache: want stashed uniqueID, got nothing")
	}

	uid, ok := val.(string)
	if !ok {
		t.Fatalf("stashed uniqueID: want string, got %T", val)
	}

	if uid != csmsUniqueID {
		t.Errorf("stashed uniqueID: want %q, got %q", csmsUniqueID, uid)
	}
}

// ── stationRespondsToClearCache tests ─────────────────────────────────────────

// Test_stationRespondsToClearCache_sendsCALLRESULT verifies that the respond
// keyword sends a CALLRESULT with status="Accepted".
func Test_stationRespondsToClearCache_sendsCALLRESULT(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:csms_call:"+stationHandle+":ClearCache", csmsUniqueID)

	fn := resolveFunc(t, patternStationRespondsClearCache)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"status":  statusAccepted,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("stationRespondsToClearCache: want nil, got %v", err)
	}

	respPayload := requireSentCallResultPayload(
		t,
		station,
		"stationRespondsToClearCache",
	)
	if respPayload["status"] != statusAccepted {
		t.Errorf(
			"payload.status: want %q, got %v",
			statusAccepted,
			respPayload["status"],
		)
	}
}
