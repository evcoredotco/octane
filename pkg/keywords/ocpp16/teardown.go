package ocpp16

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// disconnectStation implements:
//
//	Disconnect station {station:string}
//
// It closes the station's WebSocket connection. Used in Teardown sections
// to release the connection after each scenario completes.
func disconnectStation(
	_ context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")

	sv, err := state.Station(station)
	if err != nil {
		return fmt.Errorf("ocpp16: station %q: not connected: %w", station, err)
	}

	if err := sv.Close(); err != nil {
		return fmt.Errorf("ocpp16: station %q: close: %w", station, err)
	}

	state.Logf("station %q disconnected", station)

	return nil
}
