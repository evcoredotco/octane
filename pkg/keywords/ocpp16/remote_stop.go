package ocpp16

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// csmsEnqueuesRemoteStop implements:
//
//	the CSMS sends RemoteStopTransaction with transactionId {transactionId:int} to station {station:string} within {timeout:duration}
//
// It waits for an inbound RemoteStopTransaction CALL, validates the
// transactionId field, and stashes the uniqueID for the subsequent
// response keyword.
func csmsEnqueuesRemoteStop(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	transactionID := args.Int("transactionId")
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, payload, err := expectCSMSCall(ctx, state, station, "RemoteStopTransaction", timeout)
	if err != nil {
		return err
	}

	gotTxID, _ := payload["transactionId"].(float64)
	if int(gotTxID) != transactionID {
		return fmt.Errorf(
			"ocpp16: station %q: RemoteStopTransaction transactionId: want %d, got %d",
			station, transactionID, int(gotTxID),
		)
	}

	state.Stash(csmsCallIDKey(station, "RemoteStopTransaction"), uniqueID)

	state.Logf(
		"station %q received RemoteStopTransaction CALL (uniqueID=%s, transactionId=%d)",
		station, uniqueID, transactionID,
	)

	return nil
}

// stationRespondsToRemoteStop implements:
//
//	station {station:string} responds to RemoteStopTransaction with status {status:string}
//
// It sends a CALLRESULT with the given status for the pending
// RemoteStopTransaction CALL.
func stationRespondsToRemoteStop(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	status := args.String("status")

	uniqueID, err := popCSMSCallID(state, station, "RemoteStopTransaction")
	if err != nil {
		return err
	}

	if err := sendCSMSResponse(ctx, state, station, uniqueID, map[string]any{"status": status}); err != nil {
		return err
	}

	state.Logf(
		"station %q sent RemoteStopTransaction.conf (uniqueID=%s, status=%q)",
		station, uniqueID, status,
	)

	return nil
}
