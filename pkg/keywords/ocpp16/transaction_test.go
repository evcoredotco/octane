package ocpp16_test

import (
	"context"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// ── sendStartTransaction tests ────────────────────────────────────────────────

// Test_sendStartTransaction_sendsCALLFrameWithAllFields verifies that the keyword
// sends a CALL frame with action="StartTransaction" and all required payload fields.
func Test_sendStartTransaction_sendsCALLFrameWithAllFields(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	station.QueueFrame([]any{
		msgTypeCallResult, "octane-starttransaction-1",
		map[string]any{
			"transactionId": float64(transactionIDPositive),
			"idTagInfo":     map[string]any{"status": statusAccepted},
		},
	})
	state := newState(t, station)

	fn := resolveFunc(t, patternSendStartTx)
	args := api.NewArgs(map[string]any{
		"station":     stationHandle,
		"connectorId": connectorIDOne,
		"idTag":       idTagValue,
		"meterStart":  meterStartValue,
	})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Fatalf("sendStartTransaction: unexpected error: %v", err)
	}

	payload := requireSentCallPayload(
		t,
		station,
		"sendStartTransaction",
		actionStartTransaction,
	)
	if payload["connectorId"] != connectorIDOne {
		t.Errorf(
			"payload.connectorId: want %d, got %v",
			connectorIDOne,
			payload["connectorId"],
		)
	}

	if payload["idTag"] != idTagValue {
		t.Errorf("payload.idTag: want %q, got %v", idTagValue, payload["idTag"])
	}

	if payload["meterStart"] != meterStartValue {
		t.Errorf(
			"payload.meterStart: want %d, got %v",
			meterStartValue,
			payload["meterStart"],
		)
	}

	if _, exists := payload["timestamp"]; !exists {
		t.Error("payload missing timestamp field")
	}
}

// ── csmsRespondsToStartTransaction tests ──────────────────────────────────────

// Test_csmsRespondsToStartTransaction_acceptsAccepted verifies that the keyword
// passes when the CALLRESULT contains a positive transactionId and accepted status.
func Test_csmsRespondsToStartTransaction_acceptsAccepted(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-starttransaction-1",
		map[string]any{
			"transactionId": float64(transactionIDPositive),
			"idTagInfo":     map[string]any{"status": statusAccepted},
		},
	})

	sendFn := resolveFunc(t, patternSendStartTx)
	sendArgs := api.NewArgs(map[string]any{
		"station":     stationHandle,
		"connectorId": connectorIDOne,
		"idTag":       idTagValue,
		"meterStart":  meterStartValue,
	})

	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendStartTransaction: %v", err)
	}

	respondFn := resolveFunc(t, patternCSMSRespondsStartTx)
	respondArgs := api.NewArgs(map[string]any{
		"status":  statusAccepted,
		"timeout": defaultTimeout,
	})

	err := respondFn(context.Background(), state, respondArgs)
	if err != nil {
		t.Errorf("csmsRespondsToStartTransaction: want nil, got %v", err)
	}
}

// Test_csmsRespondsToStartTransaction_stashesTransactionId verifies that after
// a successful response, the transactionId is stashed under transactionIDKey.
func Test_csmsRespondsToStartTransaction_stashesTransactionId(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)

	station.QueueFrame([]any{
		msgTypeCallResult, "octane-starttransaction-1",
		map[string]any{
			"transactionId": float64(transactionIDPositive),
			"idTagInfo":     map[string]any{"status": statusAccepted},
		},
	})

	sendFn := resolveFunc(t, patternSendStartTx)
	sendArgs := api.NewArgs(map[string]any{
		"station":     stationHandle,
		"connectorId": connectorIDOne,
		"idTag":       idTagValue,
		"meterStart":  meterStartValue,
	})

	if err := sendFn(context.Background(), state, sendArgs); err != nil {
		t.Fatalf("sendStartTransaction: %v", err)
	}

	respondFn := resolveFunc(t, patternCSMSRespondsStartTx)
	respondArgs := api.NewArgs(map[string]any{
		"status":  statusAccepted,
		"timeout": defaultTimeout,
	})

	if err := respondFn(context.Background(), state, respondArgs); err != nil {
		t.Fatalf("csmsRespondsToStartTransaction: %v", err)
	}

	// transactionIDKey = "ocpp16:transaction_id"
	val, ok := state.Pop("ocpp16:transaction_id")
	if !ok {
		t.Fatal("transactionId: want stashed value, got nothing")
	}

	txID, ok := val.(int)
	if !ok {
		t.Fatalf("transactionId stash: want int, got %T", val)
	}

	if txID != transactionIDPositive {
		t.Errorf(
			"transactionId stash: want %d, got %d",
			transactionIDPositive,
			txID,
		)
	}
}

// ── startTransactionAssignsPositiveTransactionID tests ───────────────────────

// Test_startTransactionAssignsPositiveTransactionID_passesPositive verifies
// that the keyword passes when the stashed payload has a positive transactionId.
func Test_startTransactionAssignsPositiveTransactionID_passesPositive(
	t *testing.T,
) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash(lastPayloadKeyTest, map[string]any{
		"transactionId": float64(transactionIDPositive),
		"idTagInfo":     map[string]any{"status": statusAccepted},
	})

	fn := resolveFunc(t, patternPositiveTxID)
	args := api.NewArgs(map[string]any{})

	err := fn(context.Background(), state, args)
	if err != nil {
		t.Errorf(
			"startTransactionAssignsPositiveTransactionID: want nil, got %v",
			err,
		)
	}
}

// Test_startTransactionAssignsPositiveTransactionID_failsZero verifies that
// the keyword returns an error when the transactionId is zero.
func Test_startTransactionAssignsPositiveTransactionID_failsZero(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash(lastPayloadKeyTest, map[string]any{
		"transactionId": float64(transactionIDZero),
		"idTagInfo":     map[string]any{"status": statusAccepted},
	})

	fn := resolveFunc(t, patternPositiveTxID)
	args := api.NewArgs(map[string]any{})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error(
			"startTransactionAssignsPositiveTransactionID: want error for zero id, got nil",
		)
	}
}

// Test_startTransactionAssignsPositiveTransactionID_failsMissing verifies that
// the keyword returns an error when transactionId is absent from the payload.
func Test_startTransactionAssignsPositiveTransactionID_failsMissing(
	t *testing.T,
) {
	t.Parallel()

	station := mock.NewMockStation()
	state := newState(t, station)
	state.Stash(lastPayloadKeyTest, map[string]any{
		"idTagInfo": map[string]any{"status": statusAccepted},
	})

	fn := resolveFunc(t, patternPositiveTxID)
	args := api.NewArgs(map[string]any{})

	err := fn(context.Background(), state, args)
	if err == nil {
		t.Error(
			"startTransactionAssignsPositiveTransactionID: want error for missing id, got nil",
		)
	}
}
