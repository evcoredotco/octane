package ocpp16

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// errBootNotAccepted is returned by stationIsInRegisteredState when the
// BootNotification.conf status was not "Accepted".
var errBootNotAccepted = errors.New(
	"BootNotification was not accepted by the CSMS; station is not in the registered state",
)

// sendBootNotification implements:
//
//	station {station:string} sends BootNotification with reason {reason:string}
//
// It constructs a minimal OCPP 1.6 BootNotification.req payload, sends it
// as a CALL frame, and stashes the pending correlation info for the
// subsequent "the CSMS responds with status" step.
//
// The reason parameter does not appear in OCPP 1.6 BootNotification; it is
// used here as a label to distinguish different boot scenarios in step text
// and logs. The actual payload uses fixed required fields.
func sendBootNotification(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	reason := args.String("reason")

	msgID := nextMsgID(state, station, "BootNotification")

	payload := map[string]any{
		"chargePointVendor": "EVCore",
		"chargePointModel":  "OCTANE",
	}

	if err := sendCall(ctx, state, station, msgID, "BootNotification", payload); err != nil {
		return err
	}

	state.Stash(pendingKey, &pendingInfo{
		station: station,
		msgID:   msgID,
		action:  "BootNotification",
	})

	state.Logf("station %q sent BootNotification (reason=%q, msgID=%s)", station, reason, msgID)

	return nil
}

// csmsRespondsWithStatus implements:
//
//	the CSMS responds with status {status:string} within {timeout:duration}
//
// It waits for a CALLRESULT that matches the pending BootNotification CALL,
// checks the status field, and stashes the full payload for subsequent
// assertion steps (heartbeatInterval, currentTime).
//
// When status is "Accepted" it also records the station as registered.
func csmsRespondsWithStatus(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	expectedStatus := args.String("status")
	timeout := args.Duration("timeout")

	info, ok := popPending(state)
	if !ok {
		return errors.New("ocpp16: no pending BootNotification; call sendBootNotification first")
	}

	_, payload, err := expectResult(ctx, state, info.station, timeout)
	if err != nil {
		return err
	}

	gotStatus, _ := payload["status"].(string)
	if gotStatus != expectedStatus {
		return fmt.Errorf(
			"ocpp16: station %q: BootNotification.conf status: want %q, got %q",
			info.station, expectedStatus, gotStatus,
		)
	}

	state.Stash(lastPayloadKey, payload)

	if info.action == "BootNotification" && gotStatus == "Accepted" {
		state.Stash(registeredKey(info.station), true)
	}

	state.Logf(
		"station %q received BootNotification.conf status=%q",
		info.station, gotStatus,
	)

	return nil
}

// stationIsInRegisteredState implements:
//
//	station {station:string} is in the registered state
//
// It checks that a BootNotification with status "Accepted" was received
// during this scenario execution by looking up the registered stash flag.
func stationIsInRegisteredState(
	_ context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")

	val, ok := state.Pop(registeredKey(handle))
	if !ok {
		return fmt.Errorf("ocpp16: station %q: %w", handle, errBootNotAccepted)
	}

	// Re-stash so subsequent calls in the same scenario still see the flag.
	state.Stash(registeredKey(handle), val)

	return nil
}

// responseIncludesHeartbeatInterval implements:
//
//	the response includes a heartbeatInterval between {min:int} and {max:int}
//
// It inspects the most recently stashed CALLRESULT payload for a
// heartbeatInterval field and validates that it falls within [min, max].
func responseIncludesHeartbeatInterval(
	_ context.Context,
	state api.State,
	args api.Args,
) error {
	min := args.Int("min")
	max := args.Int("max")

	payload, ok := peekPayload(state)
	if !ok {
		return errors.New("ocpp16: no CALLRESULT payload stashed; call a response keyword first")
	}

	rawInterval, exists := payload["heartbeatInterval"]
	if !exists {
		return errors.New("ocpp16: BootNotification.conf payload missing heartbeatInterval field")
	}

	interval, ok := rawInterval.(float64)
	if !ok {
		return fmt.Errorf(
			"ocpp16: heartbeatInterval has unexpected type %T (want number)",
			rawInterval,
		)
	}

	intVal := int(interval)
	if intVal < min || intVal > max {
		return fmt.Errorf(
			"ocpp16: heartbeatInterval %d not in [%d, %d]",
			intVal, min, max,
		)
	}

	return nil
}

// responseIncludesCurrentTime implements:
//
//	the response includes a currentTime in ISO-8601 format
//
// It inspects the most recently stashed CALLRESULT payload for a
// currentTime field and validates that it parses as RFC 3339 / ISO-8601.
func responseIncludesCurrentTime(
	_ context.Context,
	state api.State,
	_ api.Args,
) error {
	payload, ok := peekPayload(state)
	if !ok {
		return errors.New("ocpp16: no CALLRESULT payload stashed; call a response keyword first")
	}

	rawTime, exists := payload["currentTime"]
	if !exists {
		return errors.New("ocpp16: payload missing currentTime field")
	}

	timeStr, ok := rawTime.(string)
	if !ok {
		return fmt.Errorf(
			"ocpp16: currentTime has unexpected type %T (want string)",
			rawTime,
		)
	}

	if _, err := time.Parse(time.RFC3339, timeStr); err != nil {
		return fmt.Errorf(
			"ocpp16: currentTime %q is not valid ISO-8601 / RFC 3339: %w",
			timeStr, err,
		)
	}

	return nil
}
