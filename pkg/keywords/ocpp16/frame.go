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
		return fmt.Errorf(stationNotConnectedFormat, station, err)
	}

	frame := []any{
		float64(wire.MessageTypeCall),
		msgID,
		action,
		payload,
	}

	if err := sv.Send(ctx, frame); err != nil {
		return fmt.Errorf(
			"ocpp16: station %q: send %s: %w",
			station,
			action,
			err,
		)
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
) (map[string]any, error) {
	sv, err := state.Station(station)
	if err != nil {
		return nil, fmt.Errorf(
			stationNotConnectedFormat, station, err,
		)
	}

	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	frame, err := sv.Expect(subCtx)
	if err != nil {
		return nil, fmt.Errorf(
			"ocpp16: station %q: expect response: %w", station, err,
		)
	}

	result, err := wire.ParseResult(frame)
	if err != nil {
		return nil, fmt.Errorf(
			"ocpp16: station %q: parse CALLRESULT: %w", station, err,
		)
	}

	var payload map[string]any
	if err := json.Unmarshal(result.Payload, &payload); err != nil {
		return nil, fmt.Errorf(
			"ocpp16: station %q: unmarshal payload: %w", station, err,
		)
	}

	if payload == nil {
		payload = map[string]any{}
	}

	return payload, nil
}

// expectCSMSCall waits for a CALL frame sent by the CSMS on the station's
// connection, validates that the action matches, and returns the uniqueID
// and decoded payload map. The caller must stash the uniqueID for the
// subsequent response keyword.
func expectCSMSCall(
	ctx context.Context,
	state api.State,
	station, action string,
	timeout time.Duration,
) (string, map[string]any, error) {
	sv, err := state.Station(station)
	if err != nil {
		return emptyUniqueID, nil, fmt.Errorf(
			stationNotConnectedFormat,
			station,
			err,
		)
	}

	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	frame, err := sv.Expect(subCtx)
	if err != nil {
		return emptyUniqueID, nil, fmt.Errorf(
			"ocpp16: station %q: expect %s CALL: %w",
			station,
			action,
			err,
		)
	}

	call, err := wire.ParseCall(frame)
	if err != nil {
		return emptyUniqueID, nil, fmt.Errorf(
			"ocpp16: station %q: parse inbound CALL: %w",
			station,
			err,
		)
	}

	if call.Action != action {
		return emptyUniqueID, nil, fmt.Errorf(
			"ocpp16: station %q: expected %s CALL, got action %q",
			station, action, call.Action,
		)
	}

	var payload map[string]any
	if err := json.Unmarshal(call.Payload, &payload); err != nil {
		return emptyUniqueID, nil, fmt.Errorf(
			"ocpp16: station %q: unmarshal %s payload: %w",
			station,
			action,
			err,
		)
	}

	if payload == nil {
		payload = map[string]any{}
	}

	return call.UniqueID, payload, nil
}

// sendCSMSResponse sends a CALLRESULT back to the CSMS in response to
// a CSMS-initiated CALL with the given uniqueID.
func sendCSMSResponse(
	ctx context.Context,
	state api.State,
	station, uniqueID string,
	payload map[string]any,
) error {
	sv, err := state.Station(station)
	if err != nil {
		return fmt.Errorf(stationNotConnectedFormat, station, err)
	}

	frame := []any{
		float64(wire.MessageTypeResult),
		uniqueID,
		payload,
	}

	if err := sv.Send(ctx, frame); err != nil {
		return fmt.Errorf(
			"ocpp16: station %q: send CALLRESULT: %w",
			station,
			err,
		)
	}

	return nil
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

	payload, ok := val.(map[string]any)
	if !ok {
		return nil, false
	}

	return payload, true
}

// popCSMSCallID retrieves and removes the uniqueID stashed by a
// CSMS-initiated CALL keyword (expectCSMSCall). Returns the uniqueID
// and nil, or an empty string and an error if the key is absent.
func popCSMSCallID(state api.State, station, action string) (string, error) {
	val, ok := state.Pop(csmsCallIDKey(station, action))
	if !ok {
		return emptyUniqueID, fmt.Errorf(
			"ocpp16: station %q: no %s uniqueID stashed; call the CSMS-send keyword first",
			station,
			action,
		)
	}

	uid, ok := val.(string)
	if !ok {
		return emptyUniqueID, fmt.Errorf(
			"ocpp16: station %q: %s uniqueID stash has unexpected type %T",
			station, action, val,
		)
	}

	return uid, nil
}

// popPending retrieves and removes the *pendingInfo stashed by the
// preceding "send" keyword.
func popPending(state api.State) (*pendingInfo, bool) {
	val, ok := state.Pop(pendingKey)
	if !ok {
		return nil, false
	}

	info, ok := val.(*pendingInfo)
	if !ok {
		return nil, false
	}

	return info, true
}
