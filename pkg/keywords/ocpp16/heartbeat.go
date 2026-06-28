package ocpp16

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// sendHeartbeat implements:
//
//	station {station:string} sends Heartbeat
//
// It sends an OCPP 1.6 Heartbeat.req (empty payload) and stashes the
// pending correlation info for the subsequent response keyword.
func sendHeartbeat(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")

	msgID := nextMsgID(state, station, "Heartbeat")

	if err := sendCall(ctx, state, station, msgID, "Heartbeat", map[string]any{}); err != nil {
		return err
	}

	state.Stash(pendingKey, &pendingInfo{
		station: station,
		msgID:   msgID,
		action:  "Heartbeat",
	})

	state.Logf("station %q sent Heartbeat (msgID=%s)", station, msgID)

	return nil
}

// csmsRespondsToHeartbeat implements:
//
//	the CSMS responds to the Heartbeat within {timeout:duration}
//
// It waits for the Heartbeat.conf CALLRESULT, stashes the payload for
// the subsequent currentTime assertion step.
func csmsRespondsToHeartbeat(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	timeout := args.Duration("timeout")

	info, ok := popPending(state)
	if !ok {
		return errors.New("ocpp16: no pending Heartbeat; call sendHeartbeat first")
	}

	payload, err := expectResult(ctx, state, info.station, timeout)
	if err != nil {
		return err
	}

	state.Stash(lastPayloadKey, payload)

	state.Logf("station %q received Heartbeat.conf", info.station)

	return nil
}

// heartbeatResponseIncludesCurrentTime implements:
//
//	the Heartbeat response includes a currentTime in ISO-8601 format
//
// It inspects the stashed Heartbeat.conf payload and validates the
// currentTime field parses as RFC 3339 / ISO-8601.
func heartbeatResponseIncludesCurrentTime(
	_ context.Context,
	state api.State,
	_ api.Args,
) error {
	payload, ok := peekPayload(state)
	if !ok {
		return errors.New("ocpp16: no Heartbeat.conf payload stashed; call csmsRespondsToHeartbeat first")
	}

	rawTime, exists := payload["currentTime"]
	if !exists {
		return errors.New("ocpp16: Heartbeat.conf payload missing currentTime field")
	}

	timeStr, ok := rawTime.(string)
	if !ok {
		return fmt.Errorf(
			"ocpp16: Heartbeat.conf currentTime has unexpected type %T (want string)",
			rawTime,
		)
	}

	if _, err := time.Parse(time.RFC3339, timeStr); err != nil {
		return fmt.Errorf(
			"ocpp16: Heartbeat.conf currentTime %q is not valid ISO-8601 / RFC 3339: %w",
			timeStr, err,
		)
	}

	return nil
}
