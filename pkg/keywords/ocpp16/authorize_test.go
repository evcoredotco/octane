package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── sendAuthorize tests ───────────────────────────────────────────────────────

// Test_sendAuthorize_sendsCALLFrameWithIdTag verifies that the keyword sends a
// CALL frame with action="Authorize" and the correct idTag payload field.
func Test_sendAuthorize_sendsCALLFrameWithIdTag(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCallResult, "octane-authorize-1",
		map[string]any{
			"idTagInfo": map[string]any{"status": statusAccepted},
		},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternSendAuthorize)
	args := api.NewArgs(map[string]any{
		"station": stationHandle,
		"idTag":   idTagValue,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("sendAuthorize: unexpected error: %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("sendAuthorize: want 1 sent frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame[0] != msgTypeCall {
		t.Errorf("frame[0]: want %v (CALL), got %v", msgTypeCall, frame[0])
	}

	if frame[2] != actionAuthorize {
		t.Errorf("frame[2]: want %q, got %v", actionAuthorize, frame[2])
	}

	payload, ok := frame[3].(map[string]any)
	if !ok {
		t.Fatalf("frame[3]: want map[string]any, got %T", frame[3])
	}

	if payload["idTag"] != idTagValue {
		t.Errorf("payload.idTag: want %q, got %v", idTagValue, payload["idTag"])
	}
}

// ── csmsRespondsToAuthorize tests ─────────────────────────────────────────────

// Test_csmsRespondsToAuthorize_acceptsMatchingStatus verifies that the keyword
// passes when the CALLRESULT contains the expected idTagInfo.status.
func Test_csmsRespondsToAuthorize_acceptsMatchingStatus(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-authorize-1",
		map[string]any{
			"idTagInfo": map[string]any{"status": statusAccepted},
		},
	})

	sendFn := resolveFunc(t, patternSendAuthorize)
	sendArgs := api.NewArgs(map[string]any{
		"station": stationHandle,
		"idTag":   idTagValue,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendAuthorize: %v", err)
	}

	respondFn := resolveFunc(t, patternCSMSRespondsAuth)
	respondArgs := api.NewArgs(map[string]any{
		"status":  statusAccepted,
		"timeout": defaultTimeout,
	})

	err := respondFn(context.Background(), state, respondArgs)
	if err != nil {
		t.Errorf("csmsRespondsToAuthorize: want nil, got %v", err)
	}
}

// Test_csmsRespondsToAuthorize_errorOnRejectedStatus verifies that the keyword
// returns an error when the idTagInfo.status does not match the expected value.
func Test_csmsRespondsToAuthorize_errorOnRejectedStatus(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-authorize-1",
		map[string]any{
			"idTagInfo": map[string]any{"status": statusBlocked},
		},
	})

	sendFn := resolveFunc(t, patternSendAuthorize)
	sendArgs := api.NewArgs(map[string]any{
		"station": stationHandle,
		"idTag":   idTagValue,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendAuthorize: %v", err)
	}

	respondFn := resolveFunc(t, patternCSMSRespondsAuth)
	respondArgs := api.NewArgs(map[string]any{
		"status":  statusAccepted,
		"timeout": defaultTimeout,
	})

	err := respondFn(context.Background(), state, respondArgs)
	if err == nil {
		t.Error("csmsRespondsToAuthorize: want error for rejected status, got nil")
	}
}

// Test_csmsRespondsToAuthorize_errorOnMalformedIdTagInfo verifies that the
// keyword returns an error when idTagInfo is not an object.
func Test_csmsRespondsToAuthorize_errorOnMalformedIdTagInfo(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-authorize-1",
		map[string]any{
			"idTagInfo": "notanobject",
		},
	})

	sendFn := resolveFunc(t, patternSendAuthorize)
	sendArgs := api.NewArgs(map[string]any{
		"station": stationHandle,
		"idTag":   idTagValue,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendAuthorize: %v", err)
	}

	respondFn := resolveFunc(t, patternCSMSRespondsAuth)
	respondArgs := api.NewArgs(map[string]any{
		"status":  statusAccepted,
		"timeout": defaultTimeout,
	})

	err := respondFn(context.Background(), state, respondArgs)
	if err == nil {
		t.Error("csmsRespondsToAuthorize: want error for malformed idTagInfo, got nil")
	}
}
