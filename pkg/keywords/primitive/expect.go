package primitive

import (
	"context"
	"fmt"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// TimeoutError is returned by the expect keywords when no matching frame
// arrives within the specified timeout. It carries the station handle, the
// configured timeout, and the time at which the deadline was computed (per
// the deterministic clock injected via [api.State.Now]).
//
// Use errors.As to inspect the fields:
//
//	var te *primitive.TimeoutError
//	if errors.As(err, &te) {
//	    fmt.Println("station:", te.Station)
//	    fmt.Println("timeout:", te.Timeout)
//	}
type TimeoutError struct {
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
func (e *TimeoutError) Error() string {
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
// timeout [*TimeoutError] is returned.
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
			return &TimeoutError{
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

// frameMessageType returns the OCPP-J message-type code from the first
// element of frame, and reports whether the extraction succeeded. A
// frame with zero elements or a non-float64 first element yields ok=false.
func frameMessageType(frame []any) (int, bool) {
	if len(frame) == 0 {
		return 0, false
	}

	msgTypeVal, ok := frame[0].(float64)
	if !ok {
		return 0, false
	}

	return int(msgTypeVal), true
}

// receiveTypedFrame reads from sta under dctx until a frame whose
// message-type code equals wantType is found, then returns it.
// It returns an error wrapping dctx timeout or a station error.
func receiveTypedFrame(
	dctx context.Context,
	sta api.Station,
	wantType int,
	handle string,
	deadline time.Time,
	timeout time.Duration,
) ([]any, error) {
	for {
		frame, err := sta.Expect(dctx)
		if err != nil {
			return nil, wrapExpectError(
				dctx, err, wantType, handle, deadline, timeout,
			)
		}

		msgType, ok := frameMessageType(frame)
		if ok && msgType == wantType {
			return frame, nil
		}
	}
}

// wrapExpectError maps a station Expect error to either a TimeoutError
// (when the context deadline was reached) or a wrapped fmt error.
func wrapExpectError(
	dctx context.Context,
	err error,
	messageType int,
	handle string,
	deadline time.Time,
	timeout time.Duration,
) error {
	if dctx.Err() != nil {
		return &TimeoutError{
			Station:  handle,
			Timeout:  timeout,
			Deadline: deadline,
		}
	}

	return fmt.Errorf(
		"primitive: expect frame of type %d on station %q: %w",
		messageType,
		handle,
		err,
	)
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
// On success the keyword logs the matching frame. On timeout [*TimeoutError]
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

	frame, err := receiveTypedFrame(
		dctx, sta, messageType, handle, deadline, timeout,
	)
	if err != nil {
		return err
	}

	state.Logf(
		"station %q: received frame of type %d (%d elements)",
		handle,
		messageType,
		len(frame),
	)

	return nil
}
