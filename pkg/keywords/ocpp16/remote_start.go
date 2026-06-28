package ocpp16

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// csmsEnqueuesRemoteStart implements:
//
//	the CSMS sends RemoteStartTransaction with connectorId {connectorId:int} and idTag {idTag:string} to station {station:string} within {timeout:duration}
//
// It waits for an inbound RemoteStartTransaction CALL, validates the
// connectorId and idTag fields, and stashes the uniqueID for the
// subsequent response keyword.
func csmsEnqueuesRemoteStart(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	connectorID := args.Int("connectorId")
	idTag := args.String("idTag")
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, payload, err := expectCSMSCall(ctx, state, station, actionRemoteStartTransaction, timeout)
	if err != nil {
		return err
	}

	gotConnector, err := payloadNumber(payload, fieldConnectorID, actionRemoteStartTransaction)
	if err != nil {
		return err
	}

	if int(gotConnector) != connectorID {
		return fmt.Errorf(
			"ocpp16: station %q: RemoteStartTransaction connectorId: want %d, got %d",
			station, connectorID, int(gotConnector),
		)
	}

	gotIDTag, err := payloadString(payload, fieldIDTag, actionRemoteStartTransaction)
	if err != nil {
		return err
	}

	if gotIDTag != idTag {
		return fmt.Errorf(
			"ocpp16: station %q: RemoteStartTransaction idTag: want %q, got %q",
			station, idTag, gotIDTag,
		)
	}

	state.Stash(csmsCallIDKey(station, actionRemoteStartTransaction), uniqueID)

	state.Logf(
		"station %q received RemoteStartTransaction CALL (uniqueID=%s, connector=%d, idTag=%q)",
		station, uniqueID, connectorID, idTag,
	)

	return nil
}

// stationRespondsToRemoteStart implements:
//
//	station {station:string} responds to RemoteStartTransaction with status {status:string}
//
// It sends a CALLRESULT with the given status for the pending
// RemoteStartTransaction CALL.
func stationRespondsToRemoteStart(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	status := args.String("status")

	uniqueID, err := popCSMSCallID(state, station, actionRemoteStartTransaction)
	if err != nil {
		return err
	}

	if err := sendCSMSResponse(ctx, state, station, uniqueID, map[string]any{fieldStatus: status}); err != nil {
		return err
	}

	state.Logf(
		"station %q sent RemoteStartTransaction.conf (uniqueID=%s, status=%q)",
		station, uniqueID, status,
	)

	return nil
}
