package ocpp16

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// csmsEnqueuesUnlockConnector implements:
//
//	the CSMS sends UnlockConnector with connectorId {connectorId:int} to station {station:string} within {timeout:duration}
//
// It waits for an inbound UnlockConnector CALL, validates the connectorId
// field, and stashes the uniqueID for the subsequent response keyword.
func csmsEnqueuesUnlockConnector(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	connectorID := args.Int("connectorId")
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, payload, err := expectCSMSCall(
		ctx,
		state,
		station,
		actionUnlockConnector,
		timeout,
	)
	if err != nil {
		return err
	}

	gotConnector, err := payloadNumber(
		payload,
		fieldConnectorID,
		actionUnlockConnector,
	)
	if err != nil {
		return err
	}

	if int(gotConnector) != connectorID {
		return fmt.Errorf(
			"ocpp16: station %q: UnlockConnector connectorId: want %d, got %d",
			station, connectorID, int(gotConnector),
		)
	}

	state.Stash(csmsCallIDKey(station, actionUnlockConnector), uniqueID)

	state.Logf(
		"station %q received UnlockConnector CALL (uniqueID=%s, connector=%d)",
		station, uniqueID, connectorID,
	)

	return nil
}

// stationRespondsToUnlockConnector implements:
//
//	station {station:string} responds to UnlockConnector with status {status:string}
//
// It sends a CALLRESULT with the given status for the pending
// UnlockConnector CALL.
func stationRespondsToUnlockConnector(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	status := args.String("status")

	uniqueID, err := popCSMSCallID(state, station, actionUnlockConnector)
	if err != nil {
		return err
	}

	if err := sendCSMSResponse(ctx, state, station, uniqueID, map[string]any{fieldStatus: status}); err != nil {
		return err
	}

	state.Logf(
		"station %q sent UnlockConnector.conf (uniqueID=%s, status=%q)",
		station, uniqueID, status,
	)

	return nil
}
