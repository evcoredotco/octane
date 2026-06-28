// Package api defines the public types and interfaces that every
// OCTANE keyword library consumes. It is the contract surface
// between story authors, keyword authors, and the runtime.
//
// This package contains only type definitions, enumerations, and
// interface contracts. It has no implementation logic, no
// third-party dependencies, and no imports beyond the standard
// library.
//
// # Core types
//
// [Func] is the function signature every keyword author implements.
// [Keyword] is the registration record that binds a [Func] to its
// [Pattern], [Layer], and [OCPPVersion].
// [Args] holds the named parameter values extracted by the pattern
// matcher; its typed accessors (String, Int, Duration, …) panic on
// missing keys as a registry-bug signal.
//
// # Layer and version enumerations
//
// [Layer] and [OCPPVersion] control how the registry resolves step
// text to keyword functions. Domain-layer keywords scoped to a
// specific OCPP version take precedence over primitive-layer keywords;
// see ADR 0007 for the full resolution rules.
//
// # Runtime interfaces
//
// [State] and [Station] are the runtime's surface as seen by keyword
// functions. [State] is an interface; [Station] is the wire I/O
// interface. [State.Station] returns a [StationValue] that embeds
// [Station] and promotes all its methods. Keyword libraries can be
// unit-tested against mocks without importing the runtime or transport
// packages. For unit testing, use the test doubles in
// pkg/keywords/api/mock.
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
)

// String returns the human-readable version string for an
// OCPPVersion value (e.g., "1.6").
func (v OCPPVersion) String() string {
	switch v {
	case OCPP16:
		return "1.6"
	default:
		return "unknown"
	}
}

// State is the runtime's surface as seen by keyword functions.
// It exposes station registration and lookup, a deterministic
// clock, and structured logging.
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
// NOTE(spec-003): Spec 003 §10 proposes StashPendingCallId and
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
	Station(handle string) (StationValue, error)

	// RegisterStation binds station to handle so that subsequent
	// [State.Station] calls can retrieve it. Registering the same
	// handle twice replaces the previous entry. This method is
	// called by the open-WebSocket primitive keyword after a
	// successful [transport.Dial] to make the live connection
	// available to subsequent steps.
	RegisterStation(handle string, station Station)

	// Now returns the current time from the runtime's clock.
	// Keywords MUST use this instead of [time.Now] to preserve
	// determinism across runs (constitution principle IV).
	Now() time.Time

	// Sleep blocks until d has elapsed on the runtime's clock, or
	// until ctx is cancelled. Keywords MUST use this instead of
	// [time.Sleep] so that the runtime can inject a deterministic
	// clock and tests complete without real wall-clock delay
	// (constitution principle IV).
	//
	// Returns [context.Canceled] or [context.DeadlineExceeded] if
	// the context is done before d elapses.
	Sleep(ctx context.Context, d time.Duration) error

	// Stash stores value under key in the scenario's scratch
	// space. It is used by primitive keywords to pass data
	// between steps (e.g., the frame received by
	// [expect any frame] is stashed so a subsequent assertion
	// step can inspect it). Keys are scoped to the current
	// scenario execution and are not persisted across runs.
	//
	// A common key convention is "<keyword>:<station>",
	// e.g., "last_frame:CP01".
	Stash(key string, value any)

	// Pop retrieves and removes the value stored under key.
	// Returns the value and true if the key exists; returns
	// nil and false if it does not.
	Pop(key string) (any, bool)

	// Logf emits a structured log line scoped to the current
	// step execution. The format string and arguments follow
	// [fmt.Sprintf] conventions. Log output appears in the
	// run report under the step that produced it.
	Logf(format string, args ...any)

	// CSMSBaseURL returns the base WebSocket URL of the CSMS under test,
	// as configured via --csms-endpoint or octane.yml. Lifecycle domain
	// keywords construct per-station URLs by appending "/" + stationHandle
	// to this value (e.g. "ws://localhost:9210" → "ws://localhost:9210/CP01").
	//
	// An empty string means no CSMS endpoint has been configured; lifecycle
	// keywords must return a descriptive error in that case rather than
	// attempting a connection.
	CSMSBaseURL() string
}

// Func is the function signature every keyword author implements.
// A non-nil return marks the step as failed; the error message
// becomes the finding text in the run report. The
// [context.Context] carries cancellation and the per-step timeout.
// [State] provides runtime services (station lookup, deterministic
// clock, logging). [Args] holds the named parameter values
// extracted by the pattern matcher.
//
// Keyword bodies should not call [time.Now] directly; use
// [State.Now] instead so that the runtime can inject a
// deterministic clock (constitution principle IV).
type Func func(ctx context.Context, state State, args Args) error

// Keyword is the registration record for a single keyword. It
// binds a human-readable [Pattern] to a [Func] and declares the
// keyword's layer and optional OCPP version scope.
//
// Keywords are registered at package init() time via
// registry.Register. The registry validates that no two keywords
// share the same (Layer, OCPPVersion, Pattern) tuple; a collision
// causes a startup panic.
//
// Example registration:
//
//	registry.Register(api.Keyword{
//	    Pattern:     "station {station:string} sends BootNotification",
//	    Layer:       api.LayerDomain,
//	    OCPPVersion: api.OCPP16,
//	    Func: func(ctx context.Context, s api.State, a api.Args) error {
//	        // ... drives the wire ...
//	        return nil
//	    },
//	})
type Keyword struct {
	// Pattern is the step-matching pattern using {name:type}
	// placeholder syntax (e.g., "wait {d:duration}"). The
	// pattern matcher compares step text against this string
	// case-insensitively; whitespace in the pattern matches one
	// or more whitespace characters in the step.
	Pattern string

	// Layer identifies whether this keyword belongs to the
	// primitive (transport-level) or domain (OCPP-semantic)
	// layer. Resolution order is domain first, then primitive.
	Layer Layer

	// OCPPVersion scopes the keyword to a specific OCPP
	// version. Domain-layer keywords MUST set this to a
	// non-zero value (OCPP16).
	// Primitive-layer keywords leave this as the zero value,
	// indicating they are version-agnostic and visible to all
	// stories regardless of the declared OCPP version.
	OCPPVersion OCPPVersion

	// Func is the Go function that executes the keyword's
	// logic. It receives the runtime state and the bound
	// arguments extracted from the step text by the pattern
	// matcher. A non-nil return marks the step as failed.
	Func Func
}
