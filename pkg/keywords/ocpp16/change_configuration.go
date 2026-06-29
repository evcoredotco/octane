package ocpp16

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// csmsEnqueuesChangeConfiguration implements:
//
//	the CSMS sends ChangeConfiguration with key {key:string} and value {value:string} to station {station:string} within {timeout:duration}
//
// It waits for an inbound ChangeConfiguration CALL, validates the key and
// value fields, and stashes the uniqueID for the subsequent response keyword.
func csmsEnqueuesChangeConfiguration(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	key := args.String("key")
	value := args.String("value")
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, payload, err := expectCSMSCall(
		ctx,
		state,
		station,
		actionChangeConfiguration,
		timeout,
	)
	if err != nil {
		return err
	}

	gotKey, err := payloadString(payload, "key", actionChangeConfiguration)
	if err != nil {
		return err
	}

	if gotKey != key {
		return fmt.Errorf(
			"ocpp16: station %q: ChangeConfiguration key: want %q, got %q",
			station, key, gotKey,
		)
	}

	gotValue, err := payloadString(payload, "value", actionChangeConfiguration)
	if err != nil {
		return err
	}

	if gotValue != value {
		return fmt.Errorf(
			"ocpp16: station %q: ChangeConfiguration value: want %q, got %q",
			station, value, gotValue,
		)
	}

	state.Stash(csmsCallIDKey(station, actionChangeConfiguration), uniqueID)

	state.Logf(
		"station %q received ChangeConfiguration CALL (uniqueID=%s, key=%q, value=%q)",
		station,
		uniqueID,
		key,
		value,
	)

	return nil
}

// stationRespondsToChangeConfiguration implements:
//
//	station {station:string} responds to ChangeConfiguration with status {status:string}
//
// It sends a CALLRESULT with the given status for the pending
// ChangeConfiguration CALL.
func stationRespondsToChangeConfiguration(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	status := args.String("status")

	uniqueID, err := popCSMSCallID(state, station, actionChangeConfiguration)
	if err != nil {
		return err
	}

	if err := sendCSMSResponse(ctx, state, station, uniqueID, map[string]any{fieldStatus: status}); err != nil {
		return err
	}

	state.Logf(
		"station %q sent ChangeConfiguration.conf (uniqueID=%s, status=%q)",
		station, uniqueID, status,
	)

	return nil
}
