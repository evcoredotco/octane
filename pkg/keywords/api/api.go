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
package api

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
