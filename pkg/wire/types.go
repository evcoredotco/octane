// Package wire provides OCPP-J frame parsing and serialization for CALL
// (type 2), CALLRESULT (type 3), and CALLERROR (type 4) frames per
// OCPP-J -3.4.
//
// OCPP-J frames are JSON arrays whose first element is a numeric message
// type code. This package decodes the raw []any produced by encoding/json
// into typed Go values and re-encodes them to canonical JSON arrays for
// transmission.
//
// Frame parsing entry points are ParseCall, ParseResult, and ParseError
// (implemented in parse.go). Frame serialization is Encode (encode.go).
//
// The package is OCPP-version-agnostic: it handles the wire framing layer
// only. Message-level payload schemas are the responsibility of the domain
// keyword packages.
package wire

import "encoding/json"

// MessageTypeCall is the OCPP-J message type code for a CALL frame (request).
// Value 2 is mandated by OCPP-J -3.4 and must not be changed.
const MessageTypeCall = 2

// MessageTypeResult is the OCPP-J message type code for a CALLRESULT frame.
// Value 3 is mandated by OCPP-J -3.4 and must not be changed.
const MessageTypeResult = 3

// MessageTypeError is the OCPP-J message type code for a CALLERROR frame.
// Value 4 is mandated by OCPP-J -3.4 and must not be changed.
const MessageTypeError = 4

// Call represents a decoded OCPP-J CALL frame.
//
// Wire shape: [2, "<uniqueId>", "<Action>", { ...payload... }]
//
// UniqueID and Action are always non-empty strings when parsed by ParseCall.
// Payload is the raw JSON object and may be forwarded to domain decoders
// without re-serialization.
type Call struct {
	// UniqueID is the correlation identifier chosen by the station.
	// OCPP-J -3.4 requires this to be a unique string per outstanding request.
	UniqueID string
	// Action names the OCPP operation, e.g. "BootNotification" or
	// "Authorize". It is case-sensitive per the OCPP-J specification.
	Action string
	// Payload is the raw JSON object body of the CALL frame.
	// Callers unmarshal this into the concrete request type for the Action.
	Payload json.RawMessage
}

// Result represents a decoded OCPP-J CALLRESULT frame.
//
// Wire shape: [3, "<uniqueId>", { ...payload... }]
//
// UniqueID correlates the result with a prior Call. Payload is the raw
// JSON response body and may be forwarded to domain decoders without
// re-serialization.
type Result struct {
	// UniqueID matches the UniqueID from the originating Call frame.
	UniqueID string
	// Payload is the raw JSON object body of the CALLRESULT frame.
	// Callers unmarshal this into the concrete response type for the Action.
	Payload json.RawMessage
}

// Error represents a decoded OCPP-J CALLERROR frame.
//
// Wire shape:
//
//	[4, "<uniqueId>", "<errorCode>", "<errorDescription>", { ...details... }]
type Error struct {
	// UniqueID matches the UniqueID from the originating Call frame.
	UniqueID string
	// ErrorCode is the OCPP-J error code string, e.g. "NotImplemented"
	// or "InternalError". Valid codes are defined per OCPP version.
	ErrorCode string
	// ErrorDescription is a human-readable description of the error,
	// suitable for display in reports.
	ErrorDescription string
	// Details is the optional raw JSON object with additional error context.
	// An empty object ({}) or null is represented as nil.
	Details json.RawMessage
}

// FrameShapeError is returned by ParseCall, ParseResult, and ParseError when
// the inbound JSON does not match the expected OCPP-J array shape.
//
// The Reason field pinpoints the structural violation. The Raw field carries
// the first 256 bytes of the malformed frame for diagnostic logging.
type FrameShapeError struct {
	// Reason describes what was wrong with the frame, e.g.
	// "expected array of length 4, got 2" or
	// "element 0 must be a number".
	Reason string
	// Raw is the first 256 bytes of the malformed frame, for diagnostics.
	// It is never nil; it may be an empty slice if the frame was empty.
	Raw []byte
}

// Error implements the error interface.
//
// Format: "wire: malformed frame: <reason> (raw: <first 256 bytes>)".
func (e *FrameShapeError) Error() string {
	const diagCap = 256

	raw := e.Raw

	if len(raw) > diagCap {
		raw = raw[:diagCap]
	}

	return "wire: malformed frame: " + e.Reason + " (raw: " + string(raw) + ")"
}
