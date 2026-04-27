package primitive

import (
	"context"
	"fmt"

	"github.com/octane-project/octane/pkg/keywords/api"
	"github.com/octane-project/octane/pkg/keywords/registry"
)

func init() {
	registry.Register(api.Keyword{
		Pattern:     "close station {station:string}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        closeStation,
	})
}

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

	if closeErr := sta.Close(); closeErr != nil {
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
