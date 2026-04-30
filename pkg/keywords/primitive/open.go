package primitive

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/transport"
)

// openWebSocket implements the primitive keyword:
//
//	open a WebSocket to {url:string} as station {station:string}
//
// It dials the given WebSocket URL with no subprotocol preference and
// registers the resulting [transport.Station] in the runtime state under
// the given station handle name.
func openWebSocket(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	rawURL := args.String("url")
	handle := args.String("station")

	//nolint:exhaustruct // zero values are correct defaults
	return dial(ctx, state, rawURL, handle, transport.DialOptions{})
}

// openWebSocketWithSubprotocol implements the primitive keyword:
//
//	open a WebSocket to {url:string} as station {station:string}
//	with subprotocol {subprotocol:string}
//
// It dials the given WebSocket URL offering a single subprotocol and
// registers the resulting [transport.Station] in the runtime state under
// the given station handle name.
func openWebSocketWithSubprotocol(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	rawURL := args.String("url")
	handle := args.String("station")
	subprotocol := args.String("subprotocol")

	//nolint:exhaustruct // only subprotocol is non-default
	opts := transport.DialOptions{
		Subprotocols: []string{subprotocol},
	}

	return dial(ctx, state, rawURL, handle, opts)
}

// dial performs the shared WebSocket dial logic for both open keywords.
// On success it wraps the [transport.Station] in a [stationAdapter] so
// that the [api.Station] interface is satisfied, then registers it in
// state under handle.
func dial(
	ctx context.Context,
	state api.State,
	rawURL string,
	handle string,
	opts transport.DialOptions,
) error {
	sta, err := transport.Dial(ctx, rawURL, opts)
	if err != nil {
		return fmt.Errorf(
			"primitive: open WebSocket %q as %q: %w",
			rawURL,
			handle,
			err,
		)
	}

	state.RegisterStation(handle, &stationAdapter{inner: sta})
	state.Logf("station %q connected to %s", handle, rawURL)

	return nil
}
