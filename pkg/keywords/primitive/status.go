package primitive

import (
	"context"
	"errors"
	"fmt"

	"github.com/octane-project/octane/pkg/keywords/api"
	"github.com/octane-project/octane/pkg/keywords/registry"
)

// errConnectionNotOpen is returned when the connection-is-open assertion
// fails: the station exists but is no longer connected.
var errConnectionNotOpen = errors.New("connection is not open")

// errConnectionNotClosed is returned when the connection-is-closed
// assertion fails: the station exists and is still connected.
var errConnectionNotClosed = errors.New("connection is not closed")

func init() {
	registry.Register(api.Keyword{
		Pattern:     "the connection on station {station:string} is open",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        assertConnectionOpen,
	})

	registry.Register(api.Keyword{
		Pattern:     "the connection on station {station:string} is closed",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        assertConnectionClosed,
	})
}

// assertConnectionOpen implements the assertion keyword:
//
//	the connection on station {station:string} is open
//
// It returns an error if the named station is not registered in state
// or if its connection reports as closed.
func assertConnectionOpen(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")

	sta, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf(
			"primitive: assert connection open for %q: %w",
			handle,
			err,
		)
	}

	if !sta.IsOpen() {
		return fmt.Errorf(
			"primitive: station %q: %w",
			handle,
			errConnectionNotOpen,
		)
	}

	_ = ctx

	return nil
}

// assertConnectionClosed implements the assertion keyword:
//
//	the connection on station {station:string} is closed
//
// It returns an error if the named station is not registered in state
// or if its connection reports as still open.
func assertConnectionClosed(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")

	sta, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf(
			"primitive: assert connection closed for %q: %w",
			handle,
			err,
		)
	}

	if sta.IsOpen() {
		return fmt.Errorf(
			"primitive: station %q: %w",
			handle,
			errConnectionNotClosed,
		)
	}

	_ = ctx

	return nil
}
