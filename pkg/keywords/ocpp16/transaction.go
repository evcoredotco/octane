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
	connectorID := args.Int("connectorId")
	idTag := args.String("idTag")
	meterStart := args.Int("meterStart")

	msgID := nextMsgID(state, station, "StartTransaction")

	payload := map[string]any{
		"connectorId": connectorID,
		"idTag":       idTag,
		"meterStart":  meterStart,
		"timestamp":   state.Now().Format("2006-01-02T15:04:05Z"),
	}

	if err := sendCall(ctx, state, station, msgID, "StartTransaction", payload); err != nil {
		return err
	}

	state.Stash(pendingKey, &pendingInfo{
		station: station,
		msgID:   msgID,
		action:  "StartTransaction",
	})

	state.Logf(
		"station %q sent StartTransaction (connector=%d, idTag=%q, meterStart=%d, msgID=%s)",
		station, connectorID, idTag, meterStart, msgID,
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
		return errors.New("ocpp16: no pending StartTransaction; call sendStartTransaction first")
	}

	_, payload, err := expectResult(ctx, state, info.station, timeout)
	if err != nil {
		return err
	}

	state.Stash(lastPayloadKey, payload)

	rawTagInfo, exists := payload["idTagInfo"]
	if !exists {
		return errors.New("ocpp16: StartTransaction.conf payload missing idTagInfo field")
	}

	tagInfo, ok := rawTagInfo.(map[string]any)
	if !ok {
		return fmt.Errorf(
			"ocpp16: StartTransaction.conf idTagInfo has unexpected type %T (want object)",
			rawTagInfo,
		)
	}

	gotStatus, _ := tagInfo["status"].(string)
	if gotStatus != expectedStatus {
		return fmt.Errorf(
			"ocpp16: station %q: StartTransaction.conf idTagInfo.status: want %q, got %q",
			info.station, expectedStatus, gotStatus,
		)
	}

	if rawTxID, exists := payload["transactionId"]; exists {
		if txID, ok := rawTxID.(float64); ok && txID > 0 {
			state.Stash(transactionIDKey, int(txID))
		}
	}

	state.Logf(
		"station %q received StartTransaction.conf idTagInfo.status=%q",
		info.station, gotStatus,
	)

	return nil
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
		return errors.New("ocpp16: no StartTransaction.conf payload stashed; call csmsRespondsToStartTransaction first")
	}

	rawID, exists := payload["transactionId"]
	if !exists {
		return errors.New("ocpp16: StartTransaction.conf payload missing transactionId field")
	}

	txID, ok := rawID.(float64)
	if !ok {
		return fmt.Errorf(
			"ocpp16: StartTransaction.conf transactionId has unexpected type %T (want number)",
			rawID,
		)
	}

	if txID <= 0 {
		return fmt.Errorf(
			"ocpp16: StartTransaction.conf transactionId must be positive, got %v",
			txID,
		)
	}

	return nil
}
