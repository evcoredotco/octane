package primitive

import (
	"context"
	"fmt"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

func init() {
	registry.Register(api.Keyword{
		Pattern:     "expect any frame on station {station:string} within {timeout:duration}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        expectAnyFrame,
	})

	registry.Register(api.Keyword{
		Pattern: "expect a frame of type {messageType:int} on station" +
			" {station:string} within {timeout:duration}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        expectFrameOfType,
	})
}

// ErrTimeout is returned by the expect keywords when no matching frame
// arrives within the specified timeout. It carries the station handle, the
// configured timeout, and the time at which the deadline was computed (per
// the deterministic clock injected via [api.State.Now]).
//
// Use errors.As to inspect the fields:
//
//	var te *primitive.ErrTimeout
//	if errors.As(err, &te) {
//	    fmt.Println("station:", te.Station)
//	    fmt.Println("timeout:", te.Timeout)
//	}
type ErrTimeout struct {
	// Station is the handle name of the station that produced no frame.
	Station string

	// Timeout is the duration that was configured for the expect step.
	Timeout time.Duration

	// Deadline is the time.Time at which the context deadline was set,
	// derived from state.Now().Add(Timeout). In deterministic-clock mode
	// this reflects the simulated wall-clock value, not real wall-clock time
	// (constitution principle IV).
	Deadline time.Time
}

// Error implements the error interface.
func (e *ErrTimeout) Error() string {
	return fmt.Sprintf(
		"primitive: expect on station %q timed out after %s (deadline: %s)",
		e.Station,
		e.Timeout,
		e.Deadline.Format(time.RFC3339Nano),
	)
}

// expectAnyFrame implements the primitive keyword:
//
//	expect any frame on station {station:string} within {timeout:duration}
//
// It derives a deadline from state.Now().Add(timeout) — never from
// time.Now() — so that deterministic-clock scenarios never advance real wall
// time (constitution principle IV). The keyword blocks until a frame arrives
// or the deadline elapses. On success the received frame is logged; on
// timeout [*ErrTimeout] is returned.
func expectAnyFrame(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.String("station")
	timeout := args.Duration("timeout")

	sta, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf(
			"primitive: expect any frame on station %q: %w",
			handle,
			err,
		)
	}

	deadline := state.Now().Add(timeout)
	dctx, cancel := context.WithDeadline(ctx, deadline)

	defer cancel()

	frame, expectErr := sta.Expect(dctx)
	if expectErr != nil {
		if dctx.Err() != nil {
			return &ErrTimeout{
				Station:  handle,
				Timeout:  timeout,
				Deadline: deadline,
			}
		}

		return fmt.Errorf(
			"primitive: expect any frame on station %q: %w",
			handle,
			expectErr,
		)
	}

	// Stash the received frame under "last_frame:<handle>" so that
	// subsequent steps can inspect the frame content without needing
	// a direct reference to the Station handle (spec 004 AC3).
	state.Stash("last_frame:"+handle, frame)

	state.Logf(
		"station %q: received frame (%d elements)",
		handle,
		len(frame),
	)

	return nil
}

// expectFrameOfType implements the primitive keyword:
//
//	expect a frame of type {messageType:int} on station {station:string}
//	within {timeout:duration}
//
// It repeatedly calls [api.Station.Expect] under a single shared deadline
// (derived from state.Now().Add(timeout)) until a frame whose first element
// equals messageType is received. Frames with a different message-type code
// are silently discarded and the loop continues.
//
// On success the keyword logs the matching frame. On timeout [*ErrTimeout]
// is returned; on any non-timeout station error the error is wrapped and
// returned immediately.
func expectFrameOfType(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	messageType := args.Int("messageType")
	handle := args.String("station")
	timeout := args.Duration("timeout")

	sta, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf(
			"primitive: expect frame of type %d on station %q: %w",
			messageType,
			handle,
			err,
		)
	}

	deadline := state.Now().Add(timeout)
	dctx, cancel := context.WithDeadline(ctx, deadline)

	defer cancel()

	for {
		frame, expectErr := sta.Expect(dctx)
		if expectErr != nil {
			if dctx.Err() != nil {
				return &ErrTimeout{
					Station:  handle,
					Timeout:  timeout,
					Deadline: deadline,
				}
			}

			return fmt.Errorf(
				"primitive: expect frame of type %d on station %q: %w",
				messageType,
				handle,
				expectErr,
			)
		}

		if len(frame) == 0 {
			continue
		}

		msgTypeVal, ok := frame[0].(float64)
		if !ok {
			continue
		}

		if int(msgTypeVal) == messageType {
			state.Logf(
				"station %q: received frame of type %d (%d elements)",
				handle,
				messageType,
				len(frame),
			)

			return nil
		}
	}
}
