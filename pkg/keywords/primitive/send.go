package primitive

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

// errFrameNotSlice is returned when the frame argument is not a JSON array.
var errFrameNotSlice = errors.New("frame must be []any")

func init() {
	registry.Register(api.Keyword{
		Pattern:     "send raw frame {frame:any} on station {station:string}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        sendRawFrame,
	})

	registry.Register(api.Keyword{
		Pattern: "send raw bytes {bytes:string}" +
			" on station {station:string}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        sendRawBytes,
	})
}

// sendRawFrame implements the primitive keyword:
//
//	send raw frame {frame:any} on station {station:string}
//
// The frame argument must be a []any — the decoded Go representation of an
// OCPP-J JSON array (per ADR 0006). Any other type returns [FrameShapeError].
// The frame is encoded by the transport layer and emitted on the station's
// WebSocket connection.
func sendRawFrame(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")
	raw := args.Any("frame")

	frame, ok := raw.([]any)
	if !ok {
		return fmt.Errorf(
			"primitive: send raw frame on station %q: %w, got %T",
			handle,
			errFrameNotSlice,
			raw,
		)
	}

	sta, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf(
			"primitive: send raw frame on station %q: %w",
			handle,
			err,
		)
	}

	sendErr := sta.Send(ctx, frame)
	if sendErr != nil {
		return fmt.Errorf(
			"primitive: send raw frame on station %q: %w",
			handle,
			sendErr,
		)
	}

	state.Logf(
		"station %q: sent raw frame (%d elements)",
		handle,
		len(frame),
	)

	return nil
}

// sendRawBytes implements the primitive keyword:
//
//	send raw bytes {bytes:string} on station {station:string}
//
// The bytes argument is a hex-encoded string (e.g. "deadbeef"). It is
// decoded from hex, then parsed as a JSON array into a []any value, and
// sent via [api.Station.Send]. This allows story authors to construct
// deliberately malformed or extension OCPP-J frames for negative-path
// conformance testing (spec 004 OQ1).
//
// An error is returned if:
//   - the hex string is invalid,
//   - the decoded bytes do not parse as a JSON array, or
//   - the underlying send fails.
func sendRawBytes(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")
	hexStr := args.String("bytes")

	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return fmt.Errorf(
			"primitive: send raw bytes on station %q: "+
				"invalid hex string: %w",
			handle,
			err,
		)
	}

	var frame []any

	jsonErr := json.Unmarshal(decoded, &frame)
	if jsonErr != nil {
		return fmt.Errorf(
			"primitive: send raw bytes on station %q: "+
				"decoded bytes are not a JSON array: %w",
			handle,
			jsonErr,
		)
	}

	sta, stErr := state.Station(handle)
	if stErr != nil {
		return fmt.Errorf(
			"primitive: send raw bytes on station %q: %w",
			handle,
			stErr,
		)
	}

	sendErr := sta.Send(ctx, frame)
	if sendErr != nil {
		return fmt.Errorf(
			"primitive: send raw bytes on station %q: %w",
			handle,
			sendErr,
		)
	}

	state.Logf(
		"station %q: sent raw bytes (%d bytes, %d frame elements)",
		handle,
		len(decoded),
		len(frame),
	)

	return nil
}
