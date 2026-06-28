package ocpp16

import (
	"context"
	"errors"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// sendAuthorize implements:
//
//	station {station:string} sends Authorize with idTag {idTag:string}
//
// It sends an OCPP 1.6 Authorize.req and stashes the pending correlation
// info for the subsequent response keyword.
func sendAuthorize(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	idTag := args.String("idTag")

	msgID := nextMsgID(state, station, "Authorize")

	payload := map[string]any{
		"idTag": idTag,
	}

	if err := sendCall(ctx, state, station, msgID, "Authorize", payload); err != nil {
		return err
	}

	state.Stash(pendingKey, &pendingInfo{
		station: station,
		msgID:   msgID,
		action:  "Authorize",
	})

	state.Logf("station %q sent Authorize (idTag=%q, msgID=%s)", station, idTag, msgID)

	return nil
}

// csmsRespondsToAuthorize implements:
//
//	the CSMS responds to Authorize with idTagInfo.status {status:string} within {timeout:duration}
//
// It waits for the Authorize.conf CALLRESULT, validates the nested
// idTagInfo.status field, and stashes the payload for subsequent
// assertion steps.
func csmsRespondsToAuthorize(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	expectedStatus := args.String("status")
	timeout := args.Duration("timeout")

	info, ok := popPending(state)
	if !ok {
		return errors.New("ocpp16: no pending Authorize; call sendAuthorize first")
	}

	payload, err := expectResult(ctx, state, info.station, timeout)
	if err != nil {
		return err
	}

	state.Stash(lastPayloadKey, payload)

	rawTagInfo, exists := payload["idTagInfo"]
	if !exists {
		return errors.New("ocpp16: Authorize.conf payload missing idTagInfo field")
	}

	tagInfo, ok := rawTagInfo.(map[string]any)
	if !ok {
		return fmt.Errorf(
			"ocpp16: Authorize.conf idTagInfo has unexpected type %T (want object)",
			rawTagInfo,
		)
	}

	gotStatus, err := payloadString(tagInfo, fieldStatus, "Authorize.conf idTagInfo")
	if err != nil {
		return err
	}

	if gotStatus != expectedStatus {
		return fmt.Errorf(
			"ocpp16: station %q: Authorize.conf idTagInfo.status: want %q, got %q",
			info.station, expectedStatus, gotStatus,
		)
	}

	state.Logf(
		"station %q received Authorize.conf idTagInfo.status=%q",
		info.station, gotStatus,
	)

	return nil
}
