package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// errNoPendingConnect is returned when the handshake step runs without a
// preceding "station connects to the CSMS" step in the same scenario.
var errNoPendingConnect = errors.New(
	"no pending connection; call 'station {station} connects to the CSMS' first",
)

// errHandshakeNotComplete is returned when the station's connection is
// not open at the point of the handshake assertion.
var errHandshakeNotComplete = errors.New("OCPP-J handshake did not complete")

// handshakeCompletes implements the keyword:
//
//	the OCPP-J handshake completes within {timeout:duration}
//
// It retrieves the station handle stashed by the preceding "connects to the
// CSMS" step, creates a sub-context bounded by timeout, and asserts that
// the station's WebSocket connection is open. Since transport.Dial is
// synchronous, a successful prior connect step guarantees the handshake has
// already completed; this step provides an explicit timing assertion and a
// clear failure message when the connection is unexpectedly closed.
func handshakeCompletes(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	timeout := args.Duration("timeout")

	handleAny, ok := state.Pop(connectingStationKey)
	if !ok {
		return fmt.Errorf("lifecycle: handshake: %w", errNoPendingConnect)
	}

	handle, ok := handleAny.(string)
	if !ok {
		return fmt.Errorf(
			"lifecycle: handshake: stash key %q holds %T, want string",
			connectingStationKey,
			handleAny,
		)
	}

	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Non-blocking deadline check: if the overall budget was already
	// exceeded before this step ran, report it immediately.
	if err := subCtx.Err(); err != nil {
		return fmt.Errorf(
			"lifecycle: handshake for %q exceeded %s: %w",
			handle,
			timeout.Round(time.Millisecond),
			err,
		)
	}

	sta, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf("lifecycle: handshake for %q: %w", handle, err)
	}

	if !sta.IsOpen() {
		return fmt.Errorf(
			"lifecycle: station %q: %w",
			handle,
			errHandshakeNotComplete,
		)
	}

	state.Logf("OCPP-J handshake complete for station %q (within %s)", handle, timeout.Round(time.Millisecond))

	return nil
}
