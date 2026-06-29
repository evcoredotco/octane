package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── sendStopTransaction tests ─────────────────────────────────────────────────

// Test_sendStopTransaction_sendsCALLFrameWithAllFields verifies that the keyword
// sends a CALL frame with action="StopTransaction" and all required payload fields.
func Test_sendStopTransaction_sendsCALLFrameWithAllFields(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCallResult, "octane-stoptransaction-1",
		map[string]any{},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternSendStopTx)
	args := api.NewArgs(map[string]any{
		"station":       stationHandle,
		"transactionId": transactionIDPositive,
		"meterStop":     meterStopValue,
		"reason":        stopReasonNormal,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("sendStopTransaction: unexpected error: %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("sendStopTransaction: want 1 sent frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame[0] != msgTypeCall {
		t.Errorf("frame[0]: want %v (CALL), got %v", msgTypeCall, frame[0])
	}

	if frame[2] != actionStopTransaction {
		t.Errorf("frame[2]: want %q, got %v", actionStopTransaction, frame[2])
	}

	payload, ok := frame[3].(map[string]any)
	if !ok {
		t.Fatalf("frame[3]: want map[string]any, got %T", frame[3])
	}

	if payload["transactionId"] != transactionIDPositive {
		t.Errorf("payload.transactionId: want %d, got %v", transactionIDPositive, payload["transactionId"])
	}

	if payload["meterStop"] != meterStopValue {
		t.Errorf("payload.meterStop: want %d, got %v", meterStopValue, payload["meterStop"])
	}

	if payload["reason"] != stopReasonNormal {
		t.Errorf("payload.reason: want %q, got %v", stopReasonNormal, payload["reason"])
	}

	if _, exists := payload["timestamp"]; !exists {
		t.Error("payload missing timestamp field")
	}
}

// ── csmsAcceptsStopTransaction tests ──────────────────────────────────────────

// Test_csmsAcceptsStopTransaction_passesOnEmptyConf verifies that the keyword
// passes when the CSMS returns an empty CALLRESULT (OCPP 1.6 §6.47).
func Test_csmsAcceptsStopTransaction_passesOnEmptyConf(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-stoptransaction-1",
		map[string]any{},
	})

	sendFn := resolveFunc(t, patternSendStopTx)
	sendArgs := api.NewArgs(map[string]any{
		"station":       stationHandle,
		"transactionId": transactionIDPositive,
		"meterStop":     meterStopValue,
		"reason":        stopReasonNormal,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendStopTransaction: %v", err)
	}

	acceptFn := resolveFunc(t, patternCSMSAcceptsStop)
	acceptArgs := api.NewArgs(map[string]any{"timeout": defaultTimeout})

	err := acceptFn(context.Background(), state, acceptArgs)
	if err != nil {
		t.Errorf("csmsAcceptsStopTransaction: want nil, got %v", err)
	}
}

// Test_csmsAcceptsStopTransaction_errorWithNoPending verifies that the keyword
// returns an error when called without a preceding sendStopTransaction.
func Test_csmsAcceptsStopTransaction_errorWithNoPending(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	fn := resolveFunc(t, patternCSMSAcceptsStop)
	args := api.NewArgs(map[string]any{"timeout": defaultTimeout})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error("csmsAcceptsStopTransaction: want error without prior send, got nil")
	}
}
