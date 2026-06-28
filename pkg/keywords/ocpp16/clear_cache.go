package ocpp16

import (
	"context"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// csmsEnqueuesClearCache implements:
//
//	the CSMS sends ClearCache to station {station:string} within {timeout:duration}
//
// It waits for an inbound ClearCache CALL and stashes the uniqueID for
// the subsequent response keyword. Per OCPP 1.6 §5.4 the request payload
// is always empty; no fields are validated.
func csmsEnqueuesClearCache(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, _, err := expectCSMSCall(ctx, state, station, "ClearCache", timeout)
	if err != nil {
		return err
	}

	state.Stash(csmsCallIDKey(station, "ClearCache"), uniqueID)

	state.Logf(
		"station %q received ClearCache CALL (uniqueID=%s)",
		station, uniqueID,
	)

	return nil
}

// stationRespondsToClearCache implements:
//
//	station {station:string} responds to ClearCache with status {status:string}
//
// It sends a CALLRESULT with the given status for the pending ClearCache CALL.
func stationRespondsToClearCache(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	status := args.String("status")

	uniqueID, err := popCSMSCallID(state, station, "ClearCache")
	if err != nil {
		return err
	}

	if err := sendCSMSResponse(ctx, state, station, uniqueID, map[string]any{"status": status}); err != nil {
		return err
	}

	state.Logf(
		"station %q sent ClearCache.conf (uniqueID=%s, status=%q)",
		station, uniqueID, status,
	)

	return nil
}
