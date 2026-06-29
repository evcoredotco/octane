package ocpp16

import (
	"context"
	"errors"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// sendStartTransaction implements:
//
//	station {station:string} starts a transaction on connector {connectorId:int} with idTag {idTag:string} and meterStart {meterStart:int}
//
// It sends an OCPP 1.6 StartTransaction.req with the given connector,
// idTag, and meterStart values, and stashes the pending correlation info.
func sendStartTransaction(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	connectorID := args.Int(fieldConnectorID)
	idTag := args.String("idTag")
	meterStart := args.Int("meterStart")

	msgID := nextMsgID(state, station, actionStartTransaction)

	payload := map[string]any{
		fieldConnectorID: connectorID,
		fieldIDTag:       idTag,
		"meterStart":     meterStart,
		fieldTimestamp:   state.Now().Format(iso8601SecondFormat),
	}

	if err := sendCall(ctx, state, station, msgID, actionStartTransaction, payload); err != nil {
		return err
	}

	state.Stash(pendingKey, &pendingInfo{
		station: station,
		msgID:   msgID,
		action:  actionStartTransaction,
	})

	state.Logf(
		"station %q sent StartTransaction (connector=%d, idTag=%q, meterStart=%d, msgID=%s)",
		station,
		connectorID,
		idTag,
		meterStart,
		msgID,
	)

	return nil
}

// csmsRespondsToStartTransaction implements:
//
//	the CSMS responds to StartTransaction with idTagInfo.status {status:string} within {timeout:duration}
//
// It waits for the StartTransaction.conf CALLRESULT, validates the nested
// idTagInfo.status field, and stashes the payload for subsequent
// assertion steps (e.g. transactionId check).
func csmsRespondsToStartTransaction(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	expectedStatus := args.String("status")
	timeout := args.Duration("timeout")

	info, ok := popPending(state)
	if !ok {
		return errors.New(
			"ocpp16: no pending StartTransaction; call sendStartTransaction first",
		)
	}

	payload, err := expectResult(ctx, state, info.station, timeout)
	if err != nil {
		return err
	}

	state.Stash(lastPayloadKey, payload)

	gotStatus, err := startTransactionIDTagStatus(payload)
	if err != nil {
		return err
	}

	if gotStatus != expectedStatus {
		return fmt.Errorf(
			"ocpp16: station %q: StartTransaction.conf idTagInfo.status: want %q, got %q",
			info.station,
			expectedStatus,
			gotStatus,
		)
	}

	stashPositiveTransactionID(state, payload)

	state.Logf(
		"station %q received StartTransaction.conf idTagInfo.status=%q",
		info.station, gotStatus,
	)

	return nil
}

func startTransactionIDTagStatus(payload map[string]any) (string, error) {
	rawTagInfo, exists := payload["idTagInfo"]
	if !exists {
		return "", errors.New(
			"ocpp16: StartTransaction.conf payload missing idTagInfo field",
		)
	}

	tagInfo, ok := rawTagInfo.(map[string]any)
	if !ok {
		return "", fmt.Errorf(
			"ocpp16: StartTransaction.conf idTagInfo has unexpected type %T (want object)",
			rawTagInfo,
		)
	}

	return payloadString(
		tagInfo,
		fieldStatus,
		"StartTransaction.conf idTagInfo",
	)
}

func stashPositiveTransactionID(state api.State, payload map[string]any) {
	rawTxID, exists := payload["transactionId"]
	if !exists {
		return
	}

	txID, ok := rawTxID.(float64)
	if ok && txID > positiveTransactionIDBoundary {
		state.Stash(transactionIDKey, int(txID))
	}
}

// startTransactionAssignsPositiveTransactionID implements:
//
//	the StartTransaction response assigns a positive transactionId
//
// It inspects the stashed StartTransaction.conf payload and validates that
// the transactionId field is a positive integer.
func startTransactionAssignsPositiveTransactionID(
	_ context.Context,
	state api.State,
	_ api.Args,
) error {
	payload, ok := peekPayload(state)
	if !ok {
		return errors.New(
			"ocpp16: no StartTransaction.conf payload stashed; call csmsRespondsToStartTransaction first",
		)
	}

	rawID, exists := payload["transactionId"]
	if !exists {
		return errors.New(
			"ocpp16: StartTransaction.conf payload missing transactionId field",
		)
	}

	txID, ok := rawID.(float64)
	if !ok {
		return fmt.Errorf(
			"ocpp16: StartTransaction.conf transactionId has unexpected type %T (want number)",
			rawID,
		)
	}

	if txID <= positiveTransactionIDBoundary {
		return fmt.Errorf(
			"ocpp16: StartTransaction.conf transactionId must be positive, got %v",
			txID,
		)
	}

	return nil
}
