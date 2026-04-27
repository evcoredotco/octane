package primitive

import (
	"context"

	"github.com/octane-project/octane/pkg/transport"
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
	return a.inner.Send(ctx, frame)
}

// Expect delegates to the inner [transport.Station].
func (a *stationAdapter) Expect(ctx context.Context) ([]any, error) {
	return a.inner.Expect(ctx)
}

// Close delegates to the inner [transport.Station].
func (a *stationAdapter) Close() error {
	return a.inner.Close()
}

// IsOpen delegates to the inner [transport.Station].
func (a *stationAdapter) IsOpen() bool {
	return a.inner.IsOpen()
}
