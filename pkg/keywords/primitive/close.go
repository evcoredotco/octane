package primitive

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// closeStation implements the primitive keyword:
//
//	close station {station:string}
//
// It looks up the named station in the runtime state and closes its
// WebSocket connection. An error is returned if the handle is not
// registered or if the underlying close fails.
func closeStation(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")

	sta, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf(
			"primitive: close station %q: %w",
			handle,
			err,
		)
	}

	closeErr := sta.Close()
	if closeErr != nil {
		return fmt.Errorf(
			"primitive: close station %q: %w",
			handle,
			closeErr,
		)
	}

	state.Logf("station %q closed", handle)

	// ctx is intentionally unused: Close() is synchronous and never
	// blocks, so no cancellation check is needed here.
	_ = ctx

	return nil
}
