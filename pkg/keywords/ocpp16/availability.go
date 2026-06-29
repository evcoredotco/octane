package ocpp16

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// csmsEnqueuesChangeAvailability implements:
//
//	the CSMS sends ChangeAvailability with connectorId {connectorId:int} and type {availType:string} to station {station:string} within {timeout:duration}
//
// It waits for an inbound ChangeAvailability CALL, validates the
// connectorId and type fields, and stashes the uniqueID for the
// subsequent response keyword.
func csmsEnqueuesChangeAvailability(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	connectorID := args.Int("connectorId")
	availType := args.String("availType")
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, payload, err := expectCSMSCall(
		ctx,
		state,
		station,
		actionChangeAvailability,
		timeout,
	)
	if err != nil {
		return err
	}

	gotConnector, err := payloadNumber(
		payload,
		fieldConnectorID,
		actionChangeAvailability,
	)
	if err != nil {
		return err
	}

	if int(gotConnector) != connectorID {
		return fmt.Errorf(
			"ocpp16: station %q: ChangeAvailability connectorId: want %d, got %d",
			station,
			connectorID,
			int(gotConnector),
		)
	}

	gotType, err := payloadString(payload, "type", actionChangeAvailability)
	if err != nil {
		return err
	}

	if gotType != availType {
		return fmt.Errorf(
			"ocpp16: station %q: ChangeAvailability type: want %q, got %q",
			station, availType, gotType,
		)
	}

	state.Stash(csmsCallIDKey(station, actionChangeAvailability), uniqueID)

	state.Logf(
		"station %q received ChangeAvailability CALL (uniqueID=%s, connector=%d, type=%q)",
		station,
		uniqueID,
		connectorID,
		availType,
	)

	return nil
}

// stationRespondsToChangeAvailability implements:
//
//	station {station:string} responds to ChangeAvailability with status {status:string}
//
// It sends a CALLRESULT with the given status for the pending
// ChangeAvailability CALL.
func stationRespondsToChangeAvailability(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	status := args.String("status")

	uniqueID, err := popCSMSCallID(state, station, actionChangeAvailability)
	if err != nil {
		return err
	}

	if err := sendCSMSResponse(ctx, state, station, uniqueID, map[string]any{fieldStatus: status}); err != nil {
		return err
	}

	state.Logf(
		"station %q sent ChangeAvailability.conf (uniqueID=%s, status=%q)",
		station, uniqueID, status,
	)

	return nil
}
