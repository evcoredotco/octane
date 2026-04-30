package api

import "context"

// Station is the wire-I/O surface for a single charging station
// connection. Keywords use it to send OCPP-J frames to a CSMS,
// receive frames from the wire, close the connection, and check
// whether the connection is still open.
//
// Station is an interface so that keyword unit tests can supply
// a mock without importing pkg/transport/ or any network library
// (spec 003 AC8). The runtime's concrete implementation wraps
// the WebSocket transport and is obtained via [State.Station].
//
// Frames are represented as []any — the decoded Go form of an
// OCPP-J JSON array (per ADR 0006: arrays decode to []any,
// numbers to float64). For example, a CALL frame is:
//
//	[]any{2, "messageId", "BootNotification", map[string]any{...}}
type Station interface {
	// Send transmits an OCPP-J frame to the CSMS over the
	// station's WebSocket connection. The frame must be a valid
	// OCPP-J JSON array in its decoded Go form ([]any).
	//
	// The context carries the per-step timeout. Send returns an
	// error if the write fails or the context expires.
	Send(ctx context.Context, frame []any) error

	// Expect blocks until an OCPP-J frame arrives on the
	// station's WebSocket connection or the context expires.
	//
	// The returned frame is the decoded Go form of the JSON
	// array received from the wire. An error is returned if the
	// read fails, the connection is closed, or the context
	// expires before a frame arrives.
	Expect(ctx context.Context) ([]any, error)

	// Close gracefully shuts down the WebSocket connection.
	// Subsequent calls are no-ops and return nil.
	Close() error

	// IsOpen reports whether the connection is currently open.
	// It returns false once [Close] has been called.
	IsOpen() bool
}

// StationValue is a concrete holder returned by [State.Station].
// Embedding [Station] promotes all its methods so callers can use
// v.Send(...), v.Expect(...) etc. directly without unwrapping.
type StationValue struct {
	Station
}
