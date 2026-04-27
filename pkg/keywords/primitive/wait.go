package primitive

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

func init() {
	registry.Register(api.Keyword{
		Pattern:     "wait {duration:duration}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        waitDuration,
	})
}

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

	if err := state.Sleep(ctx, dur); err != nil {
		return fmt.Errorf("primitive: wait %s: %w", dur, err)
	}

	state.Logf("waited %s", dur)

	return nil
}
