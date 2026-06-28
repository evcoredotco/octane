package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/transport"
)

// errNoCSMSEndpoint is returned when state.CSMSBaseURL() is empty.
var errNoCSMSEndpoint = errors.New(
	"no CSMS endpoint configured; run with --csms-endpoint",
)

// connectingStationKey is the stash key used to pass the station handle
// from the connect step to the subsequent handshake assertion step.
const connectingStationKey = "lifecycle:connecting_station"

// connectToCSMS implements the keyword:
//
//	station {station:string} connects to the CSMS
//
// It constructs the per-station WebSocket URL from state.CSMSBaseURL() and
// the station handle (e.g. "ws://localhost:9210/CP01"), dials with the
// "ocpp1.6" subprotocol, registers the resulting station in the runtime
// under the given handle, and stashes the handle for the subsequent
// "the OCPP-J handshake completes" assertion step.
func connectToCSMS(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")

	baseURL := state.CSMSBaseURL()
	if baseURL == "" {
		return fmt.Errorf("lifecycle: station %q: %w", handle, errNoCSMSEndpoint)
	}

	stationURL := strings.TrimRight(baseURL, "/") + "/" + handle

	//nolint:exhaustruct // only Subprotocols is non-default
	sta, err := transport.Dial(ctx, stationURL, transport.DialOptions{
		Subprotocols: []string{"ocpp1.6"},
	})
	if err != nil {
		return fmt.Errorf(
			"lifecycle: station %q: connect to CSMS at %s: %w",
			handle,
			stationURL,
			err,
		)
	}

	state.RegisterStation(handle, sta)
	state.Stash(connectingStationKey, handle)
	state.Logf("station %q connected via OCPP-J to %s", handle, stationURL)

	return nil
}
