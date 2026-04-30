package primitive

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// waitDuration implements the primitive keyword:
//
//	wait {duration:duration}
//
// It sleeps the runtime's clock by exactly the requested duration via
// [api.State.Sleep]. In deterministic-clock mode the injected clock
// advances only when the test calls Advance, so no real wall-clock time
// elapses (constitution principle IV, spec 004 AC5).
//
// If the context is cancelled before the duration elapses, the error
// returned by [api.State.Sleep] is wrapped and returned to the caller.
func waitDuration(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	dur := args.Duration("duration")

	err := state.Sleep(ctx, dur)
	if err != nil {
		return fmt.Errorf("primitive: wait %s: %w", dur, err)
	}

	state.Logf("waited %s", dur)

	return nil
}
