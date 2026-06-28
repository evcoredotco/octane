package ocpp16

import (
	"context"
	"errors"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// sendStatusNotification implements:
//
//	station {station:string} sends StatusNotification for connector {connectorId:int} with status {status:string}
//
// It sends a minimal OCPP 1.6 StatusNotification.req and stashes both the
// pending correlation info and the connector's new status for the subsequent
// assertion step.
func sendStatusNotification(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	connectorID := args.Int(fieldConnectorID)
	status := args.String("status")

	msgID := nextMsgID(state, station, "StatusNotification")

	payload := map[string]any{
		fieldConnectorID: connectorID,
		"errorCode":      "NoError",
		fieldStatus:      status,
		fieldTimestamp:   state.Now().Format(iso8601SecondFormat),
	}

	if err := sendCall(ctx, state, station, msgID, "StatusNotification", payload); err != nil {
		return err
	}

	state.Stash(pendingKey, &pendingInfo{
		station:         station,
		msgID:           msgID,
		action:          "StatusNotification",
		connectorID:     connectorID,
		connectorStatus: status,
	})

	// Record the connector state immediately; the CSMS only acknowledges,
	// it does not specify the resulting state.
	state.Stash(connectorKey(station, connectorID), status)

	state.Logf(
		"station %q sent StatusNotification connector=%d status=%q",
		station, connectorID, status,
	)

	return nil
}

// csmsAcknowledgesStatus implements:
//
//	the CSMS acknowledges the status within {timeout:duration}
//
// It waits for the CALLRESULT to the preceding StatusNotification CALL.
// OCPP 1.6 StatusNotification.conf has an empty payload, so no field
// validation is performed beyond successfully parsing the frame.
func csmsAcknowledgesStatus(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	timeout := args.Duration("timeout")

	info, ok := popPending(state)
	if !ok {
		return errors.New("ocpp16: no pending StatusNotification; call sendStatusNotification first")
	}

	payload, err := expectResult(ctx, state, info.station, timeout)
	if err != nil {
		return err
	}

	state.Stash(lastPayloadKey, payload)

	state.Logf(
		"station %q received StatusNotification.conf (connector=%d)",
		info.station, info.connectorID,
	)

	return nil
}

// connectorIsInState implements:
//
//	connector {connectorId:int} of station {station:string} is in state {state:string}
//
// It checks the connector status recorded when the preceding
// sendStatusNotification step ran.
func connectorIsInState(
	_ context.Context,
	s api.State,
	args api.Args,
) error {
	connectorID := args.Int("connectorId")
	station := args.String("station")
	expectedState := args.String("state")

	key := connectorKey(station, connectorID)

	val, ok := s.Pop(key)
	if !ok {
		return fmt.Errorf(
			"ocpp16: connector %d of station %q: no status recorded; "+
				"call sendStatusNotification first",
			connectorID, station,
		)
	}

	// Re-stash so subsequent checks in the same scenario still see the value.
	s.Stash(key, val)

	gotState, ok := val.(string)
	if !ok {
		return fmt.Errorf(
			"ocpp16: connector %d of station %q: state stash has unexpected type %T",
			connectorID, station, val,
		)
	}

	if gotState != expectedState {
		return fmt.Errorf(
			"ocpp16: connector %d of station %q: want state %q, got %q",
			connectorID, station, expectedState, gotState,
		)
	}

	return nil
}
