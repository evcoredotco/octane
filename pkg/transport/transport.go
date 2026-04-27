// Package transport provides the WebSocket client used by OCTANE to simulate
// OCPP charging stations. It wraps nhooyr.io/websocket (ADR 0003) and exposes
// a single [Station] interface that keyword authors (spec 003) call without
// coupling to a specific WebSocket implementation.
//
// TLS verification is on by default. Disabling it via [DialOptions.InsecureSkipVerify]
// emits a banner-level finding in every report (constitution principle X).
//
// All randomness and clocks consumed by callers of this package must be
// injected via the engine primitives in pkg/engine/clock and pkg/engine/rand
// (constitution principle IV). This package does not call time.Now() or
// math/rand directly.
package transport

import (
	"context"
	"crypto/tls"
	"time"
)

// Station is a live WebSocket connection to a simulated OCPP charging station.
// It is the primary interface consumed by keyword authors (spec 003).
//
// Implementations must be safe for concurrent use across goroutines: a writer
// goroutine may call Send while a reader goroutine blocks in Expect.
type Station interface {
	// Send encodes frame as a canonical OCPP-J JSON array and writes it
	// to the WebSocket. Blocks until the frame is on the wire or ctx is
	// cancelled.
	//
	// Returns [ErrStationClosed] if the station has already been closed.
	Send(ctx context.Context, frame []any) error

	// Expect blocks until an inbound OCPP-J frame arrives, ctx is cancelled,
	// or the connection is closed. Frames are delivered in FIFO order.
	//
	// Returns [ErrStationClosed] if the station has already been closed.
	// Returns [ErrFrameTooLarge] if the inbound frame exceeds the configured
	// [DialOptions.MaxFrameBytes] limit.
	Expect(ctx context.Context) ([]any, error)

	// Close gracefully shuts down the WebSocket connection.
	//
	// It is safe to call Close more than once; subsequent calls are no-ops.
	Close() error
}

// DialOptions configures a [Dial] call.
//
// All fields are optional. Zero values select sensible production defaults:
// full TLS verification, a 30-second handshake timeout, and a 1 MiB
// maximum inbound frame size.
type DialOptions struct {
	// Subprotocols lists the OCPP subprotocols to offer in the
	// Sec-WebSocket-Protocol header, in preference order.
	// Typical values: "ocpp1.6", "ocpp2.0.1", "ocpp2.1".
	//
	// If the CSMS selects a protocol not in this list, or omits the
	// Sec-WebSocket-Protocol response header entirely, Dial returns
	// [ErrSubprotocolMismatch].
	Subprotocols []string

	// TLSConfig overrides the default TLS configuration. Nil means use
	// the system default (full chain verification enabled).
	//
	// When [InsecureSkipVerify] is true, the implementor sets
	// InsecureSkipVerify on this config; callers need not set it directly.
	TLSConfig *tls.Config

	// InsecureSkipVerify disables TLS certificate verification when true.
	// Setting this to true emits a banner-level finding in every report.
	// It must never be set without an explicit operator opt-in.
	//
	// When true, the engine logs a WARNING to stderr before dialling.
	InsecureSkipVerify bool

	// MaxFrameBytes is the maximum inbound WebSocket message size in bytes.
	// Zero means use the default (1 MiB).
	//
	// Frames that exceed this limit cause [Expect] to return
	// [ErrFrameTooLarge].
	MaxFrameBytes int64

	// HandshakeTimeout limits the time allowed for the WebSocket upgrade.
	// Zero means use the default (30 seconds).
	HandshakeTimeout time.Duration
}
