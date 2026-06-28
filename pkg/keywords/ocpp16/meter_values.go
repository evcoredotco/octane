package ocpp16

import (
	"context"
	"errors"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// sendMeterValues implements:
//
//	station {station:string} sends MeterValues for connector {connectorId:int} with sampled value {value:string}
//
// It sends an OCPP 1.6 MeterValues.req with a single sampled value entry
// for the given connector, and stashes the pending correlation info.
func sendMeterValues(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	station := args.String("station")
	connectorID := args.Int("connectorId")
	value := args.String("value")

	msgID := nextMsgID(state, station, "MeterValues")

	payload := map[string]any{
		fieldConnectorID: connectorID,
		"meterValue": []any{
			map[string]any{
				fieldTimestamp: state.Now().Format(iso8601SecondFormat),
				"sampledValue": []any{
					map[string]any{"value": value, "unit": "Wh"},
				},
			},
		},
	}

	if err := sendCall(ctx, state, station, msgID, "MeterValues", payload); err != nil {
		return err
	}

	state.Stash(pendingKey, &pendingInfo{
		station: station,
		msgID:   msgID,
		action:  "MeterValues",
	})

	state.Logf(
		"station %q sent MeterValues (connector=%d, value=%q, msgID=%s)",
		station, connectorID, value, msgID,
	)

	return nil
}

// csmsAcknowledgesMeterValues implements:
//
//	the CSMS acknowledges MeterValues within {timeout:duration}
//
// It waits for the MeterValues.conf CALLRESULT. Per OCPP 1.6 §4.7 the
// confirmation payload is empty; success means the frame arrived without error.
func csmsAcknowledgesMeterValues(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	timeout := args.Duration("timeout")

	info, ok := popPending(state)
	if !ok {
		return errors.New("ocpp16: no pending MeterValues; call sendMeterValues first")
	}

	payload, err := expectResult(ctx, state, info.station, timeout)
	if err != nil {
		return err
	}

	state.Stash(lastPayloadKey, payload)

	state.Logf(
		"station %q received MeterValues.conf (acknowledged)",
		info.station,
	)

	return nil
}
