package ocpp16

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/wire"
)

// sendCall constructs an OCPP-J CALL frame from the given unique ID,
// action name, and payload map, and sends it on the station's connection.
// It returns an error if the station is not registered or if the send fails.
func sendCall(
	ctx context.Context,
	state api.State,
	station, msgID, action string,
	payload map[string]any,
) error {
	sv, err := state.Station(station)
	if err != nil {
		return fmt.Errorf("ocpp16: station %q: not connected: %w", station, err)
	}

	frame := []any{
		float64(wire.MessageTypeCall),
		msgID,
		action,
		payload,
	}

	if err := sv.Send(ctx, frame); err != nil {
		return fmt.Errorf("ocpp16: station %q: send %s: %w", station, action, err)
	}

	return nil
}

// expectResult waits for the next inbound frame from station and parses it
// as a CALLRESULT. It returns the decoded Result and the payload unmarshalled
// into a map[string]any.
//
// The sub-context created from ctx is bounded by timeout. If the context
// expires or the frame is malformed, a descriptive error is returned.
func expectResult(
	ctx context.Context,
	state api.State,
	station string,
	timeout time.Duration,
) (wire.Result, map[string]any, error) {
	sv, err := state.Station(station)
	if err != nil {
		return wire.Result{}, nil, fmt.Errorf(
			"ocpp16: station %q: not connected: %w", station, err,
		)
	}

	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	frame, err := sv.Expect(subCtx)
	if err != nil {
		return wire.Result{}, nil, fmt.Errorf(
			"ocpp16: station %q: expect response: %w", station, err,
		)
	}

	result, err := wire.ParseResult(frame)
	if err != nil {
		return wire.Result{}, nil, fmt.Errorf(
			"ocpp16: station %q: parse CALLRESULT: %w", station, err,
		)
	}

	var payload map[string]any
	if err := json.Unmarshal(result.Payload, &payload); err != nil {
		return wire.Result{}, nil, fmt.Errorf(
			"ocpp16: station %q: unmarshal payload: %w", station, err,
		)
	}

	if payload == nil {
		payload = map[string]any{}
	}

	return result, payload, nil
}

// peekPayload retrieves the most recently stashed CALLRESULT payload
// without consuming it, so that multiple assertion keywords can inspect
// the same response. It re-stashes the value after popping it.
func peekPayload(state api.State) (map[string]any, bool) {
	val, ok := state.Pop(lastPayloadKey)
	if !ok {
		return nil, false
	}

	state.Stash(lastPayloadKey, val)

	payload, _ := val.(map[string]any)

	return payload, payload != nil
}

// popPending retrieves and removes the *pendingInfo stashed by the
// preceding "send" keyword.
func popPending(state api.State) (*pendingInfo, bool) {
	val, ok := state.Pop(pendingKey)
	if !ok {
		return nil, false
	}

	info, _ := val.(*pendingInfo)

	return info, info != nil
}
