package ocpp16

import (
	"context"
	"errors"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// errNoCSMSEndpoint is returned when state.CSMSBaseURL() is empty.
var errNoCSMSEndpoint = errors.New(
	"no CSMS endpoint configured; run with --csms-endpoint",
)

// errStationNotRegistered is returned by stationIsRegistered when the
// station connection is not open.
var errStationNotRegistered = errors.New(
	"station connection is not open; run station_boot_accepted first",
)

// csmsIsReachable implements:
//
//	the CSMS is reachable
//
// It validates that a CSMS endpoint has been configured. No actual
// network probe is made — the check is that the URL is non-empty.
func csmsIsReachable(
	_ context.Context,
	state api.State,
	_ api.Args,
) error {
	if state.CSMSBaseURL() == "" {
		return errNoCSMSEndpoint
	}

	return nil
}

// operatorProvisionedIDTag implements:
//
//	the operator has provisioned id token {idTag:string} with status {status:string}
//
// OCTANE cannot provision idTags over the wire (constitution principle XII).
// This keyword is a documentation-only precondition that logs which idTag
// the operator must have configured in the CSMS before the run.
func operatorProvisionedIDTag(
	_ context.Context,
	state api.State,
	args api.Args,
) error {
	idTag := args.String("idTag")
	status := args.String("status")

	state.Logf(
		"precondition: operator must provision idTag %q with status %q in the CSMS",
		idTag, status,
	)

	return nil
}

// stationIsRegistered implements:
//
//	station {station:string} is registered to the CSMS
//
// It checks that the station's WebSocket connection is currently open.
// Registration (a completed BootNotification exchange) is implied by
// the dependency chain — the story that declares this keyword in its
// Background must depend on station_boot_accepted.
func stationIsRegistered(
	_ context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")

	sv, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf("ocpp16: station %q: %w", handle, err)
	}

	if !sv.IsOpen() {
		return fmt.Errorf("ocpp16: station %q: %w", handle, errStationNotRegistered)
	}

	return nil
}
