package primitive

import (
	"context"
	"fmt"

	"github.com/evcoreco/octane/pkg/transport"
)

// stationAdapter wraps a [transport.Station] so that it satisfies
// [api.Station]. The adapter is a thin pass-through; it adds no
// buffering or logic of its own.
//
// The adapter is not exported. Callers interact only with the
// [api.Station] interface returned by [api.State.Station].
type stationAdapter struct {
	inner transport.Station
}

// Send delegates to the inner [transport.Station].
func (a *stationAdapter) Send(ctx context.Context, frame []any) error {
	err := a.inner.Send(ctx, frame)
	if err != nil {
		return fmt.Errorf("adapter: send: %w", err)
	}

	return nil
}

// Expect delegates to the inner [transport.Station].
func (a *stationAdapter) Expect(ctx context.Context) ([]any, error) {
	frame, err := a.inner.Expect(ctx)
	if err != nil {
		return nil, fmt.Errorf("adapter: expect: %w", err)
	}

	return frame, nil
}

// Close delegates to the inner [transport.Station].
func (a *stationAdapter) Close() error {
	err := a.inner.Close()
	if err != nil {
		return fmt.Errorf("adapter: close: %w", err)
	}

	return nil
}

// IsOpen delegates to the inner [transport.Station].
func (a *stationAdapter) IsOpen() bool {
	return a.inner.IsOpen()
}
