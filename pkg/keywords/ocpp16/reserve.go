package ocpp16

import (
	"context"
	"errors"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/wire"
)

// csmsEnqueuesReserveNow implements:
//
//	the CSMS sends ReserveNow with connectorId {connectorId:int} and idTag {idTag:string} to station {station:string} within {timeout:duration}
//
// It waits for an inbound CALL frame, validates that it is a ReserveNow
// request with the expected connectorId and idTag, and stashes the
// uniqueID for the subsequent response keyword.
func csmsEnqueuesReserveNow(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	connectorID := args.Int("connectorId")
	idTag := args.String("idTag")
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, callPayload, err := expectCSMSCall(ctx, state, station, actionReserveNow, timeout)
	if err != nil {
		return err
	}

	gotConnector, err := payloadNumber(callPayload, fieldConnectorID, actionReserveNow)
	if err != nil {
		return err
	}

	if int(gotConnector) != connectorID {
		return fmt.Errorf(
			"ocpp16: station %q: ReserveNow connectorId: want %d, got %d",
			station, connectorID, int(gotConnector),
		)
	}

	gotIDTag, err := payloadString(callPayload, fieldIDTag, actionReserveNow)
	if err != nil {
		return err
	}

	if gotIDTag != idTag {
		return fmt.Errorf(
			"ocpp16: station %q: ReserveNow idTag: want %q, got %q",
			station, idTag, gotIDTag,
		)
	}

	state.Stash(reserveCallIDKey(station), uniqueID)

	state.Logf(
		"station %q received ReserveNow CALL (uniqueID=%s, connector=%d, idTag=%q)",
		station, uniqueID, connectorID, idTag,
	)

	return nil
}

// stationRespondsWithReserveNow implements:
//
//	station {station:string} responds with ReserveNow.conf status {status:string}
//
// It sends a CALLRESULT with the given status for the pending ReserveNow
// CALL, and stashes the correlation info for the subsequent CSMS-acceptance
// keyword.
func stationRespondsWithReserveNow(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	status := args.String("status")

	val, ok := state.Pop(reserveCallIDKey(station))
	if !ok {
		return fmt.Errorf(
			"ocpp16: station %q: no ReserveNow uniqueID stashed; call csmsEnqueuesReserveNow first",
			station,
		)
	}

	uniqueID, ok := val.(string)
	if !ok {
		return fmt.Errorf("ocpp16: station %q: ReserveNow uniqueID stash has unexpected type %T", station, val)
	}

	sv, err := state.Station(station)
	if err != nil {
		return fmt.Errorf(stationNotConnectedFormat, station, err)
	}

	frame := []any{
		float64(wire.MessageTypeResult),
		uniqueID,
		map[string]any{fieldStatus: status},
	}

	if err := sv.Send(ctx, frame); err != nil {
		return fmt.Errorf("ocpp16: station %q: send ReserveNow.conf: %w", station, err)
	}

	state.Stash(reserveWaitingKey, &reserveWaiting{
		station:  station,
		uniqueID: uniqueID,
	})

	state.Logf(
		"station %q sent ReserveNow.conf (uniqueID=%s, status=%q)",
		station, uniqueID, status,
	)

	return nil
}

// csmsAcceptsReserveResponse implements:
//
//	the CSMS accepts the response without error within {timeout:duration}
//
// It waits for the given duration to confirm no CALLERROR arrives from the
// CSMS. A timeout (no frame received) is the passing case; a CALLERROR
// matching the pending ReserveNow uniqueID is a failure.
func csmsAcceptsReserveResponse(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	timeout := args.Duration("timeout")

	val, ok := state.Pop(reserveWaitingKey)
	if !ok {
		return errors.New("ocpp16: no pending ReserveNow.conf; call stationRespondsWithReserveNow first")
	}

	waiting, ok := val.(*reserveWaiting)
	if !ok || waiting == nil {
		return errors.New("ocpp16: reserveWaiting stash has unexpected type")
	}

	sv, err := state.Station(waiting.station)
	if err != nil {
		return fmt.Errorf(stationNotConnectedFormat, waiting.station, err)
	}

	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	frame, err := sv.Expect(subCtx)
	if err != nil {
		// Timeout means the CSMS sent nothing — the reservation response was accepted.
		return handleReserveExpectError(state, waiting.station, err)
	}

	// If the frame is a CALLERROR matching the pending uniqueID, the CSMS rejected
	// the reservation response.
	errFrame, parseErr := wire.ParseError(frame)
	if parseErr == nil && errFrame.UniqueID == waiting.uniqueID {
		return fmt.Errorf(
			"ocpp16: station %q: CSMS rejected ReserveNow.conf with CALLERROR %s: %s",
			waiting.station, errFrame.ErrorCode, errFrame.ErrorDescription,
		)
	}

	// Any other frame is not an error response to our CALLRESULT.
	state.Logf(
		"station %q: received non-CALLERROR frame after ReserveNow.conf (CSMS accepted)",
		waiting.station,
	)

	return nil
}

func handleReserveExpectError(state api.State, station string, err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		state.Logf(
			"station %q: no CALLERROR within timeout (ReserveNow.conf accepted)",
			station,
		)

		return nil
	}

	return fmt.Errorf("ocpp16: station %q: expect after ReserveNow.conf: %w", station, err)
}
