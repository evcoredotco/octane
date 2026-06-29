package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── sendMeterValues tests ─────────────────────────────────────────────────────

// Test_sendMeterValues_sendsCALLFrameWithConnectorAndValue verifies that the
// keyword sends a CALL frame with action="MeterValues", the correct connectorId,
// and a non-empty meterValue array.
func Test_sendMeterValues_sendsCALLFrameWithConnectorAndValue(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCallResult, "octane-metervalues-1",
		map[string]any{},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternSendMeterValues)
	args := api.NewArgs(map[string]any{
		"station":     stationHandle,
		"connectorId": connectorIDOne,
		"value":       sampledValue,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("sendMeterValues: unexpected error: %v", err)
	}

	frames := station.SentFrames()
	if len(frames) != 1 {
		t.Fatalf("sendMeterValues: want 1 sent frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame[0] != msgTypeCall {
		t.Errorf("frame[0]: want %v (CALL), got %v", msgTypeCall, frame[0])
	}

	if frame[2] != actionMeterValues {
		t.Errorf("frame[2]: want %q, got %v", actionMeterValues, frame[2])
	}

	payload, ok := frame[3].(map[string]any)
	if !ok {
		t.Fatalf("frame[3]: want map[string]any, got %T", frame[3])
	}

	if payload["connectorId"] != connectorIDOne {
		t.Errorf("payload.connectorId: want %d, got %v", connectorIDOne, payload["connectorId"])
	}

	meterValue, ok := payload["meterValue"].([]any)
	if !ok {
		t.Fatalf("payload.meterValue: want []any, got %T", payload["meterValue"])
	}

	if len(meterValue) == 0 {
		t.Error("payload.meterValue: want at least one entry, got empty")
	}
}

// ── csmsAcknowledgesMeterValues tests ─────────────────────────────────────────

// Test_csmsAcknowledgesMeterValues_passesOnEmptyConf verifies that the keyword
// passes when the CSMS returns an empty CALLRESULT (OCPP 1.6 §4.7).
func Test_csmsAcknowledgesMeterValues_passesOnEmptyConf(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-metervalues-1",
		map[string]any{},
	})

	sendFn := resolveFunc(t, patternSendMeterValues)
	sendArgs := api.NewArgs(map[string]any{
		"station":     stationHandle,
		"connectorId": connectorIDOne,
		"value":       sampledValue,
	})
	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendMeterValues: %v", err)
	}

	ackFn := resolveFunc(t, patternCSMSAcksMeter)
	ackArgs := api.NewArgs(map[string]any{"timeout": defaultTimeout})

	err := ackFn(context.Background(), state, ackArgs)
	if err != nil {
		t.Errorf("csmsAcknowledgesMeterValues: want nil, got %v", err)
	}
}
