package ocpp16

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// csmsEnqueuesReset implements:
//
//	the CSMS sends Reset with type {resetType:string} to station {station:string} within {timeout:duration}
//
// It waits for an inbound Reset CALL, validates the type field, and
// stashes the uniqueID for the subsequent response keyword.
func csmsEnqueuesReset(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	resetType := args.String("resetType")
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, payload, err := expectCSMSCall(ctx, state, station, actionReset, timeout)
	if err != nil {
		return err
	}

	gotType, err := payloadString(payload, "type", actionReset)
	if err != nil {
		return err
	}

	if gotType != resetType {
		return fmt.Errorf(
			"ocpp16: station %q: Reset type: want %q, got %q",
			station, resetType, gotType,
		)
	}

	state.Stash(csmsCallIDKey(station, actionReset), uniqueID)

	state.Logf(
		"station %q received Reset CALL (uniqueID=%s, type=%q)",
		station, uniqueID, resetType,
	)

	return nil
}

// stationRespondsToReset implements:
//
//	station {station:string} responds to Reset with status {status:string}
//
// It sends a CALLRESULT with the given status for the pending Reset CALL.
func stationRespondsToReset(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	status := args.String("status")

	uniqueID, err := popCSMSCallID(state, station, actionReset)
	if err != nil {
		return err
	}

	if err := sendCSMSResponse(ctx, state, station, uniqueID, map[string]any{fieldStatus: status}); err != nil {
		return err
	}

	state.Logf(
		"station %q sent Reset.conf (uniqueID=%s, status=%q)",
		station, uniqueID, status,
	)

	return nil
}
