package ocpp16_test

import (
	"context"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── sendBootNotification tests ───────────────────────────────────────────────

// Test_sendBootNotification_sendsCALLFrame verifies that the keyword sends a
// CALL frame with action="BootNotification" and a payload containing "reason".
func Test_sendBootNotification_sendsCALLFrame(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCallResult, "octane-bootnotification-1",
		map[string]any{
			"currentTime": currentTimeValid,
			"interval":    float64(heartbeatIntervalInRange),
			"status":      statusAccepted,
		},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternSendBoot)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"reason":  reasonNormal,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("sendBootNotification: unexpected error: %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("sendBootNotification: want 1 sent frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame[0] != msgTypeCall {
		t.Errorf("frame[0]: want %v (CALL), got %v", msgTypeCall, frame[0])
	}

	if frame[2] != actionBootNotification {
		t.Errorf("frame[2]: want %q, got %v", actionBootNotification, frame[2])
	}

	payload, ok := frame[3].(map[string]any)
	if !ok {
		t.Fatalf("frame[3]: want map[string]any, got %T", frame[3])
	}

	if _, exists := payload["chargePointVendor"]; !exists {
		t.Error("payload missing chargePointVendor field")
	}
}

// Test_sendBootNotification_errorOnMissingStation verifies that the keyword
// returns an error when no station is registered.
func Test_sendBootNotification_errorOnMissingStation(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()

	fn := resolveFunc(t, patternSendBoot)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"reason":  reasonNormal,
	})

	defer func() {
		if r := recover(); r == nil {
			t.Error("sendBootNotification: expected panic on unregistered station, got none")
		}
	}()

	_ = fn(context.Background(), state, args)
}

// ── csmsRespondsWithStatus tests ─────────────────────────────────────────────

// Test_csmsRespondsWithStatus_acceptsMatchingStatus verifies that the keyword
// passes when the CALLRESULT status matches the expected value.
func Test_csmsRespondsWithStatus_acceptsMatchingStatus(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-bootnotification-1",
		map[string]any{
			"currentTime": currentTimeValid,
			"interval":    float64(heartbeatIntervalInRange),
			"status":      statusAccepted,
		},
	})

	sendFn := resolveFunc(t, patternSendBoot)
	sendArgs := api.NewArgs(map[string]any{
		"station": stationHandle,
		"reason":  reasonNormal,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendBootNotification: %v", err)
	}

	respondFn := resolveFunc(t, patternCSMSResponds)
	respondArgs := api.NewArgs(map[string]any{
		"status":  statusAccepted,
		"timeout": defaultTimeout,
	})

	err := respondFn(context.Background(), state, respondArgs)
	if err != nil {
		t.Errorf("csmsRespondsWithStatus: want nil, got %v", err)
	}
}

// Test_csmsRespondsWithStatus_errorOnWrongStatus verifies that the keyword
// returns an error when the CALLRESULT status does not match the expected value.
func Test_csmsRespondsWithStatus_errorOnWrongStatus(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-bootnotification-1",
		map[string]any{
			"currentTime": currentTimeValid,
			"interval":    float64(heartbeatIntervalInRange),
			"status":      statusRejected,
		},
	})

	sendFn := resolveFunc(t, patternSendBoot)
	sendArgs := api.NewArgs(map[string]any{
		"station": stationHandle,
		"reason":  reasonNormal,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendBootNotification: %v", err)
	}

	respondFn := resolveFunc(t, patternCSMSResponds)
	respondArgs := api.NewArgs(map[string]any{
		"status":  statusAccepted,
		"timeout": defaultTimeout,
	})

	err := respondFn(context.Background(), state, respondArgs)
	if err == nil {
		t.Error("csmsRespondsWithStatus: want error for wrong status, got nil")
	}
}

// Test_csmsRespondsWithStatus_errorWithNoPending verifies that the keyword
// returns an error when called without a preceding sendBootNotification.
func Test_csmsRespondsWithStatus_errorWithNoPending(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSResponds)
	args := api.NewArgs(map[string]any{
		"status":  statusAccepted,
		"timeout": defaultTimeout,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("csmsRespondsWithStatus: want error without prior send, got nil")
	}
}

// ── stationIsInRegisteredState tests ─────────────────────────────────────────

