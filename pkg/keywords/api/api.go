// Package api defines the public types and interfaces that every
// OCTANE keyword library consumes. It is the contract surface
// between story authors, keyword authors, and the runtime.
//
// This package contains only type definitions, enumerations, and
// interface contracts. It has no implementation logic, no
// third-party dependencies, and no imports beyond the standard
// library.
//
// The two enumerations defined here — [Layer] and [OCPPVersion] —
// control how the registry resolves step text to keyword
// functions. Domain-layer keywords scoped to a specific OCPP
// version take precedence over primitive-layer keywords; see
// ADR 0007 for the full resolution rules.
//
// The [State] and [Station] interfaces are the runtime's surface
// as seen by keyword functions. Both are interfaces (not concrete
// types) so that keyword libraries can be unit-tested against
// mocks without importing the runtime or transport packages.
package api

import (
	"context"
	"time"
)

// Layer identifies the keyword library layer. OCTANE has exactly
// two layers per ADR 0007 and constitution principle XII (no
// CSMS-specific adaptation surface): primitive and domain.
//
// Resolution order is domain first, then primitive. A domain
// keyword matching the active OCPP version always wins over a
// primitive keyword with the same pattern.
type Layer int

const (
	// LayerPrimitive marks transport-level keywords that are
	// OCPP-version-agnostic (e.g., "open WebSocket to {url:string}",
	// "wait {d:duration}"). Primitive keywords serve as escape
	// hatches; domain keywords compose them.
	LayerPrimitive Layer = iota + 1

	// LayerDomain marks OCPP-semantic keywords scoped to a
	// specific OCPP version (e.g., "station {station:string}
	// sends BootNotification with reason {reason:string}").
	// Domain keywords are the primary authoring surface for
	// story files.
	LayerDomain
)

// String returns the canonical lowercase name for a Layer value.
func (l Layer) String() string {
	switch l {
	case LayerPrimitive:
		return "primitive"
	case LayerDomain:
		return "domain"
	default:
		return "unknown"
	}
}

// OCPPVersion identifies the OCPP protocol version a keyword
// targets. The enum exists solely as a registry filter; no
// version-specific message logic appears in this package.
//
// When the resolver runs against a story declaring a specific
// OCPP version, only domain keywords registered for that version
// (plus all primitive keywords) are eligible for matching.
// Domain keywords registered for a different OCPP version are
// invisible.
type OCPPVersion int

const (
	// OCPP16 represents OCPP 1.6 (JSON / OCPP-J 1.6).
	OCPP16 OCPPVersion = iota + 1

	// OCPP201 represents OCPP 2.0.1.
	OCPP201

	// OCPP21 represents OCPP 2.1.
	OCPP21
)

// String returns the human-readable version string for an
// OCPPVersion value (e.g., "1.6", "2.0.1", "2.1").
func (v OCPPVersion) String() string {
	switch v {
	case OCPP16:
		return "1.6"
	case OCPP201:
		return "2.0.1"
	case OCPP21:
		return "2.1"
	default:
		return "unknown"
	}
}

// State is the runtime's surface as seen by keyword functions.
// It exposes exactly three capabilities: station lookup, a
// deterministic clock, and structured logging.
//
// State is an interface so that keyword libraries can be
// unit-tested against a mock (see pkg/keywords/api/mock) without
// importing pkg/runner/ or any network library. The runtime's
// concrete implementation satisfies this interface and is injected
// by the runner at execution time.
//
// Keywords MUST call [State.Now] instead of [time.Now] so that
// the runtime can inject a deterministic clock and produce
// byte-identical reports across runs (constitution principle IV).
//
// The surface is intentionally minimal per ADR 0007. New methods
// require an ADR amendment and reviewer approval.
//
// TODO(spec-003): Spec 003 §10 proposes StashPendingCallId and
// PopPendingCallId methods for the request/response keyword
// pairing pattern. ADR 0007 does not include them. If adopted,
// an ADR amendment must land first. See spec 003 OQ discussion.
type State interface {
	// Station returns the [Station] handle identified by the
	// given name. The name corresponds to the station handle
	// declared in the story's Given block (e.g., "CP01").
	//
	// An error is returned if the station handle is not known
	// to the current scenario — typically because it was not
	// declared or has already been torn down.
	Station(handle string) (Station, error)

	// Now returns the current time from the runtime's clock.
	// Keywords MUST use this instead of [time.Now] to preserve
	// determinism across runs (constitution principle IV).
	Now() time.Time

	// Logf emits a structured log line scoped to the current
	// step execution. The format string and arguments follow
	// [fmt.Sprintf] conventions. Log output appears in the
	// run report under the step that produced it.
	Logf(format string, args ...any)
}

// Station is the wire-I/O surface for a single charging station
// connection. Keywords use it to send OCPP-J frames to a CSMS
// and to receive frames from the wire.
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
}
