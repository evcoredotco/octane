package ocpp16

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// csmsEnqueuesGetConfiguration implements:
//
//	the CSMS sends GetConfiguration to station {station:string} within {timeout:duration}
//
// It waits for an inbound GetConfiguration CALL and stashes the uniqueID
// for the subsequent response keyword. Per OCPP 1.6 §5.7 the keys list
// in the request is optional; no payload fields are validated.
func csmsEnqueuesGetConfiguration(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	timeout := args.Duration("timeout")

	uniqueID, _, err := expectCSMSCall(ctx, state, station, actionGetConfiguration, timeout)
	if err != nil {
		return err
	}

	state.Stash(csmsCallIDKey(station, actionGetConfiguration), uniqueID)

	state.Logf(
		"station %q received GetConfiguration CALL (uniqueID=%s)",
		station, uniqueID,
	)

	return nil
}

// stationRespondsWithGetConfiguration implements:
//
//	station {station:string} responds to GetConfiguration with {count:int} configuration keys
//
// It sends a CALLRESULT containing count generic configuration key entries
// for the pending GetConfiguration CALL.
func stationRespondsWithGetConfiguration(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	count := args.Int("count")

	uniqueID, err := popCSMSCallID(state, station, actionGetConfiguration)
	if err != nil {
		return err
	}

	keys := make([]any, count)
	for i := configKeyStartIndex; i < count; i++ {
		keys[i] = map[string]any{
			"key":      fmt.Sprintf("ConfigKey%d", i+1),
			"readonly": false,
			"value":    "true",
		}
	}

	resp := map[string]any{
		"configurationKey": keys,
		"unknownKey":       []any{},
	}

	if err := sendCSMSResponse(ctx, state, station, uniqueID, resp); err != nil {
		return err
	}

	state.Logf(
		"station %q sent GetConfiguration.conf (uniqueID=%s, keys=%d)",
		station, uniqueID, count,
	)

	return nil
}
