package ocpp16

import (
	"context"
	"errors"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// sendStopTransaction implements:
//
//	station {station:string} stops transaction {transactionId:int} with meterStop {meterStop:int} and reason {reason:string}
//
// It sends an OCPP 1.6 StopTransaction.req. If a transactionId was
// stashed by csmsRespondsToStartTransaction, that value takes precedence
// over the step argument (allowing the story to pass 0 as a placeholder).
func sendStopTransaction(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	transactionID := args.Int("transactionId")
	meterStop := args.Int("meterStop")
	reason := args.String("reason")

	if val, ok := state.Pop(transactionIDKey); ok {
		if id, ok := val.(int); ok && id > positiveTransactionIDBoundary {
			transactionID = id
		}
	}

	msgID := nextMsgID(state, station, "StopTransaction")

	payload := map[string]any{
		"transactionId": transactionID,
		"meterStop":     meterStop,
		fieldTimestamp:  state.Now().Format(iso8601SecondFormat),
		"reason":        reason,
	}

	if err := sendCall(ctx, state, station, msgID, "StopTransaction", payload); err != nil {
		return err
	}

	state.Stash(pendingKey, &pendingInfo{
		station: station,
		msgID:   msgID,
		action:  "StopTransaction",
	})

	state.Logf(
		"station %q sent StopTransaction (transactionId=%d, meterStop=%d, reason=%q, msgID=%s)",
		station, transactionID, meterStop, reason, msgID,
	)

	return nil
}

// csmsAcceptsStopTransaction implements:
//
//	the CSMS accepts StopTransaction within {timeout:duration}
//
// It waits for the StopTransaction.conf CALLRESULT. Per OCPP 1.6 §6.47
// the confirmation payload is empty; success means the frame arrived
// without error.
func csmsAcceptsStopTransaction(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	timeout := args.Duration("timeout")

	info, ok := popPending(state)
	if !ok {
		return errors.New("ocpp16: no pending StopTransaction; call sendStopTransaction first")
	}

	payload, err := expectResult(ctx, state, info.station, timeout)
	if err != nil {
		return err
	}

	state.Stash(lastPayloadKey, payload)

	state.Logf(
		"station %q received StopTransaction.conf (accepted)",
		info.station,
	)

	return nil
}