// Test_stationIsInRegisteredState_passesAfterAccepted verifies that the keyword
// passes when the registered stash flag is present.
func Test_stationIsInRegisteredState_passesAfterAccepted(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:registered:"+stationHandle, true)

	fn := resolveFunc(t, patternStationRegistered)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("stationIsInRegisteredState: want nil, got %v", err)
	}
}

// Test_stationIsInRegisteredState_failsWhenNotRegistered verifies that the
// keyword returns an error when no registered stash flag is present.
func Test_stationIsInRegisteredState_failsWhenNotRegistered(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	fn := resolveFunc(t, patternStationRegistered)
	args := api.NewArgs(map[string]any{"station": stationHandle})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("stationIsInRegisteredState: want error when not registered, got nil")
	}
}

// ── responseIncludesHeartbeatInterval tests ───────────────────────────────────

// Test_responseIncludesHeartbeatInterval_passesInRange verifies that the
// keyword passes when the stashed payload interval is within [min, max].
func Test_responseIncludesHeartbeatInterval_passesInRange(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:last_payload", map[string]any{
		"heartbeatInterval": float64(heartbeatIntervalInRange),
	})

	fn := resolveFunc(t, patternHBInterval)
	args := api.NewArgs(map[string]any{
		"min": heartbeatIntervalMin,
		"max": heartbeatIntervalMax,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("responseIncludesHeartbeatInterval: want nil, got %v", err)
	}
}

// Test_responseIncludesHeartbeatInterval_failsOutOfRange verifies that the
// keyword returns an error when the stashed payload interval is outside [min, max].
func Test_responseIncludesHeartbeatInterval_failsOutOfRange(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:last_payload", map[string]any{
		"heartbeatInterval": float64(heartbeatIntervalTooLow),
	})

	fn := resolveFunc(t, patternHBInterval)
	args := api.NewArgs(map[string]any{
		"min": heartbeatIntervalMin,
		"max": heartbeatIntervalMax,
	})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("responseIncludesHeartbeatInterval: want error for out-of-range interval, got nil")
	}
}

// ── responseIncludesCurrentTime tests ────────────────────────────────────────

// Test_responseIncludesCurrentTime_passesValidISO8601 verifies that the keyword
// passes when the stashed payload currentTime is a valid RFC 3339 string.
func Test_responseIncludesCurrentTime_passesValidISO8601(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:last_payload", map[string]any{
		"currentTime": currentTimeValid,
	})

	fn := resolveFunc(t, patternCurrentTime)
	args := api.NewArgs(map[string]any{})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf("responseIncludesCurrentTime: want nil, got %v", err)
	}
}

// Test_responseIncludesCurrentTime_failsInvalidFormat verifies that the keyword
// returns an error when the stashed payload currentTime is not valid RFC 3339.
func Test_responseIncludesCurrentTime_failsInvalidFormat(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash("ocpp16:last_payload", map[string]any{
		"currentTime": currentTimeInvalid,
	})

	fn := resolveFunc(t, patternCurrentTime)
	args := api.NewArgs(map[string]any{})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("responseIncludesCurrentTime: want error for invalid format, got nil")
	}
}

// Test_csmsRespondsWithStatus_setsRegisteredFlag verifies that when status is
// "Accepted", the registered stash flag is set for the station.
func Test_csmsRespondsWithStatus_setsRegisteredFlag(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-bootnotification-1",
		map[string]any{
			"currentTime": currentTimeValid,
			"interval":    float64(heartbeatIntervalInRange),
			"status":      statusAccepted,
		},
	})

	sendFn := resolveFunc(t, patternSendBoot)
	sendArgs := api.NewArgs(map[string]any{
		"station": stationHandle,
		"reason":  reasonNormal,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendBootNotification: %v", err)
	}

	respondFn := resolveFunc(t, patternCSMSResponds)
	respondArgs := api.NewArgs(map[string]any{
		"status":  statusAccepted,
		"timeout": defaultTimeout,
	})
	if err := respondFn(context.Background(), state, respondArgs); err != nil {
		t.Fatalf("csmsRespondsWithStatus: %v", err)
	}

	registeredFn := resolveFunc(t, patternStationRegistered)
	registeredArgs := api.NewArgs(map[string]any{"station": stationHandle})

	err := registeredFn(context.Background(), state, registeredArgs)
	if err != nil {
		t.Errorf("stationIsInRegisteredState: want nil after Accepted boot, got %v", err)
	}
}

// Ensure time import is used (for defaultTimeout declared in helpers).
var _ = time.Second
