package lifecycle

import (
	"context"
	"errors"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// errNotConnected is returned when the station exists but IsOpen() is false.
var errNotConnected = errors.New("station is not in the connected state")

// assertConnectedState implements the keyword:
//
//	station {station:string} is in the connected state
//
// It looks up the named station in the runtime and asserts that its
// WebSocket connection is currently open.
func assertConnectedState(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")

	sta, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf("lifecycle: station %q: %w", handle, err)
	}

	if !sta.IsOpen() {
		return fmt.Errorf("lifecycle: station %q: %w", handle, errNotConnected)
	}

	// IsOpen is synchronous; context is checked for linter compliance only.
	_ = ctx

	return nil
}
