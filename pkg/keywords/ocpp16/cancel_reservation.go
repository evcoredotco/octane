package ocpp16

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// csmsEnqueuesCancelReservation implements:
//
//	the CSMS sends CancelReservation with reservationId {reservationId:int} to station {station:string} within {timeout:duration}
//
// It waits for an inbound CancelReservation CALL, validates the
// reservationId field, and stashes the uniqueID for the subsequent
// response keyword.
func csmsEnqueuesCancelReservation(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	reservationID := args.Int("reservationId")
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, payload, err := expectCSMSCall(
		ctx,
		state,
		station,
		actionCancelReservation,
		timeout,
	)
	if err != nil {
		return err
	}

	gotResID, err := payloadNumber(
		payload,
		"reservationId",
		actionCancelReservation,
	)
	if err != nil {
		return err
	}

	if int(gotResID) != reservationID {
		return fmt.Errorf(
			"ocpp16: station %q: CancelReservation reservationId: want %d, got %d",
			station,
			reservationID,
			int(gotResID),
		)
	}

	state.Stash(csmsCallIDKey(station, actionCancelReservation), uniqueID)

	state.Logf(
		"station %q received CancelReservation CALL (uniqueID=%s, reservationId=%d)",
		station,
		uniqueID,
		reservationID,
	)

	return nil
}

// stationRespondsToCancelReservation implements:
//
//	station {station:string} responds to CancelReservation with status {status:string}
//
// It sends a CALLRESULT with the given status for the pending
// CancelReservation CALL.
func stationRespondsToCancelReservation(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	status := args.String("status")

	uniqueID, err := popCSMSCallID(state, station, actionCancelReservation)
	if err != nil {
		return err
	}

	if err := sendCSMSResponse(ctx, state, station, uniqueID, map[string]any{fieldStatus: status}); err != nil {
		return err
	}

	state.Logf(
		"station %q sent CancelReservation.conf (uniqueID=%s, status=%q)",
		station, uniqueID, status,
	)

	return nil
}
